package characterization_test

import (
	"backend-v2/model"
	"backend-v2/test/characterization/helpers"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestMaterial_ImportExportRoundTrip exercises the full Excel pipeline for the
// simplest reference table (8 columns, no FK preconditions): download the
// template, fill 2 rows, import, query the DB, then export and confirm both
// rows appear in the export. Locks in column ordering, sheet name, and the
// "Да"/"Нет" boolean spelling.
func TestMaterial_ImportExportRoundTrip(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	token := helpers.LoginAsTester(t)

	// 1. Pull the template (binary download; not an envelope).
	tmplBytes, _ := helpers.Download(t, "/material/document/template", token)
	if len(tmplBytes) < 100 {
		t.Fatalf("template download too small (%d bytes); maybe the endpoint returned a JSON error", len(tmplBytes))
	}

	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "import.xlsx")
	if err := os.WriteFile(tmplPath, tmplBytes, 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	// 2. Fill two rows.
	f, err := excelize.OpenFile(tmplPath)
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	const sheet = "Материалы"

	type row struct {
		name, code, category, unit, article, hasSerial, showInReport, plannedAmount, notes string
	}
	rows := []row{
		{"Char-Cable-1", "CHAR-001", "Кабель", "м", "ART-1", "Нет", "Да", "100", "row 1"},
		{"Char-Cable-2", "CHAR-002", "Кабель", "м", "ART-2", "Да", "Нет", "50", "row 2"},
	}
	for i, r := range rows {
		excelRow := i + 2 // header is row 1
		set := func(col, val string) {
			if err := f.SetCellStr(sheet, col+strconv.Itoa(excelRow), val); err != nil {
				t.Fatalf("set %s%d: %v", col, excelRow, err)
			}
		}
		set("A", r.name)
		set("B", r.code)
		set("C", r.category)
		set("D", r.unit)
		set("E", r.article)
		set("F", r.hasSerial)
		set("G", r.showInReport)
		set("H", r.plannedAmount)
		set("I", r.notes)
	}
	if err := f.SaveAs(tmplPath); err != nil {
		t.Fatalf("save filled template: %v", err)
	}
	_ = f.Close()

	// 3. Import via multipart (form field name is "file").
	importEnv := helpers.MultipartUpload(t, "/material/document/import", token, "file", tmplPath, nil)
	helpers.AssertSuccess(t, importEnv, "POST /material/document/import")

	// 4. DB shows both rows under projectID=1.
	var dbRows []model.Material
	if err := helpers.DB().
		Where("project_id = ? AND name IN ?", uint(1), []string{"Char-Cable-1", "Char-Cable-2"}).
		Order("name").Find(&dbRows).Error; err != nil {
		t.Fatalf("query materials: %v", err)
	}
	if len(dbRows) != 2 {
		t.Fatalf("expected 2 imported materials, got %d", len(dbRows))
	}
	if dbRows[0].HasSerialNumber {
		t.Fatalf("Char-Cable-1 should have HasSerialNumber=false (Нет), got true")
	}
	if !dbRows[1].HasSerialNumber {
		t.Fatalf("Char-Cable-2 should have HasSerialNumber=true (Да), got false")
	}
	if dbRows[0].PlannedAmountForProject != 100 {
		t.Fatalf("Char-Cable-1 PlannedAmountForProject=%v, want 100", dbRows[0].PlannedAmountForProject)
	}

	// 5. Export and verify both names appear in column A.
	exportBytes, _ := helpers.Download(t, "/material/document/export", token)
	exportPath := filepath.Join(dir, "export.xlsx")
	if err := os.WriteFile(exportPath, exportBytes, 0o644); err != nil {
		t.Fatalf("write export: %v", err)
	}
	ef, err := excelize.OpenFile(exportPath)
	if err != nil {
		t.Fatalf("open export: %v", err)
	}
	defer ef.Close()

	exportedRows, err := ef.GetRows(sheet)
	if err != nil {
		t.Fatalf("read export rows: %v", err)
	}

	seen := map[string]bool{}
	for i, row := range exportedRows {
		if i == 0 {
			continue
		}
		if len(row) > 0 {
			seen[row[0]] = true
		}
	}
	for _, want := range []string{"Char-Cable-1", "Char-Cable-2"} {
		if !seen[want] {
			t.Errorf("expected %q in export column A; got rows %v", want, exportedRows)
		}
	}
}

