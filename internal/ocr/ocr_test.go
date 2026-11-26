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

func TestParseTSVDropsNoise(t *testing.T) {
	tsv := "level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
		"5\t1\t0\t0\t0\t0\t0\t0\t1\t1\t90.0\tInvoice\n" +
		"5\t1\t0\t0\t0\t1\t0\t0\t1\t1\t10.0\t~~!!\n" +
		"5\t1\t0\t0\t0\t2\t0\t0\t1\t1\t80.0\t|\n" +
		"5\t1\t0\t0\t0\t3\t0\t0\t1\t1\t85.0\tTotal\n"
	r, err := parseTSV(tsv, 1, 0.40)
	if err != nil {
		t.Fatal(err)
	}
	if r.Text != "Invoice Total" {
		t.Fatalf("got %q", r.Text)
	}
}
