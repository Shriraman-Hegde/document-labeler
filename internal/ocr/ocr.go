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

// RunTesseract preprocesses the image then runs tesseract TSV.
func RunTesseract(imagePath string, page int) (Result, error) {
	clean, cleanup, err := PrepareImage(imagePath)
	if err != nil {
		// Non-image paths (or decode failures): fall back to raw file for fixtures/PDFs.
		return runTesseractFile(imagePath, page)
	}
	defer cleanup()
	return runTesseractFile(clean, page)
}

func runTesseractFile(imagePath string, page int) (Result, error) {
	psm := tesseractPSM()
	cmd := exec.Command("tesseract", imagePath, "stdout", "--psm", psm, "tsv")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("tesseract: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return parseTSV(stdout.String(), page, minTokenConf())
}

func tesseractPSM() string {
	if v := strings.TrimSpace(os.Getenv("DIGITIZER_TESSERACT_PSM")); v != "" {
		return v
	}
	// 4 = assume a single column of text of variable sizes — better for phone photos of statements/forms.
	return "4"
}

func minTokenConf() float64 {
	v := os.Getenv("DIGITIZER_OCR_MIN_CONF")
	if v == "" {
		return 0.40
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0.40
	}
	return f
}

func parseTSV(tsv string, page int, minConf float64) (Result, error) {
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
		c := conf / 100.0
		if c < minConf {
			continue
		}
		if looksLikeNoise(text) {
			continue
		}
		tokens = append(tokens, Token{Text: text, Confidence: c, Page: page})
		parts = append(parts, text)
	}
	return Result{Tokens: tokens, Text: strings.Join(parts, " ")}, nil
}

func looksLikeNoise(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	if len(s) == 1 {
		switch s {
		case "|", "/", "\\", "_", "-", ".", ",", "'", "`", "~", "*", "•", "—", "–", "«", "»":
			return true
		}
	}
	letters := 0
	symbols := 0
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			letters++
		default:
			symbols++
		}
	}
	if letters == 0 {
		return true
	}
	if len(s) >= 4 && symbols*2 >= letters {
		return true
	}
	return false
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
