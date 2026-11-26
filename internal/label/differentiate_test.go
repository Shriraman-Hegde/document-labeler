package label

import (
	"testing"

	"github.com/Shriraman-Hegde/document-labeler/internal/llm"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

func TestFilenameNotSoleAuthority(t *testing.T) {
	d := Differentiate(DiffInput{
		OCRText:  "patient visit summary",
		Filename: "patient-invoice.pdf",
		Assist:   llm.AssistResponse{DocType: "unknown", Confidence: 0.2},
	})
	if d.Type != "unknown" {
		t.Fatalf("filename alone must not classify, got %s", d.Type)
	}
}

func TestHeadingsAndAI(t *testing.T) {
	d := Differentiate(DiffInput{
		OCRText:  "Invoice\nPO Number PO-55501\nSupplier Acme",
		Filename: "scan001.png",
		Assist:   llm.AssistResponse{DocType: "invoice", Confidence: 0.9, Headings: []string{"invoice"}},
	})
	if d.Type != "invoice" {
		t.Fatalf("got %s", d.Type)
	}
}

func TestSelectModeOverride(t *testing.T) {
	m := SelectMode(schema.ModeFullyLearned, "invoice", nil, 0.9)
	if m != schema.ModeFullyLearned {
		t.Fatal(m)
	}
}

func TestSoftKeys(t *testing.T) {
	links := ExtractSoftKeys("PO Number: PO-55501 Part PN-100")
	if len(links) == 0 {
		t.Fatal("expected soft keys")
	}
}
