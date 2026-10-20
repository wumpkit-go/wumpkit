/************************************************************************************
 *
 * wumpkit, A Lightweight Go library for Discord API
 *
 * SPDX-License-Identifier: BSD-3-Clause
 *
 * Copyright 2025 Marouane Souiri
 *
 * Licensed under the BSD 3-Clause License.
 * See the LICENSE file for details.
 *
 ************************************************************************************/

package wumpkit

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wumpkit-go/wumpkit/internal/ratelimit"
)

/*********************************
 *   Constants & Global Errors   *
 *********************************/

const (
	apiVersion          = "v10"
	defaultBaseURL      = "https://discord.com/api/" + apiVersion
	defaultUserAgent    = "DiscordBot (wumpkit, v0.0.1)"
	headerReason        = "X-Audit-Log-Reason"
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	headerUserAgent     = "User-Agent"
	headerAccept        = "Accept"
	headerRetryAfter    = "Retry-After"
	contentTypeJSON     = "application/json"
)

// ErrMaxRetriesExceeded is a sentinel error returned when an HTTP request exceeds its maximum configured retries.
var ErrMaxRetriesExceeded = errors.New("wumpkit: max retries exceeded")

// Buffer pool for multipart request encoding to minimize heap allocations.
var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

/***********************
 *   Configuration     *
 ***********************/

// RequesterConfig holds configuration for the HTTP REST requester.
type RequesterConfig struct {
	// BaseURL is the base URL for Discord API requests, which must target the API version supported by the library ("https://discord.com/api/v10").
	// Change this to use a reverse proxy or alternative endpoint compatible with API v10.
	BaseURL string

	// UserAgent is the User-Agent header value sent on all requests.
	UserAgent string

	// MaxRetries is the maximum number of retries for failed retryable requests. Defaults to 3.
	MaxRetries int

	// HTTPClient is the custom HTTP client used for executing requests.
	HTTPClient *http.Client

	// RateLimitStore is the pluggable storage backend for rate limiting. Defaults to in-memory.
	RateLimitStore RateLimitStore

	// GlobalRateLimit is the global requests/sec limit. Defaults to 50.
	GlobalRateLimit uint
}

// DefaultRequesterConfig returns standard requester configuration defaults optimized for high throughput and low latency.
func DefaultRequesterConfig() RequesterConfig {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return RequesterConfig{
		BaseURL:         defaultBaseURL,
		UserAgent:       defaultUserAgent,
		MaxRetries:      3,
		GlobalRateLimit: 50,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           dialer.DialContext,
				MaxIdleConns:          1000,
				MaxIdleConnsPerHost:   200,
				MaxConnsPerHost:       400,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ForceAttemptHTTP2:     true,
				WriteBufferSize:       32 * 1024,
				ReadBufferSize:        32 * 1024,
			},
		},
	}
}

/***********************
 *   Requester         *
 ***********************/

// requester handles HTTP requests with high-performance connection pooling, retry backoff, and rate limiting.
type requester struct {
	config         RequesterConfig
	logger         Logger
	rateLimitStore ratelimit.Store
	rateLimiter    *ratelimit.Limiter
	authHeader     string
}

// newRequester creates a new Requester with the given configuration.
func newRequester(token string, config RequesterConfig, logger Logger) *requester {
	def := DefaultRequesterConfig()
	config.BaseURL = cmp.Or(config.BaseURL, def.BaseURL)
	config.UserAgent = cmp.Or(config.UserAgent, def.UserAgent)
	config.MaxRetries = cmp.Or(config.MaxRetries, def.MaxRetries)
	config.HTTPClient = cmp.Or(config.HTTPClient, def.HTTPClient)
	config.GlobalRateLimit = cmp.Or(config.GlobalRateLimit, def.GlobalRateLimit)

	store := cmp.Or[RateLimitStore](config.RateLimitStore, ratelimit.NewMemoryStore())

	rateLimiter := ratelimit.New(
		ratelimit.WithStore(store),
		ratelimit.WithGlobalLimit(config.GlobalRateLimit),
	)

	// precompute normalized Authorization header value once
	var authHeader string
	if token != "" {
		if strings.HasPrefix(token, "Bot ") || strings.HasPrefix(token, "Bearer ") {
			authHeader = token
		} else {
			authHeader = "Bot " + token
		}
	}

	return &requester{
		config:         config,
		logger:         logger,
		rateLimitStore: store,
		rateLimiter:    rateLimiter,
		authHeader:     authHeader,
	}
}

// Shutdown gracefully closes idle connections in the underlying HTTP transport.
func (r *requester) Shutdown() {
	if r.config.HTTPClient != nil {
		if tr, ok := r.config.HTTPClient.Transport.(interface{ CloseIdleConnections() }); ok {
			tr.CloseIdleConnections()
		}
	}
}

