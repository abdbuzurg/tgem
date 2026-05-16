package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// postDownload sends an authenticated POST and returns the raw response body.
// Used for file-returning endpoints (e.g. /return/report) that respond with a
// binary attachment rather than the JSON envelope.
func postDownload(t *testing.T, path, token string, body any) []byte {
	t.Helper()
	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, helpers.BaseURL()+path, bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s: expected 200, got %d: %s", path, resp.StatusCode, string(raw))
	}
	return raw
}

// TestInvoiceReturn_Team_FullLifecycle exercises the entire team→warehouse
// return flow over HTTP with non-serial-number materials (the dominant case
// in production): create → paginated list → document (xlsx) → unique filters
// → edit → report → confirm (pdf upload) → document (pdf) → delete.
func TestInvoiceReturn_Team_FullLifecycle(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Cable F", "CBL-F", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)

	leader := helpers.Worker(t, 1, "Lead F", "Бригадир")
	team := helpers.Team(t, 1, "T-F", "+992900000001", "Acme", []uint{leader.ID})
	helpers.TeamStock(t, 1, team.ID, cost.ID, 100)

	district := helpers.District(t, 1, "D-F")
	acceptor := helpers.Worker(t, 1, "Acc F", "Кладовщик")

	// --- Create ---
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
		Items: []dto.InvoiceReturnItem{{MaterialID: mat.ID, Amount: 30, Notes: "30 back"}},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/return/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /return/")
	var created model.InvoiceReturn
	helpers.MustDecode(t, createEnv, &created)
	if created.ID == 0 || created.DeliveryCode == "" {
		t.Fatalf("create returned empty invoice: %+v", created)
	}
	idStr := strconv.FormatUint(uint64(created.ID), 10)

	// --- Paginated list (type=team) ---
	listEnv := helpers.AuthedJSON(t, "GET", "/return/paginated?page=1&limit=25&type=team", token, nil)
	helpers.AssertSuccess(t, listEnv, "GET /return/paginated?type=team")
	if !strings.Contains(string(listEnv.Data), created.DeliveryCode) {
		t.Fatalf("paginated list missing %s: %s", created.DeliveryCode, string(listEnv.Data))
	}

	// --- Document download before confirmation (xlsx) ---
	body, _ := helpers.Download(t, "/return/document/"+created.DeliveryCode, token)
	if len(body) == 0 {
		t.Fatalf("xlsx document download empty")
	}

	// --- Unique filters (regression: plural returner_type bug) ---
	uc := helpers.AuthedJSON(t, "GET", "/return/unique/code", token, nil)
	helpers.AssertSuccess(t, uc, "GET /return/unique/code")
	if !strings.Contains(string(uc.Data), created.DeliveryCode) {
		t.Fatalf("unique/code missing %s: %s", created.DeliveryCode, string(uc.Data))
	}
	ut := helpers.AuthedJSON(t, "GET", "/return/unique/team", token, nil)
	helpers.AssertSuccess(t, ut, "GET /return/unique/team")
	if !strings.Contains(string(ut.Data), "T-F") {
		t.Fatalf("unique/team missing team number T-F: %s", string(ut.Data))
	}

	// --- Edit (change amount 30 -> 45) ---
	updateBody := createBody
	updateBody.Details = created
	updateBody.Items = []dto.InvoiceReturnItem{{MaterialID: mat.ID, Amount: 45, Notes: "edited"}}
	updEnv := helpers.AuthedJSON(t, "PATCH", "/return/", token, updateBody)
	helpers.AssertSuccess(t, updEnv, "PATCH /return/")

	// --- Report (type=team) returns a binary xlsx attachment ---
	repBytes := postDownload(t, "/return/report", token, dto.InvoiceReturnReportFilterRequest{
		ReturnerType: "team",
	})
	if len(repBytes) == 0 {
		t.Fatalf("report download empty")
	}

	// --- Confirm (pdf upload) and verify material movement ---
	before := helpers.MaterialLocationSnapshot(t, 1)
	pdfPath := filepath.Join(t.TempDir(), "c.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	confEnv := helpers.MultipartUpload(t, "/return/confirm/"+idStr, token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confEnv, "POST /return/confirm/:id")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)
	if got := diff[helpers.LocKey("warehouse", 0, cost.ID)]; got != 45 {
		t.Errorf("warehouse after confirm = %v, want 45", got)
	}
	if got := diff[helpers.LocKey("team", team.ID, cost.ID)]; got != 55 {
		t.Errorf("team after confirm = %v, want 55 (100-45)", got)
	}

	// --- Document download after confirmation (pdf) ---
	pdfBody, _ := helpers.Download(t, "/return/document/"+created.DeliveryCode, token)
	if len(pdfBody) == 0 {
		t.Fatalf("pdf document download empty")
	}

	// --- Delete ---
	delEnv := helpers.AuthedJSON(t, "DELETE", "/return/"+idStr, token, nil)
	helpers.AssertSuccess(t, delEnv, "DELETE /return/:id")
}

// TestInvoiceReturn_Object_FullLifecycle exercises the object→team return
// variant: an object returns material to a team. Non-serial materials.
func TestInvoiceReturn_Object_FullLifecycle(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Pole O", "POLE-O", "шт")
	cost := helpers.MaterialCost(t, mat.ID, 10.0, 12.0)

	district := helpers.District(t, 1, "D-O")
	leader := helpers.Worker(t, 1, "Lead O", "Бригадир")
	team := helpers.Team(t, 1, "T-O", "+992900000002", "Acme", []uint{leader.ID})

	// Create a SIP object (simplest polymorphic subtype).
	sipEnv := helpers.AuthedJSON(t, "POST", "/sip/", token, dto.SIPObjectCreate{
		BaseInfo:              model.Object{Name: "SIP-Ret-1", Status: "active"},
		DetailedInfo:          model.SIP_Object{AmountFeeders: 2},
		Supervisors:           []uint{},
		Teams:                 []uint{},
		NourashedByTPObjectID: []uint{},
	})
	helpers.AssertSuccess(t, sipEnv, "POST /sip/")
	var sip model.SIP_Object
	helpers.MustDecode(t, sipEnv, &sip)

	var object model.Object
	if err := helpers.DB().
		Where("type = ? AND object_detailed_id = ?", "sip_objects", sip.ID).
		First(&object).Error; err != nil {
		t.Fatalf("find object row: %v", err)
	}

	// Seed object stock (location_type=object).
	objStock := model.MaterialLocation{
		ProjectID:      1,
		MaterialCostID: cost.ID,
		LocationID:     object.ID,
		LocationType:   "object",
		Amount:         50,
	}
	if err := helpers.DB().Create(&objStock).Error; err != nil {
		t.Fatalf("seed object stock: %v", err)
	}

	// --- Create object→team return ---
	createBody := dto.InvoiceReturn{
		Details: model.InvoiceReturn{
			DistrictID:         district.ID,
			ReturnerType:       "object",
			ReturnerID:         object.ID,
			AcceptorType:       "team",
			AcceptorID:         team.ID,
			AcceptedByWorkerID: leader.ID,
			DateOfInvoice:      time.Date(2025, 5, 2, 12, 0, 0, 0, time.UTC),
		},
		Items: []dto.InvoiceReturnItem{{MaterialID: mat.ID, Amount: 20, Notes: "obj back"}},
	}
	createEnv := helpers.AuthedJSON(t, "POST", "/return/", token, createBody)
	helpers.AssertSuccess(t, createEnv, "POST /return/ (object)")
	var created model.InvoiceReturn
	helpers.MustDecode(t, createEnv, &created)
	idStr := strconv.FormatUint(uint64(created.ID), 10)

	// --- Paginated list (type=object) ---
	listEnv := helpers.AuthedJSON(t, "GET", "/return/paginated?page=1&limit=25&type=object", token, nil)
	helpers.AssertSuccess(t, listEnv, "GET /return/paginated?type=object")
	if !strings.Contains(string(listEnv.Data), created.DeliveryCode) {
		t.Fatalf("object paginated list missing %s: %s", created.DeliveryCode, string(listEnv.Data))
	}

	// --- Document download (xlsx) ---
	body, _ := helpers.Download(t, "/return/document/"+created.DeliveryCode, token)
	if len(body) == 0 {
		t.Fatalf("object xlsx document download empty")
	}

	// --- unique/object (regression: plural returner_type bug) ---
	uo := helpers.AuthedJSON(t, "GET", "/return/unique/object", token, nil)
	helpers.AssertSuccess(t, uo, "GET /return/unique/object")
	if !strings.Contains(string(uo.Data), "SIP-Ret-1") {
		t.Fatalf("unique/object missing SIP-Ret-1: %s", string(uo.Data))
	}

	// --- Confirm and verify object→team movement ---
	before := helpers.MaterialLocationSnapshot(t, 1)
	pdfPath := filepath.Join(t.TempDir(), "co.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF, 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
	confEnv := helpers.MultipartUpload(t, "/return/confirm/"+idStr, token, "file", pdfPath, nil)
	helpers.AssertSuccess(t, confEnv, "POST /return/confirm/:id (object)")

	after := helpers.MaterialLocationSnapshot(t, 1)
	diff := before.Diff(after)
	if got := diff[helpers.LocKey("object", object.ID, cost.ID)]; got != 30 {
		t.Errorf("object after confirm = %v, want 30 (50-20)", got)
	}
	if got := diff[helpers.LocKey("team", team.ID, cost.ID)]; got != 20 {
		t.Errorf("team after confirm = %v, want 20", got)
	}
}

// TestInvoiceReturn_OverReturn_GracefulError is a regression test for the
// index-out-of-range panic in buildInvoiceReturnItems: returning more than is
// physically available at the source must yield a normal failure envelope
// (success:false) rather than a 500/panic.
func TestInvoiceReturn_OverReturn_GracefulError(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	mat := helpers.Material(t, 1, "Wire OV", "WIRE-OV", "м")
	cost := helpers.MaterialCost(t, mat.ID, 5.0, 6.0)
	leader := helpers.Worker(t, 1, "Lead OV", "Бригадир")
	team := helpers.Team(t, 1, "T-OV", "+992900000020", "Acme", []uint{leader.ID})
	helpers.TeamStock(t, 1, team.ID, cost.ID, 10) // only 10 available
	district := helpers.District(t, 1, "D-OV")
	acceptor := helpers.Worker(t, 1, "Acc OV", "Кладовщик")

	createEnv := helpers.AuthedJSON(t, "POST", "/return/", token, dto.InvoiceReturn{
		Details: model.InvoiceReturn{
			DistrictID:         district.ID,
			ReturnerType:       "team",
			ReturnerID:         team.ID,
			AcceptorType:       "warehouse",
			AcceptorID:         0,
			AcceptedByWorkerID: acceptor.ID,
			DateOfInvoice:      time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC),
		},
		Items: []dto.InvoiceReturnItem{{MaterialID: mat.ID, Amount: 50}}, // > 10
	})
	msg := helpers.AssertFailure(t, createEnv, "POST /return/ over-return")
	if !strings.Contains(msg, "Недостаточно материала") {
		t.Fatalf("expected insufficient-material error, got: %q", msg)
	}
}
