package ocr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Token is one OCR word with confidence 0..1.
type Token struct {
	Text       string
	Confidence float64
	Page       int
}

// Result is OCR output for a document.
type Result struct {
	Tokens []Token
	Text   string
}

// MeanConfidence returns average word confidence, or 0 if empty.
func (r Result) MeanConfidence() float64 {
	if len(r.Tokens) == 0 {
		return 0
	}
	var sum float64
	for _, t := range r.Tokens {
		sum += t.Confidence
	}
	return sum / float64(len(r.Tokens))
}

// Weak reports the KTD4 weak-OCR gate.
func (r Result) Weak(minMean float64, minTokens int) bool {
	if len(r.Tokens) < minTokens {
		return true
	}
	return r.MeanConfidence() < minMean
}

// RunTesseract runs tesseract TSV on an image path.
func RunTesseract(imagePath string, page int) (Result, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "--psm", "6", "tsv")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("tesseract: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return parseTSV(stdout.String(), page)
}

func parseTSV(tsv string, page int) (Result, error) {
	lines := strings.Split(tsv, "\n")
	var tokens []Token
	var parts []string
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 12 {
			continue
		}
		// level text is last column; conf is index 10
		level, _ := strconv.Atoi(cols[0])
		if level != 5 { // word
			continue
		}
		text := cols[len(cols)-1]
		if text == "" {
			continue
		}
		conf, _ := strconv.ParseFloat(cols[10], 64)
		if conf < 0 {
			continue
		}
		tokens = append(tokens, Token{Text: text, Confidence: conf / 100.0, Page: page})
		parts = append(parts, text)
	}
	return Result{Tokens: tokens, Text: strings.Join(parts, " ")}, nil
}

// Available reports whether tesseract is on PATH.
func Available() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

// WriteTempImage is a helper for tests.
func WriteTempImage(dir, name string, data []byte) (string, error) {
	p := dir + "/" + name
	return p, os.WriteFile(p, data, 0o644)
}