// isRetryableStatus checks if an HTTP response status code indicates a temporary failure worth retrying.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, // 429 rate limited
		http.StatusInternalServerError, // 500 internal Discord error
		http.StatusBadGateway,          // 502 bad Gateway / Cloudflare issue
		http.StatusServiceUnavailable,  // 503 service Unavailable / Discord restart
		http.StatusGatewayTimeout,      // 504 gateway Timeout
		524:                            // cloudflare timeout
		return true
	default:
		return false
	}
}

// computeBackoff calculates an exponential backoff duration with randomized jitter to eliminate thundering herds.
func computeBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	backoff := baseDelay * time.Duration(1<<attempt)
	if backoff > maxDelay || backoff <= 0 {
		backoff = maxDelay
	}
	half := int64(backoff / 2)
	if half <= 0 {
		half = 1
	}
	jitter := time.Duration(rand.Int64N(half + 1))
	return (backoff / 2) + jitter
}

// Request represents an HTTP request to be executed by the requester.
type Request struct {
	// Body is the raw JSON request body or other data sent to the API.
	Body []byte
	// Method is the HTTP verb (GET, POST, PUT, PATCH, DELETE, etc., using net/http Method constants).
	Method string
	// URL is the endpoint path (e.g., "/guilds/123") relative to the base URL.
	URL string
	// Reason is the value for the "X-Audit-Log-Reason" header, useful for audit logs.
	Reason string
	// ContentType overrides the default "application/json" header if set.
	ContentType string
	// NoAuth indicates whether to skip token-based authentication for this request.
	NoAuth bool
	// AuthHeader allows setting a custom Authorization header (e.g., "Bearer <user_token>").
	AuthHeader string
	// Headers allows specifying additional HTTP headers.
	Headers map[string]string
}

// MultipartRequest represents a multipart/form-data request with payload_json and file attachments.
type MultipartRequest struct {
	// Method is the HTTP verb (POST, PATCH, PUT, using net/http Method constants).
	Method string
	// URL is the endpoint path relative to the base URL.
	URL string
	// Reason is the audit log reason header value.
	Reason string
	// PayloadJSON is the JSON string/bytes sent in the payload_json field.
	PayloadJSON []byte
	// Files is the list of files to upload as form parts.
	Files []File
	// ExtraFields contains additional string form fields (e.g. for sticker creation).
	ExtraFields map[string]string
}

// DoRequest executes an HTTP request and returns the response body as an io.ReadCloser.
// The caller is responsible for closing the returned body.
func (r *requester) DoRequest(ctx context.Context, req Request) (io.ReadCloser, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	fullURL := r.config.BaseURL + req.URL

	var lastErr error
	var lastStatus int
	var lastBody []byte

	for i := range r.config.MaxRetries + 1 {
		if i > 0 {
			r.logger.Debug("retrying request",
				"method", req.Method,
				"url", fullURL,
				"attempt", i+1,
			)
		}

		// use http.NoBody when body is empty to avoid any heap reader allocations
		var bodyReader io.Reader = http.NoBody
		if len(req.Body) > 0 {
			bodyReader = bytes.NewReader(req.Body)
		}

		httpRequest, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed creating request: %w", err)
		}

		// set Authentication
		if req.AuthHeader != "" {
			httpRequest.Header.Set(headerAuthorization, req.AuthHeader)
		} else if !req.NoAuth && r.authHeader != "" {
			httpRequest.Header.Set(headerAuthorization, r.authHeader)
		}

		httpRequest.Header.Set(headerUserAgent, r.config.UserAgent)
		httpRequest.Header.Set(headerAccept, contentTypeJSON)

		// set Content-Type
		if req.ContentType != "" {
			httpRequest.Header.Set(headerContentType, req.ContentType)
		} else if len(req.Body) > 0 || req.Method == http.MethodPost || req.Method == http.MethodPatch || req.Method == http.MethodPut {
			httpRequest.Header.Set(headerContentType, contentTypeJSON)
		}

		// set custom headers
		for k, v := range req.Headers {
			httpRequest.Header.Set(k, v)
		}

		// set Audit Log Reason
		if req.Reason != "" {
			httpRequest.Header.Set(headerReason, req.Reason)
		}

		// wait for rate limit quota or global pause
		if r.rateLimiter != nil {
			if err := r.rateLimiter.Wait(ctx, req.URL, req.Method); err != nil {
				return nil, err
			}
		}

		resp, err := r.config.HTTPClient.Do(httpRequest)
		if resp != nil && r.rateLimiter != nil {
			r.rateLimiter.Update(req.URL, req.Method, resp.Header, resp.StatusCode)
		}

		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = err
			if i == r.config.MaxRetries {
				break
			}
			backoff := computeBackoff(i, 100*time.Millisecond, 3*time.Second)
			r.logger.Warn("request network error, retrying",
				"method", req.Method,
				"endpoint", req.URL,
				"error", err,
				"backoff", backoff.String(),
			)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}

		// success status (2xx)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp.Body, nil
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastStatus = resp.StatusCode
		lastBody = bodyBytes

		// non-retryable client errors (4xx except 429): fail fast immediately without retries
		if !isRetryableStatus(resp.StatusCode) {
			return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		if i == r.config.MaxRetries {
			break
		}

		// rate Limited (429): sleep according to Retry-After header with small jitter
		if resp.StatusCode == http.StatusTooManyRequests {
			var sleepDuration time.Duration = 500 * time.Millisecond
			if retryAfterHeader := resp.Header.Get(headerRetryAfter); retryAfterHeader != "" {
				if s, parseErr := strconv.ParseFloat(retryAfterHeader, 64); parseErr == nil && s > 0 {
					sleepDuration = time.Duration(s * float64(time.Second))
				}
			}
			// add 10-50ms random jitter to avoid thundering herd on rate limit expiration
			jitter := time.Duration(rand.Int64N(40)+10) * time.Millisecond
			sleepDuration += jitter

			r.logger.Warn("rate limited by Discord",
				"method", req.Method,
				"endpoint", req.URL,
				"retry_after", sleepDuration.String(),
			)
			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}

		// server Error (5xx / 524): exponential backoff with jitter
		backoff := computeBackoff(i, 100*time.Millisecond, 3*time.Second)
		r.logger.Warn("request failed with retryable status",
			"method", req.Method,
			"endpoint", req.URL,
			"status", resp.StatusCode,
			"backoff", backoff.String(),
		)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w for %s %s: %w", ErrMaxRetriesExceeded, req.Method, req.URL, lastErr)
	}
	if lastStatus != 0 {
		return nil, fmt.Errorf("%w for %s %s (status %d): %s", ErrMaxRetriesExceeded, req.Method, req.URL, lastStatus, string(lastBody))
	}
	return nil, fmt.Errorf("%w for %s %s", ErrMaxRetriesExceeded, req.Method, req.URL)
}

