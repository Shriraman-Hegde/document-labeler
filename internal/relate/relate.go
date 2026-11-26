// Package relate builds typed links and quantity checks from labeled documents.
package relate

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Shriraman-Hegde/document-labeler/internal/label"
	"github.com/Shriraman-Hegde/document-labeler/internal/schema"
)

// Key is a shared identifier on a relation.
type Key struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Relation is a typed link between two documents.
type Relation struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Kind     string   `json:"kind"`
	Keys     []Key    `json:"keys"`
	Warnings []string `json:"warnings,omitempty"`
}

// QuantityEntry is one document's quantity for a part.
type QuantityEntry struct {
	File     string  `json:"file"`
	Role     string  `json:"role"`
	Quantity float64 `json:"quantity"`
	Text     string  `json:"text"`
}

// Check is a quantity reconciliation across documents.
type Check struct {
	Kind    string          `json:"kind"`
	Part    string          `json:"part"`
	Status  string          `json:"status"` // match | mismatch
	Entries []QuantityEntry `json:"entries"`
}

// Doc is the relate view of a labeled value file.
type Doc struct {
	File       string
	Type       schema.DocType
	Department string
	Promoted   bool
	SoftLinks  []label.SoftLink
	Fields     []label.FieldValue
	OCRText    string
	Skip       bool // unknown / rescan
}

// Result is relations + checks for a document set.
type Result struct {
	Relations []Relation `json:"relations"`
	Checks    []Check    `json:"checks"`
}

var qtyPattern = regexp.MustCompile(`(?i)\bquantity[\s#:.-]*(\d+(?:\.\d+)?)`)

// Build computes typed relations and quantity checks.
// Cross-department links are only emitted when both docs are promoted
// (or both share the same department).
func Build(docs []Doc) Result {
	var live []Doc
	for _, d := range docs {
		if d.Skip || d.Type == "" || d.Type == "unknown" {
			continue
		}
		live = append(live, d)
	}
	var rels []Relation
	for i := 0; i < len(live); i++ {
		for j := i + 1; j < len(live); j++ {
			if !linkAllowed(live[i], live[j]) {
				continue
			}
			rels = append(rels, pair(live[i], live[j])...)
		}
	}
	sort.Slice(rels, func(i, j int) bool {
		if rels[i].From != rels[j].From {
			return rels[i].From < rels[j].From
		}
		if rels[i].To != rels[j].To {
			return rels[i].To < rels[j].To
		}
		return rels[i].Kind < rels[j].Kind
	})
	return Result{Relations: rels, Checks: quantityChecks(live)}
}

func linkAllowed(a, b Doc) bool {
	if a.Department == b.Department {
		return true
	}
	return a.Promoted && b.Promoted
}

func pair(a, b Doc) []Relation {
	var out []Relation
	poA, poB := keyVal(a, "poNumber"), keyVal(b, "poNumber")
	samePO := poA != "" && poA == poB
	if samePO {
		kind := "same_po"
		switch {
		case types(a, b, "invoice", "purchase_order"):
			kind = "invoice_for_po"
		case types(a, b, "packing_slip", "purchase_order"):
			kind = "shipment_for_po"
		case types(a, b, "goods_receipt", "purchase_order"):
			kind = "receipt_for_po"
		}
		out = append(out, link(a, b, kind, Key{Name: "poNumber", Value: poA}))
	}

	delA, delB := keyVal(a, "deliveryNumber"), keyVal(b, "deliveryNumber")
	sameDelivery := delA != "" && delA == delB
	parts := shared(partSet(a), partSet(b))
	if types(a, b, "goods_receipt", "packing_slip") && (sameDelivery || (samePO && len(parts) > 0)) {
		var keys []Key
		if sameDelivery {
			keys = append(keys, Key{Name: "deliveryNumber", Value: delA})
		}
		if samePO {
			keys = append(keys, Key{Name: "poNumber", Value: poA})
		}
		for _, p := range parts {
			keys = append(keys, Key{Name: "partNumber", Value: p})
		}
		out = append(out, link(a, b, "receipt_for_shipment", keys...))
	}

	lots := shared(lotSet(a), lotSet(b))
	if len(lots) > 0 {
		kind := "same_lot"
		if oneOf(a, b, "inspection_report") && otherIs(a, b, "inspection_report", "goods_receipt", "packing_slip") {
			kind = "inspection_for_lot"
		}
		var keys []Key
		for _, l := range lots {
			keys = append(keys, Key{Name: "lot", Value: l})
		}
		out = append(out, link(a, b, kind, keys...))
	}

	if len(parts) > 0 && (samePO || len(lots) > 0) {
		var keys []Key
		if samePO {
			keys = append(keys, Key{Name: "poNumber", Value: poA})
		}
		for _, p := range parts {
			keys = append(keys, Key{Name: "partNumber", Value: p})
		}
		for _, l := range lots {
			keys = append(keys, Key{Name: "lot", Value: l})
		}
		out = append(out, link(a, b, "same_part", keys...))
	}

	if da, db := keyVal(a, "drawingNumber"), keyVal(b, "drawingNumber"); da != "" && da == db {
		out = append(out, link(a, b, "same_drawing", Key{Name: "drawingNumber", Value: da}))
	}
	return out
}

