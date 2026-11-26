package schema

import "fmt"

// Mode is a schema learning strategy.
type Mode string

const (
	ModeFullyLearned     Mode = "fully_learned"
	ModeSeededExtensible Mode = "seeded_extensible"
	ModeKnownDynamic     Mode = "known_dynamic"
)

func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeFullyLearned, ModeSeededExtensible, ModeKnownDynamic, "":
		return Mode(s), nil
	default:
		return "", fmt.Errorf("unknown schema mode %q", s)
	}
}

// Status is catalog lifecycle for a type version.
type Status string

const (
	StatusPrivate  Status = "private"
	StatusPromoted Status = "promoted"
)

// DocType is a branded document type id.
type DocType string

// FieldDef is a heading/key pair in a schema version.
type FieldDef struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// Conflict notes divergent labels kept together.
type Conflict struct {
	Keys    []string `json:"keys"`
	Labels  []string `json:"labels"`
	Message string   `json:"message"`
}

// Version is one schema version for a document type.
type Version struct {
	Type       DocType    `json:"type"`
	Version    int        `json:"version"`
	Status     Status     `json:"status"`
	Department string     `json:"department"`
	Fields     []FieldDef `json:"fields"`
	Conflicts  []Conflict `json:"conflicts,omitempty"`
	Overlays   []FieldDef `json:"overlays,omitempty"` // department overlays on promoted types
}

// Catalog is a department's schema store.
type Catalog struct {
	Department string    `json:"department"`
	Versions   []Version `json:"versions"`
}
