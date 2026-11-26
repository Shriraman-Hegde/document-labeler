package schema

import (
	"path/filepath"
	"testing"
)

func TestSeedAndPromote(t *testing.T) {
	dir := t.TempDir()
	store := Store{Root: dir}
	cat, err := store.Load("stores")
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Versions) < 5 {
		t.Fatalf("expected seed types, got %d", len(cat.Versions))
	}
	if err := cat.Promote("invoice"); err != nil {
		t.Fatal(err)
	}
	if cat.Latest("invoice").Status != StatusPromoted {
		t.Fatal("expected promoted")
	}
	if err := store.Save(cat); err != nil {
		t.Fatal(err)
	}
	cat2, err := store.Load("stores")
	if err != nil {
		t.Fatal(err)
	}
	if cat2.Latest("invoice").Status != StatusPromoted {
		t.Fatal("persist promote")
	}
	_ = filepath.Join(dir, "schemas", "stores", "catalog.json")
}

func TestDivergentLabelsConflict(t *testing.T) {
	cat := &Catalog{Department: "audit"}
	cat.EnsurePrivate("invoice", "audit", []FieldDef{{Key: "poNumber", Label: "PO No."}})
	_, conf := cat.AddFieldIfMissing("invoice", FieldDef{Key: "poNumberAlt", Label: "PO Number"})
	if conf == nil {
		t.Fatal("expected conflict flag")
	}
	if len(cat.Latest("invoice").Fields) != 2 {
		t.Fatal("both fields kept")
	}
}
