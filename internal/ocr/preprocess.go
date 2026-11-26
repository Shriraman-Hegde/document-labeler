package ocr

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
)

const defaultMaxEdge = 2400

// PrepareImage normalizes a phone/camera scan for Tesseract:
// EXIF orientation, grayscale, contrast, optional downscale, light sharpen.
// Caller must run cleanup when finished (removes the temp PNG).
func PrepareImage(srcPath string) (cleanPath string, cleanup func(), err error) {
	cleanup = func() {}
	img, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return "", cleanup, fmt.Errorf("open image: %w", err)
	}

	img = imaging.Grayscale(img)
	img = imaging.AdjustContrast(img, contrastAmount())
	img = imaging.Sharpen(img, 0.6)
	img = limitEdge(img, maxEdge())

	dir, err := os.MkdirTemp("", "digitizer-ocr-*")
	if err != nil {
		return "", cleanup, err
	}
	cleanup = func() { _ = os.RemoveAll(dir) }

	cleanPath = filepath.Join(dir, "page.png")
	if err := imaging.Save(img, cleanPath); err != nil {
		cleanup()
		cleanup = func() {}
		return "", cleanup, fmt.Errorf("save preprocessed: %w", err)
	}
	return cleanPath, cleanup, nil
}

func limitEdge(img image.Image, maxEdge int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	long := w
	if h > long {
		long = h
	}
	if maxEdge <= 0 || long <= maxEdge {
		return img
	}
	scale := float64(maxEdge) / float64(long)
	return imaging.Resize(img, int(float64(w)*scale), int(float64(h)*scale), imaging.Lanczos)
}

func maxEdge() int {
	v := os.Getenv("DIGITIZER_OCR_MAX_EDGE")
	if v == "" {
		return defaultMaxEdge
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 800 {
		return defaultMaxEdge
	}
	return n
}

func contrastAmount() float64 {
	v := os.Getenv("DIGITIZER_OCR_CONTRAST")
	if v == "" {
		return 18
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 18
	}
	return f
}
