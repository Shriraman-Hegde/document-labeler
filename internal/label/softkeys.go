package label

import (
	"regexp"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

var keyPatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"poNumber", regexp.MustCompile(`(?i)\bPO(?:\s*Number)?[\s#:.-]*([A-Z0-9-]{3,})`)},
	{"lot", regexp.MustCompile(`(?i)\blot[\s#:.-]*([A-Z0-9-]{2,})`)},
	{"partNumber", regexp.MustCompile(`(?i)\b(?:part(?:\s*Number)?|pn)[\s#:.-]*([A-Z0-9-]{2,})`)},
	{"deliveryNumber", regexp.MustCompile(`(?i)\b(?:delivery|dn)[\s#:.-]*([A-Z0-9-]{2,})`)},
	{"drawingNumber", regexp.MustCompile(`(?i)\b(?:drawing|dwg)[\s#:.-]*([A-Z0-9-]{2,})`)},
}

// ExtractSoftKeys finds shared-key hints in OCR text.
func ExtractSoftKeys(text string) []SoftLink {
	var out []SoftLink
	seen := map[string]bool{}
	for _, p := range keyPatterns {
		m := p.re.FindStringSubmatch(text)
		if len(m) < 2 {
			continue
		}
		val := normalizeKey(m[1])
		id := p.name + "=" + val
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, SoftLink{Name: p.name, Value: val})
	}
	return out
}

func normalizeKey(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// AllowCrossDept reports whether soft keys may cross departments (promoted only).
func AllowCrossDept(v *schema.Version) bool {
	return v != nil && v.Status == schema.StatusPromoted
}
