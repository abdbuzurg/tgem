package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"strconv"
	"testing"
)

// TestObject_CreateTP_WritesTwoRows locks in the polymorphism contract:
// POST /api/tp/ creates one row in tp_objects and one row in objects with
// type='tp_objects' (note plural-s) pointing back via object_detailed_id.
func TestObject_CreateTP_WritesTwoRows(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	body := dto.TPObjectCreate{
		BaseInfo: model.Object{
			Name:   "TP-Char-1",
			Status: "active",
		},
		DetailedInfo: model.TP_Object{
			Model:        "TM-160",
			VoltageClass: "10/0.4",
		},
		Supervisors: []uint{},
		Teams:       []uint{},
	}

	env := helpers.AuthedJSON(t, "POST", "/tp/", token, body)
	helpers.AssertSuccess(t, env, "POST /tp/")

	var created model.TP_Object
	helpers.MustDecode(t, env, &created)
	if created.ID == 0 {
		t.Fatalf("expected TP_Object.ID > 0, got %+v", created)
	}
	if created.Model != "TM-160" {
		t.Fatalf("expected model TM-160, got %q", created.Model)
	}

	var object model.Object
	if err := helpers.DB().
		Where("type = ? AND object_detailed_id = ?", "tp_objects", created.ID).
		First(&object).Error; err != nil {
		t.Fatalf("expected Object row pointing at tp_objects.id=%d: %v", created.ID, err)
	}
	if object.Name != "TP-Char-1" {
		t.Fatalf("Object.Name = %q, want TP-Char-1", object.Name)
	}
	if object.ProjectID != 1 {
		t.Fatalf("Object.ProjectID = %d, want 1 (set from auth context)", object.ProjectID)
	}

	var tpRow model.TP_Object
	if err := helpers.DB().First(&tpRow, created.ID).Error; err != nil {
		t.Fatalf("expected tp_objects row id=%d: %v", created.ID, err)
	}
	if tpRow.VoltageClass != "10/0.4" {
		t.Fatalf("tp_objects.voltageClass = %q, want 10/0.4", tpRow.VoltageClass)
	}

	helpers.AssertJSONGolden(t, "object_polymorphism/create_tp", env.Data)
}

// TestObject_CreateSIP_SimplestSubtype exercises the same two-row pattern via
// the subtype with the fewest required fields.
func TestObject_CreateSIP_SimplestSubtype(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	body := dto.SIPObjectCreate{
		BaseInfo: model.Object{
			Name:   "SIP-Char-1",
			Status: "active",
		},
		DetailedInfo: model.SIP_Object{
			AmountFeeders: 4,
		},
		Supervisors:           []uint{},
		Teams:                 []uint{},
		NourashedByTPObjectID: []uint{},
	}

	env := helpers.AuthedJSON(t, "POST", "/sip/", token, body)
	helpers.AssertSuccess(t, env, "POST /sip/")

	var created model.SIP_Object
	helpers.MustDecode(t, env, &created)
	if created.ID == 0 {
		t.Fatalf("expected SIP_Object.ID > 0, got %+v", created)
	}
	if created.AmountFeeders != 4 {
		t.Fatalf("AmountFeeders = %d, want 4", created.AmountFeeders)
	}

	var object model.Object
	if err := helpers.DB().
		Where("type = ? AND object_detailed_id = ?", "sip_objects", created.ID).
		First(&object).Error; err != nil {
		t.Fatalf("expected Object row pointing at sip_objects.id=%d: %v", created.ID, err)
	}
	if object.Name != "SIP-Char-1" {
		t.Fatalf("Object.Name = %q, want SIP-Char-1", object.Name)
	}
}

// TestObject_GetByID_ReturnsOnlyObjectRow locks in the lazy-loading shape:
// GET /object/:id returns only the model.Object polymorphism shell, not the
// specialized fields. Clients that need TP/SIP details must call the typed
// endpoint.
func TestObject_GetByID_ReturnsOnlyObjectRow(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	createEnv := helpers.AuthedJSON(t, "POST", "/tp/", token, dto.TPObjectCreate{
		BaseInfo:     model.Object{Name: "TP-GetByID", Status: "active"},
		DetailedInfo: model.TP_Object{Model: "TM-100", VoltageClass: "6/0.4"},
		Supervisors:  []uint{},
		Teams:        []uint{},
	})
	helpers.AssertSuccess(t, createEnv, "POST /tp/")

	var tp model.TP_Object
	helpers.MustDecode(t, createEnv, &tp)

	// Find the matching objects row id (auto-incremented separately).
	var obj model.Object
	if err := helpers.DB().
		Where("type = ? AND object_detailed_id = ?", "tp_objects", tp.ID).
		First(&obj).Error; err != nil {
		t.Fatalf("locate object row: %v", err)
	}

	// GET /object/:id with the OBJECT id (not the tp_object id).
	getEnv := helpers.AuthedJSON(t, "GET", "/object/"+strconv.FormatUint(uint64(obj.ID), 10), token, nil)
	helpers.AssertSuccess(t, getEnv, "GET /object/:id")

	var got model.Object
	helpers.MustDecode(t, getEnv, &got)

	if got.ID != obj.ID {
		t.Fatalf("returned id %d, want %d", got.ID, obj.ID)
	}
	if got.Type != "tp_objects" {
		t.Fatalf("returned type %q, want tp_objects", got.Type)
	}
	if got.ObjectDetailedID != tp.ID {
		t.Fatalf("returned objectDetailedID %d, want %d", got.ObjectDetailedID, tp.ID)
	}

	// Lock-in: response is the polymorphism shell only — no specialized fields.
	helpers.AssertJSONGolden(t, "object_polymorphism/get_by_id_tp", getEnv.Data)
}

// TestObject_GetByID_NotFound locks the apperr.NotFound contract on the pilot
// endpoint: GET /object/<unused id> returns the standard envelope with the
// generic Russian "Запись не найдена" message — no GORM internals leaked.
func TestObject_GetByID_NotFound(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	env := helpers.AuthedJSON(t, "GET", "/object/99999", token, nil)
	msg := helpers.AssertFailure(t, env, "GET /object/99999")
	if msg != "Запись не найдена" {
		t.Fatalf("error = %q, want %q", msg, "Запись не найдена")
	}
	if !env.Permission {
		t.Fatalf("permission flag should remain true on a not-found error, got false")
	}
}

// TestObject_GetByID_BadID locks the apperr.InvalidInput contract: a
// non-numeric :id parameter returns "Некорректный идентификатор" with no
// strconv error text leaked.
func TestObject_GetByID_BadID(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	env := helpers.AuthedJSON(t, "GET", "/object/abc", token, nil)
	msg := helpers.AssertFailure(t, env, "GET /object/abc")
	if msg != "Некорректный идентификатор" {
		t.Fatalf("error = %q, want %q", msg, "Некорректный идентификатор")
	}
	if !env.Permission {
		t.Fatalf("permission flag should remain true on a parse error, got false")
	}
}
