package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigFromEnvRequiresURL(t *testing.T) {
	t.Setenv("DIGITIZER_LLAMA_URL", "")
	t.Setenv("DIGITIZER_ALLOW_STUB", "")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("expected error without URL or stub")
	}
}

func TestConfigFromEnvStub(t *testing.T) {
	t.Setenv("DIGITIZER_LLAMA_URL", "")
	t.Setenv("DIGITIZER_ALLOW_STUB", "1")
	c, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Ready(); err != nil {
		t.Fatal(err)
	}
	out, err := c.Assist(AssistRequest{OCRText: "Invoice PO-1", Filename: "a.png"})
	if err != nil {
		t.Fatal(err)
	}
	if out.DocType != "invoice" {
		t.Fatalf("got %s", out.DocType)
	}
}

func TestAssistAgainstFakeServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health", "/v1/models":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/v1/chat/completions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{
						"content": `{"docType":"packing_slip","confidence":0.91,"headings":["packing slip"],"improvedText":""}`,
					}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "qwen", HTTPClient: srv.Client()}
	if err := c.Ready(); err != nil {
		t.Fatal(err)
	}
	out, err := c.Assist(AssistRequest{OCRText: "Packing Slip", Filename: "x.png"})
	if err != nil {
		t.Fatal(err)
	}
	if out.DocType != "packing_slip" || out.Confidence < 0.9 {
		t.Fatalf("%+v", out)
	}
}

func TestReadyFailsClosed(t *testing.T) {
	c := &Client{BaseURL: "http://127.0.0.1:1", Model: "qwen", HTTPClient: http.DefaultClient, AllowStub: false}
	if err := c.Ready(); err == nil {
		t.Fatal("expected unreachable error")
	}
}
