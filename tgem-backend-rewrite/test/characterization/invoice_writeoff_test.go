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

// TestInvoiceWriteOff_LossWarehouse_DecrementsAndCreatesWriteOffLocation
// covers the simplest writeoff variant: stock disappears from warehouse and a
// matching "loss-warehouse" tracking row appears at locationID=0. Note the
// writeOffType "loss-warehouse" doubles as the LocationType of the destination
// row — that pattern is the contract the migration must preserve.
//
// We deliberately do not exercise "loss-team": the Confirmation hardcodes
// locationID=0 when reading team stock (invoice_writeoff_service.go:335),
// which doesn't match how output places team stock at locationID=teamID. That
// looks like a real bug; characterization-test scope is to lock the current
// shape, not exercise broken paths. A fix is queued for phase 7.
func TestInvoiceWriteOff_LossWarehouse_DecrementsAndCreatesWriteOffLocation(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Lost Wire", "WIRE-LOSS", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)
	helpers.WarehouseStock(t, 1, cost.ID, 50)

	before := helpers.MaterialLocationSnapshot(t, 1)

	createBody := dto.InvoiceWriteOff{
		Details: model.InvoiceWriteOff{
			WriteOffType:       "loss-warehouse",
			WriteOffLocationID: 0,
			Notes:              "char writeoff",
		},
		Items: []dto.InvoiceWriteOffItem{
			{MaterialID: mat.ID, Amount: 7, Notes: "lost"},
		},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/invoice-writeoff/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /invoice-writeoff/")

	var created model.InvoiceWriteOff
	helpers.MustDecode(t, createEnv, &created)
	if created.DeliveryCode != "С-01-00001" {
		t.Fatalf("DeliveryCode = %q, want С-01-00001", created.DeliveryCode)
	}

	pdfPath := filepath.Join(t.TempDir(), "writeoff-confirm.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write tmp pdf: %v", err)
	}
	confirmEnv := helpers.MultipartUpload(t,
		"/invoice-writeoff/confirm/"+strconv.FormatUint(uint64(created.ID), 10),
		token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confirmEnv, "POST /invoice-writeoff/confirm/:id")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)

	whKey := helpers.LocKey("warehouse", 0, cost.ID)
	if got, want := diff[whKey], 43.0; got != want {
		t.Errorf("warehouse %s = %v, want %v (50 - 7)", whKey, got, want)
	}

	lossKey := helpers.LocKey("loss-warehouse", 0, cost.ID)
	if got, want := diff[lossKey], 7.0; got != want {
		t.Errorf("loss-warehouse %s = %v, want %v (newly created)", lossKey, got, want)
	}
}
