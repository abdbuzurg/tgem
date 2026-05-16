package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"strings"
	"testing"
	"time"
)

// TestInvoiceCorrection_LeaderlessTeam_AppearsInCorrectionList is a regression
// test for the INNER-JOIN list-drop bug: an unconfirmed invoice-object whose
// team has NO leader must still appear in the correction queue. Before the
// fix, ListInvoiceCorrectionsPaginated INNER JOINed team_leaders, so any
// invoice-object on a leaderless team was uncorrectable (count > 0 but data
// empty — the exact production symptom). Verified to fail when the join is
// reverted to INNER.
func TestInvoiceCorrection_LeaderlessTeam_AppearsInCorrectionList(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Wire LL", "WIRE-LL", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)
	// Team with NO leader.
	team := helpers.Team(t, 1, "T-LL", "+992900000011", "Acme", []uint{})
	helpers.TeamStock(t, 1, team.ID, cost.ID, 50)
	district := helpers.District(t, 1, "D-LL")

	// Object with NO supervisors.
	sipEnv := helpers.AuthedJSON(t, "POST", "/sip/", token, dto.SIPObjectCreate{
		BaseInfo:              model.Object{Name: "OBJ-LL", Status: "active"},
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

	// Unconfirmed invoice-object → belongs in the correction queue.
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

	listEnv := helpers.AuthedJSON(t, "GET", "/invoice-correction/paginated?page=1&limit=25", token, nil)
	helpers.AssertSuccess(t, listEnv, "GET /invoice-correction/paginated")
	if !strings.Contains(string(listEnv.Data), "ПО-01-00001") {
		t.Fatalf("leaderless-team invoice-object missing from correction list: %s", string(listEnv.Data))
	}
}
