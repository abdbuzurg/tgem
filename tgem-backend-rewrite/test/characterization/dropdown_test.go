package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"testing"
	"time"
)

// TestInvoiceInput_UniqueCode_ReturnsServerGeneratedDeliveryCodes locks in the
// dropdown contract: the endpoint returns []DataForSelect[string] ordered by
// invoice id descending, with both label and value set to deliveryCode.
func TestInvoiceInput_UniqueCode_ReturnsServerGeneratedDeliveryCodes(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// invoice_inputs has FKs on warehouse_manager_worker_id / released_worker_id
	// → workers. Reuse the seeded tester worker for both.
	var tester model.Worker
	if err := helpers.DB().Where("name = ?", "Test User").First(&tester).Error; err != nil {
		t.Fatalf("locate tester worker: %v", err)
	}

	// Seed two InvoiceInput rows directly. We bypass the full Create flow
	// because the dropdown test cares about list shape, not how rows arrive.
	rows := []model.InvoiceInput{
		{ProjectID: 1, WarehouseManagerWorkerID: tester.ID, ReleasedWorkerID: tester.ID, DeliveryCode: "П-01-00001", DateOfInvoice: time.Now()},
		{ProjectID: 1, WarehouseManagerWorkerID: tester.ID, ReleasedWorkerID: tester.ID, DeliveryCode: "П-01-00002", DateOfInvoice: time.Now()},
	}
	if err := helpers.DB().Create(&rows).Error; err != nil {
		t.Fatalf("seed invoice_inputs: %v", err)
	}

	env := helpers.AuthedJSON(t, "GET", "/input/unique/code", token, nil)
	helpers.AssertSuccess(t, env, "GET /input/unique/code")

	var got []dto.DataForSelect[string]
	helpers.MustDecode(t, env, &got)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d (%+v)", len(got), got)
	}
	// Order = id DESC, so the second-inserted code shows up first.
	if got[0].Value != "П-01-00002" || got[0].Label != "П-01-00002" {
		t.Errorf("got[0] = %+v, want both fields = П-01-00002", got[0])
	}
	if got[1].Value != "П-01-00001" || got[1].Label != "П-01-00001" {
		t.Errorf("got[1] = %+v, want both fields = П-01-00001", got[1])
	}

	helpers.AssertJSONGolden(t, "dropdown/invoice_input_unique_code", env.Data)
}
