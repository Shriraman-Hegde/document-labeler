package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store persists department catalogs under root/schemas/<dept>/catalog.json.
type Store struct {
	Root string
}

func (s Store) path(dept string) string {
	return filepath.Join(s.Root, "schemas", dept, "catalog.json")
}

func (s Store) Load(dept string) (*Catalog, error) {
	p := s.path(dept)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			c := &Catalog{Department: dept, Versions: nil}
			for _, v := range DefaultSeed() {
				v.Department = dept
				c.Versions = append(c.Versions, v)
			}
			return c, nil
		}
		return nil, err
	}
	var c Catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	c.Department = dept
	return &c, nil
}

func (s Store) Save(c *Catalog) error {
	p := s.path(c.Department)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

func (c *Catalog) Latest(t DocType) *Version {
	var best *Version
	for i := range c.Versions {
		v := &c.Versions[i]
		if v.Type != t {
			continue
		}
		if best == nil || v.Version > best.Version {
			best = v
		}
	}
	return best
}

func (c *Catalog) EnsurePrivate(t DocType, dept string, fields []FieldDef) Version {
	if cur := c.Latest(t); cur != nil {
		return *cur
	}
	v := Version{
		Type:       t,
		Version:    1,
		Status:     StatusPrivate,
		Department: dept,
		Fields:     fields,
	}
	c.Versions = append(c.Versions, v)
	return v
}

func (c *Catalog) Promote(t DocType) error {
	v := c.Latest(t)
	if v == nil {
		return fmt.Errorf("no schema for type %s", t)
	}
	v.Status = StatusPromoted
	return nil
}

func (c *Catalog) AddConflict(t DocType, conf Conflict) {
	v := c.Latest(t)
	if v == nil {
		return
	}
	v.Conflicts = append(v.Conflicts, conf)
}

func (c *Catalog) AddFieldIfMissing(t DocType, f FieldDef) (added bool, conflict *Conflict) {
	v := c.Latest(t)
	if v == nil {
		return false, nil
	}
	for _, existing := range v.Fields {
		if strings.EqualFold(existing.Key, f.Key) {
			return false, nil
		}
		if similarLabel(existing.Label, f.Label) && !strings.EqualFold(existing.Key, f.Key) {
			conf := Conflict{
				Keys:    []string{existing.Key, f.Key},
				Labels:  []string{existing.Label, f.Label},
				Message: "divergent labels kept as separate fields",
			}
			v.Fields = append(v.Fields, f)
			v.Conflicts = append(v.Conflicts, conf)
			return true, &conf
		}
	}
	v.Fields = append(v.Fields, f)
	return true, nil
}

func similarLabel(a, b string) bool {
	na := normalizeLabel(a)
	nb := normalizeLabel(b)
	if na == nb {
		return true
	}
	if strings.Contains(na, nb) || strings.Contains(nb, na) {
		return true
	}
	skip := map[string]bool{"no": true, "number": true, "num": true, "the": true}
	for _, x := range strings.Fields(na) {
		if len(x) < 2 || skip[x] {
			continue
		}
		for _, y := range strings.Fields(nb) {
			if x == y {
				return true
			}
		}
	}
	return false
}

func normalizeLabel(s string) string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer("#", "", ".", "", ":", "", "-", " ", "_", " ")
	s = repl.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}
