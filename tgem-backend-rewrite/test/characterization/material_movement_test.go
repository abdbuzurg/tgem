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

// TestMaterialLocation_Live_AfterMultiStepLifecycle drives a full
// input → output → return chain through the public API and asserts that
// /material-location/live exposes the resulting balances per location type.
// This is the cross-cutting safety net that catches a sqlc rewrite that
// silently changes how multiple invoice flavors aggregate into the live view.
func TestMaterialLocation_Live_AfterMultiStepLifecycle(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Wire X-cut", "WIRE-XCUT", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)

	leader := helpers.Worker(t, 1, "Leader X", "Бригадир")
	team := helpers.Team(t, 1, "T-X", "+9921111111", "Acme", []uint{leader.ID})
	district := helpers.District(t, 1, "D-X")
	wm := helpers.Worker(t, 1, "WM X", "Кладовщик")
	recipient := helpers.Worker(t, 1, "Recipient X", "Электромонтер")
	acceptor := helpers.Worker(t, 1, "Acceptor X", "Кладовщик")

	pdf := func(name string) string {
		p := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(p, minimalPDF, 0o644); err != nil {
			t.Fatalf("write pdf: %v", err)
		}
		return p
	}

	// Step 1: input 50 → warehouse 50.
	createInput := helpers.AuthedJSON(t, "POST", "/input/", token, dto.InvoiceInput{
		Details: model.InvoiceInput{WarehouseManagerWorkerID: wm.ID},
		Items: []dto.InvoiceInputMaterial{
			{MaterialData: model.InvoiceMaterials{MaterialCostID: cost.ID, Amount: 50}},
		},
	})
	helpers.AssertSuccess(t, createInput, "input create")
	var inv model.InvoiceInput
	helpers.MustDecode(t, createInput, &inv)
	confInput := helpers.MultipartUpload(t,
		"/input/confirm/"+strconv.FormatUint(uint64(inv.ID), 10),
		token, "file", pdf("in.pdf"), nil)
	helpers.AssertSuccess(t, confInput, "input confirm")

	// Step 2: output 30 to team → warehouse 20, team 30.
	createOut := helpers.AuthedJSON(t, "POST", "/output/", token, dto.InvoiceOutput{
		Details: model.InvoiceOutput{
			DistrictID:               district.ID,
			WarehouseManagerWorkerID: wm.ID,
			RecipientWorkerID:        recipient.ID,
			TeamID:                   team.ID,
		},
		Items: []dto.InvoiceOutputItem{{MaterialID: mat.ID, Amount: 30}},
	})
	helpers.AssertSuccess(t, createOut, "output create")
	var out model.InvoiceOutput
	helpers.MustDecode(t, createOut, &out)
	confOut := helpers.MultipartUpload(t,
		"/output/confirm/"+strconv.FormatUint(uint64(out.ID), 10),
		token, "file", pdf("out.pdf"), nil)
	helpers.AssertSuccess(t, confOut, "output confirm")

	// Step 3: return 10 team → warehouse → warehouse 30, team 20.
	createRet := helpers.AuthedJSON(t, "POST", "/return/", token, dto.InvoiceReturn{
		Details: model.InvoiceReturn{
			DistrictID:         district.ID,
			ReturnerType:       "team",
			ReturnerID:         team.ID,
			AcceptorType:       "warehouse",
			AcceptorID:         0,
			AcceptedByWorkerID: acceptor.ID,
		},
		Items: []dto.InvoiceReturnItem{{MaterialID: mat.ID, Amount: 10}},
	})
	helpers.AssertSuccess(t, createRet, "return create")
	var ret model.InvoiceReturn
	helpers.MustDecode(t, createRet, &ret)
	confRet := helpers.MultipartUpload(t,
		"/return/confirm/"+strconv.FormatUint(uint64(ret.ID), 10),
		token, "file", pdf("ret.pdf"), nil)
	helpers.AssertSuccess(t, confRet, "return confirm")

	// Direct DB sanity check — the source of truth.
	snap := helpers.MaterialLocationSnapshot(t, 1)
	if got := snap[helpers.LocKey("warehouse", 0, cost.ID)]; got != 30 {
		t.Errorf("after lifecycle, warehouse = %v, want 30", got)
	}
	if got := snap[helpers.LocKey("team", team.ID, cost.ID)]; got != 20 {
		t.Errorf("after lifecycle, team = %v, want 20", got)
	}

	// Live endpoint mirrors the snapshot for warehouse.
	liveWh := helpers.AuthedJSON(t, "GET", "/material-location/live?locationType=warehouse", token, nil)
	helpers.AssertSuccess(t, liveWh, "GET /material-location/live (warehouse)")
	var whRows []dto.MaterialLocationLiveView
	helpers.MustDecode(t, liveWh, &whRows)
	if len(whRows) != 1 {
		t.Fatalf("expected 1 warehouse row, got %d (%+v)", len(whRows), whRows)
	}
	if whRows[0].Amount != 30 {
		t.Errorf("warehouse live amount = %v, want 30", whRows[0].Amount)
	}

	// And for team.
	liveTeam := helpers.AuthedJSON(t, "GET", "/material-location/live?locationType=team", token, nil)
	helpers.AssertSuccess(t, liveTeam, "GET /material-location/live (team)")
	var teamRows []dto.MaterialLocationLiveView
	helpers.MustDecode(t, liveTeam, &teamRows)
	if len(teamRows) != 1 {
		t.Fatalf("expected 1 team row, got %d (%+v)", len(teamRows), teamRows)
	}
	if teamRows[0].Amount != 20 {
		t.Errorf("team live amount = %v, want 20", teamRows[0].Amount)
	}
}

// TestMaterialLocation_UniqueTeams_OnlyShowsTeamsWithStock locks in the
// /material-location/unique/team contract: the dropdown is filtered on
// material_locations.amount > 0 AND location_type='team'. After the
// lifecycle drives team stock to zero, the team disappears from the dropdown.
func TestMaterialLocation_UniqueTeams_OnlyShowsTeamsWithStock(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// One team that holds stock; one team that does NOT.
	mat := helpers.Material(t, 1, "Wire UT", "WIRE-UT", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)

	leader1 := helpers.Worker(t, 1, "L1", "Бригадир")
	leader2 := helpers.Worker(t, 1, "L2", "Бригадир")
	hasStock := helpers.Team(t, 1, "T-HAS", "+9921111111", "Acme", []uint{leader1.ID})
	helpers.TeamStock(t, 1, hasStock.ID, cost.ID, 25)
	noStock := helpers.Team(t, 1, "T-NONE", "+9921111112", "Acme", []uint{leader2.ID})

	env := helpers.AuthedJSON(t, "GET", "/material-location/unique/team", token, nil)
	helpers.AssertSuccess(t, env, "GET /material-location/unique/team")

	var rows []dto.TeamDataForSelect
	helpers.MustDecode(t, env, &rows)

	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 team in dropdown, got %d (%+v)", len(rows), rows)
	}
	if rows[0].ID != hasStock.ID {
		t.Errorf("expected team %d (T-HAS), got %d", hasStock.ID, rows[0].ID)
	}
	if rows[0].TeamLeaderName != "L1" {
		t.Errorf("TeamLeaderName = %q, want L1", rows[0].TeamLeaderName)
	}
	_ = noStock // documented presence; deliberately absent from the dropdown
}
