package label

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

// FieldValue is extracted text with confidence.
type FieldValue struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// SoftLink is a shared-key hint.
type SoftLink struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ValueFile is the per-document value artifact.
type ValueFile struct {
	Source       string             `json:"source"`
	Department   string             `json:"department"`
	Type         schema.DocType     `json:"type"`
	SchemaVer    int                `json:"schemaVersion"`
	SchemaStatus schema.Status      `json:"schemaStatus"`
	Mode         schema.Mode        `json:"mode"`
	RescanNeeded bool               `json:"rescanNeeded,omitempty"`
	Fields       []FieldValue       `json:"fields"`
	SoftLinks    []SoftLink         `json:"softLinks,omitempty"`
	Conflicts    []schema.Conflict  `json:"conflicts,omitempty"`
	OCRText      string             `json:"ocrText,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
}

// Writer writes value JSON under root/values/<dept>/.
type Writer struct {
	Root string
}

func (w Writer) Write(dept string, vf ValueFile) (string, error) {
	sub := dept
	if vf.Type == "unknown" || vf.RescanNeeded {
		sub = filepath.Join(dept, "unknown")
	}
	dir := filepath.Join(w.Root, "values", sub)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	base := filepath.Base(vf.Source)
	name := time.Now().UTC().Format("20060102T150405") + "-" + base + ".json"
	path := filepath.Join(dir, name)
	vf.CreatedAt = time.Now().UTC()
	b, err := json.MarshalIndent(vf, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, b, 0o644)
}
