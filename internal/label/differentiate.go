package label

import (
	"path/filepath"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/llm"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

// DiffInput is differentiation signals.
type DiffInput struct {
	OCRText    string
	Filename   string
	Assist     llm.AssistResponse
	Catalog    *schema.Catalog
}

// DiffResult is type judgment (filename never sole authority).
type DiffResult struct {
	Type       schema.DocType
	Confidence float64
	Headings   []string
	Reason     string
}

// Differentiate combines headings + AI; filename is a weak hint only.
func Differentiate(in DiffInput) DiffResult {
	text := strings.ToLower(in.OCRText)
	aiType := schema.DocType(strings.TrimSpace(in.Assist.DocType))
	aiConf := in.Assist.Confidence
	headings := in.Assist.Headings

	fromPage := matchHeadings(text)
	fileHint := filenameHint(in.Filename)

	// Headings and AI together: agree → strong; disagree → prefer page headings.
	if fromPage != "" && aiType != "" && aiType != "unknown" {
		if fromPage == aiType {
			conf := max(aiConf, 0.9)
			if fileHint == fromPage {
				conf = min(0.95, conf+0.03)
			}
			return DiffResult{Type: fromPage, Confidence: conf, Headings: headings, Reason: "headings+ai"}
		}
		return DiffResult{Type: fromPage, Confidence: 0.8, Headings: headings, Reason: "headings_over_ai"}
	}
	if fromPage != "" {
		return DiffResult{Type: fromPage, Confidence: 0.75, Headings: headings, Reason: "headings"}
	}
	if aiType != "" && aiType != "unknown" {
		conf := aiConf
		if fileHint == aiType {
			conf = min(0.95, conf+0.05) // filename only boosts
		}
		return DiffResult{Type: aiType, Confidence: conf, Headings: headings, Reason: "ai_primary"}
	}
	// Filename alone is NOT enough
	if fileHint != "" {
		return DiffResult{Type: "unknown", Confidence: 0.2, Headings: headings, Reason: "filename_hint_only"}
	}
	return DiffResult{Type: "unknown", Confidence: 0.1, Headings: headings, Reason: "none"}
}

func matchHeadings(text string) schema.DocType {
	checks := []struct {
		t    schema.DocType
		need []string
	}{
		{"invoice", []string{"invoice"}},
		{"purchase_order", []string{"purchase order", "po number"}},
		{"packing_slip", []string{"packing slip", "packing"}},
		{"goods_receipt", []string{"goods receipt", "grn"}},
		{"inspection_report", []string{"inspection"}},
	}
	for _, c := range checks {
		for _, n := range c.need {
			if strings.Contains(text, n) {
				return c.t
			}
		}
	}
	return ""
}

func filenameHint(name string) schema.DocType {
	base := strings.ToLower(filepath.Base(name))
	switch {
	case containsToken(base, "invoice"):
		return "invoice"
	case containsToken(base, "packing"):
		return "packing_slip"
	case containsToken(base, "receipt") || containsToken(base, "grn"):
		return "goods_receipt"
	case containsToken(base, "inspection") || containsToken(base, "coa"):
		return "inspection_report"
	case containsToken(base, "po") || containsToken(base, "order"):
		// whole-token-ish: avoid "reorder"
		if strings.Contains(base, "reorder") {
			return ""
		}
		return "purchase_order"
	}
	return ""
}

func containsToken(s, tok string) bool {
	return strings.Contains(s, tok)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// SelectMode implements KTD8 auto-select.
func SelectMode(override schema.Mode, t schema.DocType, cat *schema.Catalog, conf float64) schema.Mode {
	if override != "" {
		return override
	}
	if t == "" || t == "unknown" || conf < 0.4 {
		return schema.ModeFullyLearned
	}
	if cat != nil {
		if v := cat.Latest(t); v != nil {
			return schema.ModeKnownDynamic
		}
	}
	for _, s := range schema.DefaultSeed() {
		if s.Type == t {
			return schema.ModeSeededExtensible
		}
	}
	return schema.ModeFullyLearned
}
