package wumpkit

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// small valid PNG image (1x1 transparent pixel)
const base64Png = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR4nGNgYAAAAAMAAWgmWQ0AAAAASUVORK5CYII="

func decodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic("invalid base64 in test: " + err.Error())
	}
	return b
}

func TestFile_Download(t *testing.T) {
	imageData, _ := base64.StdEncoding.DecodeString(base64Png)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imageData)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	baseName := "testfile.jpg"

	savedPath, err := DownloadFile(context.Background(), server.URL, baseName, tempDir)
	require.NoError(t, err)

	wantPath := filepath.Join(tempDir, "testfile.png")
	assert.Equal(t, wantPath, savedPath)
	assert.FileExists(t, savedPath)
}

func TestFile_NewImageFile(t *testing.T) {
	imgPath := filepath.Join(t.TempDir(), "test.png")
	err := os.WriteFile(imgPath, decodeBase64(base64Png), 0o644)
	require.NoError(t, err)

	dataURI, err := NewImageFile(imgPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(dataURI, "data:image/png;base64,"))
}

func TestFile_NewImageFile_NonImage(t *testing.T) {
	txtPath := filepath.Join(t.TempDir(), "not_image.txt")
	err := os.WriteFile(txtPath, []byte("hello world"), 0o644)
	require.NoError(t, err)

	_, err = NewImageFile(txtPath)
	assert.Error(t, err)
}
