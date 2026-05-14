package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// minimalPDF is the smallest spec-conformant PDF we can attach to the
// confirmation upload. The Confirmation handler only checks the .pdf
// extension; bytes don't matter beyond that.
var minimalPDF = []byte("%PDF-1.0\n%%EOF\n")

// TestInvoiceInput_CreateConfirm_IncrementsWarehouse exercises the full
// invoice-input lifecycle and asserts that confirmation increments the
// warehouse material_locations row by the invoice item amount.
func TestInvoiceInput_CreateConfirm_IncrementsWarehouse(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// Fixtures: a material with one cost variant + warehouse-manager worker.
	mat := helpers.Material(t, 1, "Cable A", "CBL-A", "м")
	cost := helpers.MaterialCost(t, mat.ID, 10.0, 12.0)
	wm := helpers.Worker(t, 1, "Warehouse Manager", "Кладовщик")

	before := helpers.MaterialLocationSnapshot(t, 1)
	if len(before) != 0 {
		t.Fatalf("expected empty material_locations before invoice-input, got %v", before.Keys())
	}

	// 1. Create.
	createBody := dto.InvoiceInput{
		Details: model.InvoiceInput{
			WarehouseManagerWorkerID: wm.ID,
			Notes:                    "char input",
		},
		Items: []dto.InvoiceInputMaterial{
			{
				MaterialData: model.InvoiceMaterials{
					MaterialCostID: cost.ID,
					Amount:         25,
					Notes:          "row 1",
				},
				SerialNumbers: []string{},
			},
		},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/input/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /input/")

	var created model.InvoiceInput
	helpers.MustDecode(t, createEnv, &created)
	if created.ID == 0 {
		t.Fatal("expected created invoice id > 0")
	}
	if created.DeliveryCode != "П-01-00001" {
		t.Fatalf("DeliveryCode = %q, want П-01-00001", created.DeliveryCode)
	}
	if created.Confirmed {
		t.Fatal("freshly created invoice should have Confirmed=false")
	}

	// Before confirmation, no material_locations rows yet.
	mid := helpers.MaterialLocationSnapshot(t, 1)
	if len(mid) != 0 {
		t.Fatalf("expected no material_locations before Confirmation, got %v", mid)
	}

	// 2. Confirmation requires a multipart PDF upload.
	pdfPath := filepath.Join(t.TempDir(), "char-confirm.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write tmp pdf: %v", err)
	}
	confirmEnv := helpers.MultipartUpload(t,
		"/input/confirm/"+strconv.FormatUint(uint64(created.ID), 10),
		token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confirmEnv, "POST /input/confirm/:id")

	// 3. Diff: a new warehouse:0:<costID> key with amount 25.
	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)
	wantKey := helpers.LocKey("warehouse", 0, cost.ID)
	if got := diff[wantKey]; got != 25 {
		t.Fatalf("warehouse diff for %s = %v, want 25 (full diff %v)", wantKey, got, diff)
	}
	if len(diff) != 1 {
		t.Fatalf("expected exactly one changed key, got %v", diff)
	}

	helpers.AssertJSONGolden(t, "invoice_input/create_response", createEnv.Data)
}

// TestInvoiceInput_GetPaginated_AfterCreate locks in the paginated envelope
// shape and projection columns after a single confirmed input exists.
func TestInvoiceInput_GetPaginated_AfterCreate(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Cable B", "CBL-B", "м")
	cost := helpers.MaterialCost(t, mat.ID, 10.0, 12.0)
	wm := helpers.Worker(t, 1, "WM 1", "Кладовщик")

	createEnv := helpers.AuthedJSON(t, "POST", "/input/", token, dto.InvoiceInput{
		Details: model.InvoiceInput{
			WarehouseManagerWorkerID: wm.ID,
			Notes:                    "p1",
		},
		Items: []dto.InvoiceInputMaterial{
			{MaterialData: model.InvoiceMaterials{MaterialCostID: cost.ID, Amount: 10}},
		},
	})
	helpers.AssertSuccess(t, createEnv, "POST /input/")

	// /input/paginated has a quirk: filter params default to "" then strconv.Atoi
	// that empty string, so explicit zero values are required even for "all".
	const q = "?page=1&limit=10&warehouseManagerWorkerID=0&releasedWorkerID=0"
	listEnv := helpers.AuthedJSON(t, "GET", "/input/paginated"+q, token, nil)
	helpers.AssertSuccess(t, listEnv, "GET /input/paginated")

	helpers.AssertJSONGolden(t, "invoice_input/paginated_after_create", listEnv.Data)
}
