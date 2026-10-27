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
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// File represents a file to be uploaded to Discord in multipart requests.
type File struct {
	// Name is the name of the file (including extension, e.g. "image.png").
	Name string

	// Description is an optional description (alt text) for the attachment (max 1024 characters).
	Description string

	// Reader contains the data of the file.
	Reader io.Reader

	// ContentType is the MIME type of the file. If empty, it will be sniffed or defaulted.
	ContentType string
}

// NewFile creates a new File from an io.Reader.
func NewFile(name string, reader io.Reader, description ...string) File {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	return File{
		Name:        name,
		Reader:      reader,
		Description: desc,
	}
}

// NewFileFromDisk opens a file on disk and returns a File ready for upload.
func NewFileFromDisk(path string, description ...string) (File, error) {
	file, err := os.Open(path)
	if err != nil {
		return File{}, err
	}
	name := filepath.Base(path)
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	return File{
		Name:        name,
		Reader:      file,
		Description: desc,
	}, nil
}

// DownloadFile downloads a file from the given URL and saves it in the specified directory.
// The filename is derived from baseName and the Content-Type returned by the server.
// Returns the full path of the saved file.
func DownloadFile(ctx context.Context, url, baseName, dir string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if url == "" {
		return "", errors.New("URL is empty")
	}

	reqHead, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create head request: %w", err)
	}
	resp, err := http.DefaultClient.Do(reqHead)
	if err != nil {
		return "", fmt.Errorf("failed to fetch headers: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch headers: status %d", resp.StatusCode)
	}

	contentType := cmp.Or(resp.Header.Get("Content-Type"), "application/octet-stream")

	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		exts = []string{filepath.Ext(baseName)}
	}
	ext := exts[0]

	name := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	finalName := name + ext
	fullPath := filepath.Join(dir, finalName)

	reqGet, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create get request: %w", err)
	}
	respGet, err := http.DefaultClient.Do(reqGet)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file: %w", err)
	}
	defer func() { _ = respGet.Body.Close() }()

	if respGet.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch file: status %d", respGet.StatusCode)
	}

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	_, err = io.Copy(outFile, respGet.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// Base64Image represents a base64-encoded image data URI string.
type Base64Image = string

// NewImageFile reads an image file and returns its base64 data URI string.
//
// Example output: "data:image/png;base64,<base64-encoded-bytes>"
func NewImageFile(path string) (Base64Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("not an image file: detected MIME type %s", mimeType)
	}

	if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("invalid image data: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}
