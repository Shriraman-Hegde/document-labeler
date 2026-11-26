package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Shriraman-Hegde/document-labeler/internal/llm"
)

func TestHealth(t *testing.T) {
	t.Setenv("DIGITIZER_ALLOW_STUB", "1")
	client, err := llm.ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	h := Handler(Config{Root: t.TempDir(), LLM: client})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestScanUpload(t *testing.T) {
	t.Setenv("DIGITIZER_ALLOW_STUB", "1")
	client, err := llm.ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	h := Handler(Config{Root: root, LLM: client})

	fixture := filepath.Join("..", "..", "testdata", "fixtures", "invoice.png")
	b, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("dept", "stores")
	part, err := w.CreateFormFile("file", "invoice.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(b); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/scan", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	body, _ := io.ReadAll(rr.Body)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, body)
	}
}

func TestAPIIndex(t *testing.T) {
	t.Setenv("DIGITIZER_ALLOW_STUB", "1")
	client, _ := llm.ConfigFromEnv()
	h := Handler(Config{Root: t.TempDir(), LLM: client})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("document-labeler")) {
		t.Fatal("expected API index JSON")
	}
}

func TestCORSPreflight(t *testing.T) {
	t.Setenv("DIGITIZER_ALLOW_STUB", "1")
	client, _ := llm.ConfigFromEnv()
	h := Handler(Config{Root: t.TempDir(), LLM: client})
	req := httptest.NewRequest(http.MethodOptions, "/v1/scan", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("missing CORS header")
	}
}
