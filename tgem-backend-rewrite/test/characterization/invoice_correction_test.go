package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"testing"
	"time"
)

// TestInvoiceCorrection_AdjustsTeamAndObjectLocations exercises the
// invoice-correction flow that *actually* moves stock from team to object.
// Locks in the contract: after correction, (team, teamID, X) decreases and
// (object, objectID, X) increases by the same amount, AND the underlying
// InvoiceObject row is flipped to ConfirmedByOperator=true.
func TestInvoiceCorrection_AdjustsTeamAndObjectLocations(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// Material + team stock + object + invoice-object as preconditions.
	mat := helpers.Material(t, 1, "Wire Corr", "WIRE-CORR", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)
	leader := helpers.Worker(t, 1, "Team Leader Corr", "Бригадир")
	team := helpers.Team(t, 1, "T-Corr", "+9921111111", "Acme", []uint{leader.ID})
	helpers.TeamStock(t, 1, team.ID, cost.ID, 50)
	district := helpers.District(t, 1, "D-Corr")

	// SIP object via the typed endpoint.
	sipEnv := helpers.AuthedJSON(t, "POST", "/sip/", token, dto.SIPObjectCreate{
		BaseInfo:              model.Object{Name: "OBJ-Corr", Status: "active"},
		DetailedInfo:          model.SIP_Object{AmountFeeders: 1},
		Supervisors:           []uint{},
		Teams:                 []uint{team.ID},
		NourashedByTPObjectID: []uint{},
	})
	helpers.AssertSuccess(t, sipEnv, "POST /sip/")
	var sip model.SIP_Object
	helpers.MustDecode(t, sipEnv, &sip)
	var obj model.Object
	if err := helpers.DB().Where("type = ? AND object_detailed_id = ?", "sip_objects", sip.ID).First(&obj).Error; err != nil {
		t.Fatalf("locate sip object: %v", err)
	}

	// Pre-existing InvoiceObject (correction edits an existing record).
	ioEnv := helpers.AuthedJSON(t, "POST", "/invoice-object/", token, dto.InvoiceObjectCreate{
		Details: model.InvoiceObject{
			DistrictID:    district.ID,
			ObjectID:      obj.ID,
			TeamID:        team.ID,
			DateOfInvoice: time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC),
		},
		Items:      []dto.InvoiceObjectItem{{MaterialID: mat.ID, Amount: 10}},
		Operations: []dto.InvoiceObjectOperation{},
	})
	helpers.AssertSuccess(t, ioEnv, "POST /invoice-object/")

	var io model.InvoiceObject
	if err := helpers.DB().Where("project_id = ?", uint(1)).First(&io).Error; err != nil {
		t.Fatalf("locate invoice_object: %v", err)
	}
	if io.ConfirmedByOperator {
		t.Fatal("freshly created invoice_object should have ConfirmedByOperator=false")
	}

	before := helpers.MaterialLocationSnapshot(t, 1)

	// 2. Correction: redirect 8 units from team to object.
	corrBody := dto.InvoiceCorrectionCreate{
		Details: dto.InvoiceCorrectionCreateDetails{
			InvoiceObjectID:  io.ID,
			DateOfCorrection: time.Date(2025, 5, 2, 12, 0, 0, 0, time.UTC),
			// OperatorWorkerID is overwritten by the controller from the auth token.
		},
		Items: []dto.InvoiceCorrectionMaterialsData{
			{MaterialID: mat.ID, MaterialAmount: 8, Notes: "corr"},
		},
		Operations: []dto.InvoiceCorrectionOperationsData{},
	}
	corrEnv := helpers.AuthedJSON(t, "POST", "/invoice-correction/", token, corrBody)
	helpers.AssertSuccess(t, corrEnv, "POST /invoice-correction/")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)

	teamKey := helpers.LocKey("team", team.ID, cost.ID)
	objKey := helpers.LocKey("object", obj.ID, cost.ID)
	if got, want := diff[teamKey], 42.0; got != want {
		t.Errorf("team %s = %v, want %v (50 - 8)", teamKey, got, want)
	}
	if got, want := diff[objKey], 8.0; got != want {
		t.Errorf("object %s = %v, want %v (newly created)", objKey, got, want)
	}

	// InvoiceObject was flipped.
	var refreshed model.InvoiceObject
	if err := helpers.DB().First(&refreshed, io.ID).Error; err != nil {
		t.Fatalf("reload invoice_object: %v", err)
	}
	if !refreshed.ConfirmedByOperator {
		t.Error("after correction, ConfirmedByOperator should be true")
	}
}
