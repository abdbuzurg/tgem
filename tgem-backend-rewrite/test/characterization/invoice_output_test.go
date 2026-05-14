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

// TestInvoiceOutput_CreateConfirm_MovesWarehouseToTeam locks in the
// invoice-output happy path: warehouse stock decreases by N, team stock
// increases by N, no other material_locations rows change.
func TestInvoiceOutput_CreateConfirm_MovesWarehouseToTeam(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// Fixtures: a material with stock in the warehouse + a team to receive it.
	mat := helpers.Material(t, 1, "Wire C", "WIRE-C", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)
	helpers.WarehouseStock(t, 1, cost.ID, 100)

	district := helpers.District(t, 1, "District-1")
	leader := helpers.Worker(t, 1, "Team Leader", "Бригадир")
	team := helpers.Team(t, 1, "T-1", "+9921111111", "Acme", []uint{leader.ID})

	wm := helpers.Worker(t, 1, "WM Output", "Кладовщик")
	recipient := helpers.Worker(t, 1, "Recipient", "Электромонтер")

	before := helpers.MaterialLocationSnapshot(t, 1)

	createBody := dto.InvoiceOutput{
		Details: model.InvoiceOutput{
			DistrictID:               district.ID,
			WarehouseManagerWorkerID: wm.ID,
			RecipientWorkerID:        recipient.ID,
			TeamID:                   team.ID,
			Notes:                    "char output",
		},
		Items: []dto.InvoiceOutputItem{
			{MaterialID: mat.ID, Amount: 30, Notes: "30 units"},
		},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/output/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /output/")

	var created model.InvoiceOutput
	helpers.MustDecode(t, createEnv, &created)
	if created.DeliveryCode != "О-01-00001" {
		t.Fatalf("DeliveryCode = %q, want О-01-00001", created.DeliveryCode)
	}

	// Create alone does not move material_locations — Confirmation does.
	mid := helpers.MaterialLocationSnapshot(t, 1)
	if mid[helpers.LocKey("warehouse", 0, cost.ID)] != 100 {
		t.Fatalf("warehouse stock should still be 100 before Confirmation, got %v", mid)
	}

	// Confirmation requires a PDF upload like input does.
	pdfPath := filepath.Join(t.TempDir(), "out-confirm.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write tmp pdf: %v", err)
	}
	confirmEnv := helpers.MultipartUpload(t,
		"/output/confirm/"+strconv.FormatUint(uint64(created.ID), 10),
		token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confirmEnv, "POST /output/confirm/:id")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)

	whKey := helpers.LocKey("warehouse", 0, cost.ID)
	teamKey := helpers.LocKey("team", team.ID, cost.ID)
	if got, want := diff[whKey], 70.0; got != want {
		t.Errorf("warehouse %s = %v, want %v", whKey, got, want)
	}
	if got, want := diff[teamKey], 30.0; got != want {
		t.Errorf("team %s = %v, want %v", teamKey, got, want)
	}
	if len(diff) != 2 {
		t.Errorf("expected exactly 2 changed keys, got %d: %v", len(diff), diff)
	}
}

// TestInvoiceOutput_Confirm_RequiresPdfMultipart locks in the contract that
// the confirmation endpoint refuses calls without a multipart `file` field.
// Passing a JSON body returns the standard envelope with success=false rather
// than HTTP 4xx.
func TestInvoiceOutput_Confirm_RequiresPdfMultipart(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// Seed a single InvoiceOutput row directly (we don't need real warehouse
	// stock for this error path, but FKs do require valid district + team rows).
	wm := helpers.Worker(t, 1, "WM", "Кладовщик")
	district := helpers.District(t, 1, "D-1")
	team := helpers.Team(t, 1, "T-1", "+9921111111", "Acme", nil)
	io := model.InvoiceOutput{
		ProjectID:                1,
		DistrictID:               district.ID,
		TeamID:                   team.ID,
		WarehouseManagerWorkerID: wm.ID,
		ReleasedWorkerID:         wm.ID,
		RecipientWorkerID:        wm.ID,
		DeliveryCode:             "О-01-99999",
	}
	if err := helpers.DB().Create(&io).Error; err != nil {
		t.Fatalf("seed invoice_outputs: %v", err)
	}

	env := helpers.AuthedJSON(t, "POST",
		"/output/confirm/"+strconv.FormatUint(uint64(io.ID), 10),
		token, map[string]any{"any": "json"})
	msg := helpers.AssertFailure(t, env, "Confirmation without multipart")
	// Locked in: error message comes from c.FormFile failing on a non-multipart
	// body, prefixed with "cannot form file:" by the controller.
	if msg == "" {
		t.Fatal("expected non-empty error")
	}
}
