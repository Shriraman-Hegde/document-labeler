package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/engine"
	"github.com/Shriraman-Hegde/document-labeler/internal/llm"
	"github.com/Shriraman-Hegde/document-labeler/internal/relate"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

// Config for the HTTP server.
type Config struct {
	Addr string
	Root string
	LLM  *llm.Client
}

// Handler returns the API mux (UI lives in a separate repo).
func Handler(cfg Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		handleScan(w, r, cfg)
	})
	mux.HandleFunc("/v1/relate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "GET or POST", http.StatusMethodNotAllowed)
			return
		}
		dept := r.URL.Query().Get("dept")
		docs, err := relate.LoadValueDir(cfg.Root, dept)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, relate.Build(docs))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "document-labeler",
			"ui":      "hosted separately — set DIGITIZER_API on the UI to this origin",
			"health":  "/health",
			"scan":    "POST /v1/scan",
			"relate":  "GET /v1/relate?dept=",
		})
	})
	return withCORS(mux)
}

func handleScan(w http.ResponseWriter, r *http.Request, cfg Config) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "invalid multipart: "+err.Error(), http.StatusBadRequest)
		return
	}
	dept := strings.TrimSpace(r.FormValue("dept"))
	if dept == "" {
		http.Error(w, "dept is required", http.StatusBadRequest)
		return
	}
	modeRaw := strings.TrimSpace(r.FormValue("mode"))
	mode, err := schema.ParseMode(modeRaw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		files = r.MultipartForm.File["files"]
	}
	if len(files) == 0 {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}

	tmpDir, err := os.MkdirTemp("", "digitizer-upload-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	eng := engine.Config{Root: cfg.Root, Department: dept, ModeOverride: mode, LLM: cfg.LLM}
	var outcomes []engine.Outcome
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		name := filepath.Base(fh.Filename)
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".pdf" {
			_ = src.Close()
			http.Error(w, "unsupported type "+ext, http.StatusBadRequest)
			return
		}
		dstPath := filepath.Join(tmpDir, name)
		dst, err := os.Create(dstPath)
		if err != nil {
			_ = src.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(dst, src); err != nil {
			_ = src.Close()
			_ = dst.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = src.Close()
		_ = dst.Close()

		o, err := engine.ProcessFile(eng, dstPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %v", name, err), http.StatusUnprocessableEntity)
			return
		}
		outcomes = append(outcomes, o)
	}

	docs, err := relate.LoadValueDir(cfg.Root, dept)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rel := relate.Build(docs)
	writeJSON(w, http.StatusOK, map[string]any{
		"documents": outcomes,
		"relations": rel.Relations,
		"checks":    rel.Checks,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