// DoMultipartRequest executes a multipart/form-data HTTP request with JSON payload and file attachments.
func (r *requester) DoMultipartRequest(ctx context.Context, req MultipartRequest) (io.ReadCloser, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	w := multipart.NewWriter(buf)

	// write extra string fields if any
	for k, v := range req.ExtraFields {
		if err := w.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("failed to write multipart field %s: %w", k, err)
		}
	}

	// write payload_json if provided
	if len(req.PayloadJSON) > 0 {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="payload_json"`)
		h.Set(headerContentType, contentTypeJSON)
		part, err := w.CreatePart(h)
		if err != nil {
			return nil, fmt.Errorf("failed to create payload_json part: %w", err)
		}
		if _, err := part.Write(req.PayloadJSON); err != nil {
			return nil, fmt.Errorf("failed to write payload_json: %w", err)
		}
	}

	// write files
	for i, file := range req.Files {
		fieldName := fmt.Sprintf("files[%d]", i)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, file.Name))
		contentType := file.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		h.Set(headerContentType, contentType)

		part, err := w.CreatePart(h)
		if err != nil {
			return nil, fmt.Errorf("failed to create file part %s: %w", file.Name, err)
		}
		if file.Reader != nil {
			if _, err := io.Copy(part, file.Reader); err != nil {
				return nil, fmt.Errorf("failed to write file content %s: %w", file.Name, err)
			}
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// copy out bytes for safe concurrent execution
	bodyBytes := make([]byte, buf.Len())
	copy(bodyBytes, buf.Bytes())

	return r.DoRequest(ctx, Request{
		Method:      req.Method,
		URL:         req.URL,
		Body:        bodyBytes,
		ContentType: w.FormDataContentType(),
		Reason:      req.Reason,
	})
}

// requestJSON sends a request, unmarshals the JSON response into T, and returns (T, error).
func requestJSON[T any](ctx context.Context, r *requester, req Request) (T, error) {
	var zero T
	body, err := r.DoRequest(ctx, req)
	if err != nil {
		return zero, err
	}
	defer body.Close()

	var result T
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return zero, fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return result, nil
}

// requestMultipartJSON sends a multipart request, unmarshals the JSON response into T, and returns (T, error).
func requestMultipartJSON[T any](ctx context.Context, r *requester, req MultipartRequest) (T, error) {
	var zero T
	body, err := r.DoMultipartRequest(ctx, req)
	if err != nil {
		return zero, err
	}
	defer body.Close()

	var result T
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return zero, fmt.Errorf("failed to decode multipart JSON response: %w", err)
	}
	return result, nil
}

// requestVoid sends a request where response body is discarded, returning error on failure.
func requestVoid(ctx context.Context, r *requester, req Request) error {
	body, err := r.DoRequest(ctx, req)
	if err != nil {
		return err
	}
	body.Close()
	return nil
}
