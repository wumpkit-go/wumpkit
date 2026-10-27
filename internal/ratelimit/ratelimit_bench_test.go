package ratelimit_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/wumpkit/wumpkit/internal/ratelimit"
)

func BenchmarkNormalizeRoute(b *testing.B) {
	url := "/api/v10/channels/872712139712913438/messages/872712150509047809/reactions/PandaOhShit:863985751205085195/@me"
	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = ratelimit.NormalizeRoute(url, "DELETE")
	}
}

func BenchmarkParseHeaders(b *testing.B) {
	headers := make(http.Header)
	headers.Set("X-RateLimit-Limit", "5")
	headers.Set("X-RateLimit-Remaining", "4")
	headers.Set("X-RateLimit-Reset-After", "0.250")
	headers.Set("X-RateLimit-Bucket", "abcd1234efgh5678")
	headers.Set("X-RateLimit-Scope", "user")

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = ratelimit.ParseHeaders(headers, 200)
	}
}

func BenchmarkLimiter_WaitFastPath(b *testing.B) {
	limiter := ratelimit.New(ratelimit.WithGlobalLimit(10000000))
	ctx := context.Background()
	route := "/channels/123456789012345678/messages"
	method := "POST"

	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "10000000")
	h.Set("X-RateLimit-Remaining", "10000000")
	h.Set("X-RateLimit-Reset-After", "100.0")
	h.Set("X-RateLimit-Bucket", "benchBucket")
	limiter.Update(route, method, h, 200)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = limiter.Wait(ctx, route, method)
	}
}

func BenchmarkLimiter_WaitParallelSameBucket(b *testing.B) {
	limiter := ratelimit.New(ratelimit.WithGlobalLimit(10000000))
	ctx := context.Background()
	route := "/channels/123456789012345678/messages"
	method := "POST"

	h := make(http.Header)
	h.Set("X-RateLimit-Limit", "100000000")
	h.Set("X-RateLimit-Remaining", "100000000")
	h.Set("X-RateLimit-Reset-After", "100.0")
	h.Set("X-RateLimit-Bucket", "benchBucket")
	limiter.Update(route, method, h, 200)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = limiter.Wait(ctx, route, method)
		}
	})
}

func BenchmarkLimiter_WaitParallelMultiChannel(b *testing.B) {
	limiter := ratelimit.New(ratelimit.WithGlobalLimit(10000000))
	ctx := context.Background()

	// pre-populate 1000 channels
	for i := range 1000 {
		route := fmt.Sprintf("/channels/12345678901234%04d/messages", i)
		h := make(http.Header)
		h.Set("X-RateLimit-Limit", "100000000")
		h.Set("X-RateLimit-Remaining", "100000000")
		h.Set("X-RateLimit-Reset-After", "100.0")
		h.Set("X-RateLimit-Bucket", fmt.Sprintf("bucket_%d", i))
		limiter.Update(route, "POST", h, 200)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			route := fmt.Sprintf("/channels/12345678901234%04d/messages", i%1000)
			_ = limiter.Wait(ctx, route, "POST")
			i++
		}
	})
}
