package helpers

import (
	"backend-v2/model"
	"testing"

	"github.com/shopspring/decimal"
)

// Material inserts a model.Material under the given projectID.
func Material(t *testing.T, projectID uint, name, code, unit string) model.Material {
	t.Helper()
	m := model.Material{
		ProjectID: projectID,
		Name:      name,
		Code:      code,
		Unit:      unit,
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("create material: %v", err)
	}
	return m
}

// MaterialCost adds a cost variant for the given material.
func MaterialCost(t *testing.T, materialID uint, prime, m19 float64) model.MaterialCost {
	t.Helper()
	c := model.MaterialCost{
		MaterialID:       materialID,
		CostPrime:        decimal.NewFromFloat(prime),
		CostM19:          decimal.NewFromFloat(m19),
		CostWithCustomer: decimal.NewFromFloat(m19),
	}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("create material_cost: %v", err)
	}
	return c
}

// Worker inserts a model.Worker. JobTitleInProject and Name are required by
// downstream business logic; the rest are optional but populated to keep
// goldens stable.
func Worker(t *testing.T, projectID uint, name, jobInProject string) model.Worker {
	t.Helper()
	w := model.Worker{
		ProjectID:         projectID,
		Name:              name,
		JobTitleInProject: jobInProject,
		JobTitleInCompany: jobInProject,
		MobileNumber:      "+9920000001",
	}
	if err := db.Create(&w).Error; err != nil {
		t.Fatalf("create worker: %v", err)
	}
	return w
}

// Team inserts a model.Team. leaderIDs are linked via team_leaders.
func Team(t *testing.T, projectID uint, number, mobile, company string, leaderIDs []uint) model.Team {
	t.Helper()
	team := model.Team{
		ProjectID:    projectID,
		Number:       number,
		MobileNumber: mobile,
		Company:      company,
	}
	if err := db.Create(&team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	for _, lid := range leaderIDs {
		link := model.TeamLeaders{TeamID: team.ID, LeaderWorkerID: lid}
		if err := db.Create(&link).Error; err != nil {
			t.Fatalf("create team_leader: %v", err)
		}
	}
	return team
}

// District inserts a model.District.
func District(t *testing.T, projectID uint, name string) model.District {
	t.Helper()
	d := model.District{ProjectID: projectID, Name: name}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("create district: %v", err)
	}
	return d
}

// WarehouseStock inserts a material_locations row at (LocationType=warehouse,
// LocationID=0). Used to skip an input flow when only stock is needed.
func WarehouseStock(t *testing.T, projectID, materialCostID uint, amount float64) model.MaterialLocation {
	t.Helper()
	loc := model.MaterialLocation{
		ProjectID:      projectID,
		MaterialCostID: materialCostID,
		LocationID:     0,
		LocationType:   "warehouse",
		Amount:         amount,
	}
	if err := db.Create(&loc).Error; err != nil {
		t.Fatalf("create warehouse stock: %v", err)
	}
	return loc
}

// TeamStock inserts a material_locations row at (LocationType=team,
// LocationID=teamID).
func TeamStock(t *testing.T, projectID, teamID, materialCostID uint, amount float64) model.MaterialLocation {
	t.Helper()
	loc := model.MaterialLocation{
		ProjectID:      projectID,
		MaterialCostID: materialCostID,
		LocationID:     teamID,
		LocationType:   "team",
		Amount:         amount,
	}
	if err := db.Create(&loc).Error; err != nil {
		t.Fatalf("create team stock: %v", err)
	}
	return loc
}
