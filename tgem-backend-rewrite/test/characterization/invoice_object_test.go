package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"testing"
	"time"
)

// TestInvoiceObject_Create_DoesNotMoveMaterialLocations locks in a non-obvious
// contract: invoice-object Create writes invoice_objects, invoice_materials,
// and invoice_operations rows, but does NOT decrement team stock. Despite the
// "team materials" naming and the on-the-fly cap of item amounts to available
// team stock, the actual material_locations table is untouched. (Reductions
// happen later via invoice-correction or other flows.)
//
// This is exactly the kind of subtle current-state contract a sqlc rewrite
// could accidentally change. The test exists to catch that.
func TestInvoiceObject_Create_DoesNotMoveMaterialLocations(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// Object + team + team stock + district.
	mat := helpers.Material(t, 1, "Wire O", "WIRE-O", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)

	leader := helpers.Worker(t, 1, "Team Leader O", "Бригадир")
	team := helpers.Team(t, 1, "T-O", "+9921111111", "Acme", []uint{leader.ID})
	helpers.TeamStock(t, 1, team.ID, cost.ID, 50)
	district := helpers.District(t, 1, "D-O")

	// SIP object via the typed endpoint, so we exercise polymorphism plumbing.
	sipEnv := helpers.AuthedJSON(t, "POST", "/sip/", token, dto.SIPObjectCreate{
		BaseInfo:              model.Object{Name: "OBJ-O", Status: "active"},
		DetailedInfo:          model.SIP_Object{AmountFeeders: 2},
		Supervisors:           []uint{},
		Teams:                 []uint{team.ID},
		NourashedByTPObjectID: []uint{},
	})
	helpers.AssertSuccess(t, sipEnv, "POST /sip/")
	var sip model.SIP_Object
	helpers.MustDecode(t, sipEnv, &sip)

	var obj model.Object
	if err := helpers.DB().Where("type = ? AND object_detailed_id = ?", "sip_objects", sip.ID).First(&obj).Error; err != nil {
		t.Fatalf("locate sip object row: %v", err)
	}

	before := helpers.MaterialLocationSnapshot(t, 1)

	createBody := dto.InvoiceObjectCreate{
		Details: model.InvoiceObject{
			DistrictID:    district.ID,
			ObjectID:      obj.ID,
			TeamID:        team.ID,
			DateOfInvoice: time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC),
		},
		Items: []dto.InvoiceObjectItem{
			{MaterialID: mat.ID, Amount: 15, Notes: "char object"},
		},
		Operations: []dto.InvoiceObjectOperation{},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/invoice-object/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /invoice-object/")

	// Material locations untouched.
	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)
	if len(diff) != 0 {
		t.Fatalf("expected NO material_locations changes, got %v", diff)
	}

	// invoice_objects row exists with the expected DeliveryCode.
	var io model.InvoiceObject
	if err := helpers.DB().Where("project_id = ?", uint(1)).First(&io).Error; err != nil {
		t.Fatalf("locate invoice_object: %v", err)
	}
	// Note: invoice-object uses prefix "ПО" rather than "О"; same UniqueCodeGeneration format.
	if io.DeliveryCode != "ПО-01-00001" {
		t.Errorf("DeliveryCode = %q, want ПО-01-00001", io.DeliveryCode)
	}

	// invoice_materials row(s) recorded under InvoiceType="object".
	var ims []model.InvoiceMaterials
	if err := helpers.DB().Where("invoice_id = ? AND invoice_type = ?", io.ID, "object").Find(&ims).Error; err != nil {
		t.Fatalf("locate invoice_materials: %v", err)
	}
	if len(ims) == 0 {
		t.Fatal("expected at least one invoice_materials row with invoice_type='object'")
	}
	totalRecorded := 0.0
	for _, im := range ims {
		totalRecorded += im.Amount
	}
	if totalRecorded != 15 {
		t.Errorf("invoice_materials total amount = %v, want 15", totalRecorded)
	}
}
