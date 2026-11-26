package ocr

import (
	"path/filepath"
	"runtime"
	"testing"
)

func fixture(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "fixtures", name)
}

func TestRunTesseractInvoice(t *testing.T) {
	if !Available() {
		t.Skip("tesseract not installed")
	}
	r, err := RunTesseract(fixture("invoice.png"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Tokens) == 0 {
		t.Fatal("expected OCR tokens")
	}
	if r.Weak(0.55, 12) {
		t.Logf("invoice marked weak (mean=%.2f tokens=%d) — still ok for fixture variance", r.MeanConfidence(), len(r.Tokens))
	}
	low := r.Text
	if low == "" {
		t.Fatal("empty text")
	}
}

func TestRunTesseractPackingSlip(t *testing.T) {
	if !Available() {
		t.Skip("tesseract not installed")
	}
	r, err := RunTesseract(fixture("packing_slip.png"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Tokens) == 0 {
		t.Fatal("expected tokens")
	}
}

func TestWeakGate(t *testing.T) {
	r := Result{Tokens: []Token{{Text: "a", Confidence: 0.2}}}
	if !r.Weak(0.55, 12) {
		t.Fatal("expected weak")
	}
}
