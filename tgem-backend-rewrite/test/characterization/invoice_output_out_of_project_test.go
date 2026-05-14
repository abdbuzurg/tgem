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

// TestInvoiceOutputOutOfProject_CreateConfirm exercises the "ship to another
// project" lifecycle. Lock-in: warehouse:0:X decreases by N AND a new
// out-of-project:0:X row appears with N. The out-of-project location uses
// LocationID=0 across the board (it does not key off NameOfProject). Phase 7
// fix: this flavor now uses its own "output-out-of-project" invoice counter
// and a distinct "ОВ" prefix, so codes no longer collide with regular outputs
// (the project-Турсунзода "G-2" bug).
func TestInvoiceOutputOutOfProject_CreateConfirm(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Wire OOP", "WIRE-OOP", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)
	helpers.WarehouseStock(t, 1, cost.ID, 60)

	before := helpers.MaterialLocationSnapshot(t, 1)

	createBody := dto.InvoiceOutputOutOfProject{
		Details: model.InvoiceOutputOutOfProject{
			NameOfProject: "External Project Alpha",
			Notes:         "char oop",
		},
		Items: []dto.InvoiceOutputItem{
			{MaterialID: mat.ID, Amount: 12, Notes: "12 units"},
		},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/invoice-output-out-of-project/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /invoice-output-out-of-project/")

	var created model.InvoiceOutputOutOfProject
	helpers.MustDecode(t, createEnv, &created)
	if created.DeliveryCode != "ОВ-01-00001" {
		t.Fatalf("DeliveryCode = %q, want ОВ-01-00001 (own counter + ОВ prefix)", created.DeliveryCode)
	}

	pdfPath := filepath.Join(t.TempDir(), "oop-confirm.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write tmp pdf: %v", err)
	}
	confirmEnv := helpers.MultipartUpload(t,
		"/invoice-output-out-of-project/confirm/"+strconv.FormatUint(uint64(created.ID), 10),
		token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confirmEnv, "POST /invoice-output-out-of-project/confirm/:id")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)

	whKey := helpers.LocKey("warehouse", 0, cost.ID)
	if got, want := diff[whKey], 48.0; got != want {
		t.Errorf("warehouse %s = %v, want %v (60 - 12)", whKey, got, want)
	}

	oopKey := helpers.LocKey("out-of-project", 0, cost.ID)
	if got, want := diff[oopKey], 12.0; got != want {
		t.Errorf("out-of-project %s = %v, want %v (newly created at locationID=0)", oopKey, got, want)
	}
}
