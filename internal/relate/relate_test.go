package relate

import (
	"testing"

	"github.com/Shriraman-Hegde/document-labeler/internal/label"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

func TestInvoiceForPO(t *testing.T) {
	docs := []Doc{
		{
			File: "po.png", Type: "purchase_order", Department: "stores",
			SoftLinks: []label.SoftLink{{Name: "poNumber", Value: "PO-55501"}, {Name: "partNumber", Value: "PN-100"}},
			Fields:    []label.FieldValue{{Key: "quantity", Text: "10"}},
		},
		{
			File: "inv.png", Type: "invoice", Department: "stores",
			SoftLinks: []label.SoftLink{{Name: "poNumber", Value: "PO-55501"}, {Name: "partNumber", Value: "PN-100"}},
			Fields:    []label.FieldValue{{Key: "quantity", Text: "10"}},
		},
	}
	r := Build(docs)
	if len(r.Relations) == 0 {
		t.Fatal("expected relations")
	}
	found := false
	for _, rel := range r.Relations {
		if rel.Kind == "invoice_for_po" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing invoice_for_po: %+v", r.Relations)
	}
	if len(r.Checks) != 1 || r.Checks[0].Status != "match" {
		t.Fatalf("checks=%+v", r.Checks)
	}
}

func TestQuantityMismatch(t *testing.T) {
	docs := []Doc{
		{File: "po.png", Type: "purchase_order", Department: "stores",
			SoftLinks: []label.SoftLink{{Name: "partNumber", Value: "PN-1"}},
			Fields:    []label.FieldValue{{Key: "quantity", Text: "10"}}},
		{File: "ship.png", Type: "packing_slip", Department: "stores",
			SoftLinks: []label.SoftLink{{Name: "partNumber", Value: "PN-1"}},
			Fields:    []label.FieldValue{{Key: "quantity", Text: "8"}}},
	}
	r := Build(docs)
	if len(r.Checks) != 1 || r.Checks[0].Status != "mismatch" {
		t.Fatalf("%+v", r.Checks)
	}
}

func TestCrossDeptRequiresPromote(t *testing.T) {
	docs := []Doc{
		{File: "a.png", Type: "invoice", Department: "stores", Promoted: false,
			SoftLinks: []label.SoftLink{{Name: "poNumber", Value: "PO-1"}}},
		{File: "b.png", Type: "purchase_order", Department: "audit", Promoted: false,
			SoftLinks: []label.SoftLink{{Name: "poNumber", Value: "PO-1"}}},
	}
	r := Build(docs)
	if len(r.Relations) != 0 {
		t.Fatalf("expected no cross-dept links without promote: %+v", r.Relations)
	}
	docs[0].Promoted = true
	docs[1].Promoted = true
	r = Build(docs)
	if len(r.Relations) == 0 {
		t.Fatal("expected links after promote")
	}
}

func TestFromValueFile(t *testing.T) {
	d := FromValueFile(label.ValueFile{
		Source: "x.png", Type: "unknown", SchemaStatus: schema.StatusPrivate, RescanNeeded: true,
	})
	if !d.Skip {
		t.Fatal("expected skip")
	}
}