func quantityChecks(docs []Doc) []Check {
	type acc struct {
		display string
		entries []QuantityEntry
	}
	byPart := map[string]*acc{}
	for _, d := range docs {
		role := quantityRole(d.Type)
		if role == "" {
			continue
		}
		part := firstPart(d)
		qty, qtyText, ok := firstQty(d)
		if !ok || part == "" {
			continue
		}
		slot := byPart[part]
		if slot == nil {
			slot = &acc{display: part}
			byPart[part] = slot
		}
		slot.entries = append(slot.entries, QuantityEntry{
			File:     d.File,
			Role:     role,
			Quantity: qty,
			Text:     qtyText,
		})
	}
	var parts []string
	for k, slot := range byPart {
		if len(slot.entries) >= 2 {
			parts = append(parts, k)
		}
	}
	sort.Strings(parts)
	var checks []Check
	for _, k := range parts {
		slot := byPart[k]
		sort.Slice(slot.entries, func(i, j int) bool {
			if slot.entries[i].File != slot.entries[j].File {
				return slot.entries[i].File < slot.entries[j].File
			}
			return slot.entries[i].Role < slot.entries[j].Role
		})
		status := "match"
		base := slot.entries[0].Quantity
		for _, e := range slot.entries[1:] {
			if math.Abs(e.Quantity-base) > 0.001 {
				status = "mismatch"
				break
			}
		}
		checks = append(checks, Check{
			Kind:    "quantity",
			Part:    slot.display,
			Status:  status,
			Entries: slot.entries,
		})
	}
	return checks
}

func quantityRole(t schema.DocType) string {
	switch t {
	case "purchase_order":
		return "ordered"
	case "invoice":
		return "invoiced"
	case "packing_slip":
		return "shipped"
	case "goods_receipt":
		return "received"
	case "inspection_report":
		return "inspected"
	default:
		return ""
	}
}

func keyVal(d Doc, name string) string {
	for _, s := range d.SoftLinks {
		if s.Name == name {
			return normalize(s.Value)
		}
	}
	for _, f := range d.Fields {
		if f.Key == name && f.Text != "" {
			return normalize(f.Text)
		}
	}
	return ""
}

func partSet(d Doc) map[string]bool {
	m := map[string]bool{}
	if v := keyVal(d, "partNumber"); v != "" {
		m[v] = true
	}
	return m
}

func lotSet(d Doc) map[string]bool {
	m := map[string]bool{}
	if v := keyVal(d, "lot"); v != "" {
		m[v] = true
	}
	return m
}

func firstPart(d Doc) string {
	return keyVal(d, "partNumber")
}

func firstQty(d Doc) (float64, string, bool) {
	for _, f := range d.Fields {
		if f.Key == "quantity" && f.Text != "" {
			if q, ok := parseQty(f.Text); ok {
				return q, f.Text, true
			}
		}
	}
	if m := qtyPattern.FindStringSubmatch(d.OCRText); len(m) > 1 {
		if q, ok := parseQty(m[1]); ok {
			return q, m[1], true
		}
	}
	return 0, "", false
}

func parseQty(s string) (float64, bool) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func shared(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if b[k] {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func types(a, b Doc, x, y schema.DocType) bool {
	return (a.Type == x && b.Type == y) || (a.Type == y && b.Type == x)
}

func oneOf(a, b Doc, t schema.DocType) bool {
	return a.Type == t || b.Type == t
}

func otherIs(a, b Doc, self schema.DocType, allowed ...schema.DocType) bool {
	other := a
	if a.Type == self {
		other = b
	}
	for _, t := range allowed {
		if other.Type == t {
			return true
		}
	}
	return false
}

func link(a, b Doc, kind string, keys ...Key) Relation {
	from, to := a.File, b.File
	if from > to {
		from, to = to, from
	}
	return Relation{From: from, To: to, Kind: kind, Keys: keys}
}

func normalize(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	return strings.ReplaceAll(s, " ", "")
}

// FromValueFile adapts a value artifact into a Doc.
func FromValueFile(vf label.ValueFile) Doc {
	return Doc{
		File:       vf.Source,
		Type:       vf.Type,
		Department: vf.Department,
		Promoted:   vf.SchemaStatus == schema.StatusPromoted,
		SoftLinks:  vf.SoftLinks,
		Fields:     vf.Fields,
		OCRText:    vf.OCRText,
		Skip:       vf.RescanNeeded || vf.Type == "unknown",
	}
}
