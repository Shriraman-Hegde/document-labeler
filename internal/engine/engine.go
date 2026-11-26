package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/ingest"
	"github.com/Shriraman-Hegde/document-labeler/internal/label"
	"github.com/Shriraman-Hegde/document-labeler/internal/llm"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

// Config drives a scan run.
type Config struct {
	Root       string
	Department string
	ModeOverride schema.Mode
	LLM        *llm.Client
}

// Outcome is one file result.
type Outcome struct {
	Source     string
	ValuePath  string
	Type       schema.DocType
	Mode       schema.Mode
	Rescan     bool
	Message    string
}

func envFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// ProcessFile runs the digitize pipeline for one path.
func ProcessFile(cfg Config, path string) (Outcome, error) {
	doc, err := ingest.Load(path)
	if err != nil {
		return Outcome{}, err
	}
	defer doc.Close()

	ocrRes, err := doc.OCRAll()
	if err != nil {
		return Outcome{}, err
	}

	minMean := envFloat("DIGITIZER_WEAK_OCR_MEAN", 0.55)
	minTok := envInt("DIGITIZER_WEAK_OCR_TOKENS", 12)
	weak := ocrRes.Weak(minMean, minTok)

	assist, err := cfg.LLM.Assist(llm.AssistRequest{
		OCRText:  ocrRes.Text,
		Filename: filepath.Base(path),
		WeakOCR:  weak,
	})
	if err != nil {
		return Outcome{}, err
	}
	text := ocrRes.Text
	if weak && assist.Improved != "" {
		text = assist.Improved
	}

	store := schema.Store{Root: cfg.Root}
	cat, err := store.Load(cfg.Department)
	if err != nil {
		return Outcome{}, err
	}

	diff := label.Differentiate(label.DiffInput{
		OCRText:  text,
		Filename: path,
		Assist:   assist,
		Catalog:  cat,
	})

	rescanConf := envFloat("DIGITIZER_RESCAN_CONF", 0.4)
	headingCount := countHeadingish(text)
	if diff.Type == "unknown" || diff.Confidence < rescanConf || headingCount < 4 {
		vw := label.Writer{Root: cfg.Root}
		vf := label.ValueFile{
			Source:       path,
			Department:   cfg.Department,
			Type:         "unknown",
			RescanNeeded: true,
			OCRText:      text,
			Mode:         schema.ModeFullyLearned,
		}
		vp, err := vw.Write(cfg.Department, vf)
		if err != nil {
			return Outcome{}, err
		}
		return Outcome{
			Source:    path,
			ValuePath: vp,
			Type:      "unknown",
			Rescan:    true,
			Message:   "rescan needed",
		}, nil
	}

	mode := label.SelectMode(cfg.ModeOverride, diff.Type, cat, diff.Confidence)
	fields := fieldsFromText(text, diff.Type, cat)
	ver := cat.EnsurePrivate(diff.Type, cfg.Department, seedFields(diff.Type))
	var conflicts []schema.Conflict
	for _, f := range fields {
		_, conf := cat.AddFieldIfMissing(diff.Type, schema.FieldDef{Key: f.Key, Label: f.Label})
		if conf != nil {
			conflicts = append(conflicts, *conf)
		}
	}
	if err := store.Save(cat); err != nil {
		return Outcome{}, err
	}
	ver = *cat.Latest(diff.Type)

	soft := label.ExtractSoftKeys(text)
	if !label.AllowCrossDept(&ver) {
		// still record on value file; cross-dept index is out of band for v1
	}

	vw := label.Writer{Root: cfg.Root}
	vf := label.ValueFile{
		Source:       path,
		Department:   cfg.Department,
		Type:         diff.Type,
		SchemaVer:    ver.Version,
		SchemaStatus: ver.Status,
		Mode:         mode,
		Fields:       fields,
		SoftLinks:    soft,
		Conflicts:    conflicts,
		OCRText:      text,
	}
	vp, err := vw.Write(cfg.Department, vf)
	if err != nil {
		return Outcome{}, err
	}

	// move/copy to labelled on success
	labelledDir := filepath.Join(cfg.Root, "labelled", cfg.Department)
	_ = os.MkdirAll(labelledDir, 0o755)
	dest := filepath.Join(labelledDir, filepath.Base(path))
	_ = copyFile(path, dest)

	return Outcome{
		Source:    path,
		ValuePath: vp,
		Type:      diff.Type,
		Mode:      mode,
		Message:   fmt.Sprintf("labelled as %s mode=%s", diff.Type, mode),
	}, nil
}

func seedFields(t schema.DocType) []schema.FieldDef {
	for _, s := range schema.DefaultSeed() {
		if s.Type == t {
			return s.Fields
		}
	}
	return nil
}

func fieldsFromText(text string, t schema.DocType, cat *schema.Catalog) []label.FieldValue {
	var out []label.FieldValue
	base := seedFields(t)
	if v := cat.Latest(t); v != nil {
		base = v.Fields
	}
	soft := label.ExtractSoftKeys(text)
	byName := map[string]string{}
	for _, s := range soft {
		byName[s.Name] = s.Value
	}
	for _, f := range base {
		fv := label.FieldValue{Key: f.Key, Label: f.Label, Confidence: 0.7}
		if val, ok := byName[f.Key]; ok {
			fv.Text = val
			fv.Confidence = 0.9
		}
		out = append(out, fv)
	}
	// Fill quantity from OCR when the field is still empty.
	for i := range out {
		if out[i].Key == "quantity" && out[i].Text == "" {
			if q := quantityFromOCR(text); q != "" {
				out[i].Text = q
				out[i].Confidence = 0.85
			}
		}
	}
	return out
}

var qtyOCR = regexp.MustCompile(`(?i)\bquantity[\s#:.-]*(\d+(?:\.\d+)?)`)

func quantityFromOCR(text string) string {
	m := qtyOCR.FindStringSubmatch(text)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func countHeadingish(text string) int {
	low := strings.ToLower(text)
	n := 0
	for _, w := range []string{"invoice", "po", "order", "packing", "receipt", "inspection", "supplier", "quantity", "part", "lot", "delivery"} {
		if strings.Contains(low, w) {
			n++
		}
	}
	return n
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}
