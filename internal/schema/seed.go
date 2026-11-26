package schema

// DefaultSeed returns the v1 manufacturing seed set (KTD7).
func DefaultSeed() []Version {
	seeds := []struct {
		t      DocType
		fields []FieldDef
	}{
		{"invoice", []FieldDef{
			{Key: "invoiceNumber", Label: "Invoice"},
			{Key: "poNumber", Label: "PO Number"},
			{Key: "supplier", Label: "Supplier"},
			{Key: "quantity", Label: "Quantity"},
		}},
		{"purchase_order", []FieldDef{
			{Key: "poNumber", Label: "PO Number"},
			{Key: "supplier", Label: "Supplier"},
			{Key: "partNumber", Label: "Part Number"},
			{Key: "quantity", Label: "Quantity"},
		}},
		{"packing_slip", []FieldDef{
			{Key: "deliveryNumber", Label: "Delivery"},
			{Key: "poNumber", Label: "PO Number"},
			{Key: "partNumber", Label: "Part Number"},
			{Key: "quantity", Label: "Quantity"},
		}},
		{"goods_receipt", []FieldDef{
			{Key: "receiptNumber", Label: "Goods Receipt"},
			{Key: "poNumber", Label: "PO Number"},
			{Key: "partNumber", Label: "Part Number"},
			{Key: "quantity", Label: "Quantity"},
		}},
		{"inspection_report", []FieldDef{
			{Key: "lot", Label: "Lot"},
			{Key: "partNumber", Label: "Part Number"},
			{Key: "inspection", Label: "Inspection"},
		}},
	}
	out := make([]Version, 0, len(seeds))
	for _, s := range seeds {
		out = append(out, Version{
			Type:      s.t,
			Version:   1,
			Status:    StatusPrivate,
			Fields:    s.fields,
			Conflicts: nil,
		})
	}
	return out
}
