package ingest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/ocr"
)

// Document is a loaded scan ready for OCR.
type Document struct {
	Path       string
	Ext        string
	ImagePaths []string
	Cleanup    []string
}

// Load prepares images for OCR from PDF/PNG/JPEG.
func Load(path string) (*Document, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg":
		return &Document{Path: path, Ext: ext, ImagePaths: []string{path}}, nil
	case ".pdf":
		return loadPDF(path)
	default:
		return nil, fmt.Errorf("unsupported format %s (want pdf/png/jpeg)", ext)
	}
}

func loadPDF(path string) (*Document, error) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm required for PDF: %w", err)
	}
	dir, err := os.MkdirTemp("", "digitizer-pdf-*")
	if err != nil {
		return nil, err
	}
	prefix := filepath.Join(dir, "page")
	cmd := exec.Command("pdftoppm", "-png", path, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("pdftoppm: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	entries, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(entries) == 0 {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("pdftoppm produced no pages")
	}
	return &Document{Path: path, Ext: ".pdf", ImagePaths: entries, Cleanup: []string{dir}}, nil
}

// Close removes temp artifacts.
func (d *Document) Close() {
	for _, p := range d.Cleanup {
		_ = os.RemoveAll(p)
	}
}

// OCRAll runs Tesseract on every page image.
func (d *Document) OCRAll() (ocr.Result, error) {
	var all ocr.Result
	var texts []string
	for i, img := range d.ImagePaths {
		r, err := ocr.RunTesseract(img, i+1)
		if err != nil {
			return ocr.Result{}, err
		}
		all.Tokens = append(all.Tokens, r.Tokens...)
		texts = append(texts, r.Text)
	}
	all.Text = strings.Join(texts, "\n")
	return all, nil
}
