package relate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/label"
)

// LoadValueDir loads all value JSON files under root/values (optionally one dept).
func LoadValueDir(root, dept string) ([]Doc, error) {
	base := filepath.Join(root, "values")
	if dept != "" {
		base = filepath.Join(base, dept)
	}
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil, nil
	}
	var docs []Doc
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var vf label.ValueFile
		if err := json.Unmarshal(b, &vf); err != nil {
			return err
		}
		docs = append(docs, FromValueFile(vf))
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return docs, nil
}
