package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// TestInvoiceReturn_TeamToWarehouse_HappyPath returns 20 units of stock from a
// team back to the warehouse. Locks in the contract that the returner location
// decreases by N and the acceptor location increases by N — and that an
// existing acceptor row is updated rather than duplicated.
func TestInvoiceReturn_TeamToWarehouse_HappyPath(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// Material + cost.
	mat := helpers.Material(t, 1, "Wire R", "WIRE-R", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)

	// Team holds 80 units.
	leader := helpers.Worker(t, 1, "Team Leader R", "Бригадир")
	team := helpers.Team(t, 1, "T-R", "+9921111111", "Acme", []uint{leader.ID})
	helpers.TeamStock(t, 1, team.ID, cost.ID, 80)

	// Warehouse already has 5 units of the same cost variant — Confirmation
	// should bump that row to 25, not create a second row.
	helpers.WarehouseStock(t, 1, cost.ID, 5)

	district := helpers.District(t, 1, "D-R")
	acceptor := helpers.Worker(t, 1, "Acceptor", "Кладовщик")

	before := helpers.MaterialLocationSnapshot(t, 1)

	createBody := dto.InvoiceReturn{
		Details: model.InvoiceReturn{
			DistrictID:         district.ID,
			ReturnerType:       "team",
			ReturnerID:         team.ID,
			AcceptorType:       "warehouse",
			AcceptorID:         0,
			AcceptedByWorkerID: acceptor.ID,
			DateOfInvoice:      time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC),
		},
		Items: []dto.InvoiceReturnItem{
			{MaterialID: mat.ID, Amount: 20, Notes: "20 back"},
		},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/return/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /return/")

	var created model.InvoiceReturn
	helpers.MustDecode(t, createEnv, &created)
	if created.DeliveryCode != "В-01-00001" {
		t.Fatalf("DeliveryCode = %q, want В-01-00001", created.DeliveryCode)
	}

	pdfPath := filepath.Join(t.TempDir(), "ret-confirm.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write tmp pdf: %v", err)
	}
	confirmEnv := helpers.MultipartUpload(t,
		"/return/confirm/"+strconv.FormatUint(uint64(created.ID), 10),
		token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confirmEnv, "POST /return/confirm/:id")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)

	whKey := helpers.LocKey("warehouse", 0, cost.ID)
	teamKey := helpers.LocKey("team", team.ID, cost.ID)
	if got, want := diff[whKey], 25.0; got != want {
		t.Errorf("warehouse %s = %v, want %v (existing 5 + returned 20)", whKey, got, want)
	}
	if got, want := diff[teamKey], 60.0; got != want {
		t.Errorf("team %s = %v, want %v (was 80 - returned 20)", teamKey, got, want)
	}
}
