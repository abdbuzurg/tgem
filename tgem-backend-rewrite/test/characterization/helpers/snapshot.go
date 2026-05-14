package helpers

import (
	"backend-v2/model"
	"fmt"
	"sort"
	"testing"
)

// LocSnapshot keys MaterialLocation rows by "<type>:<locID>:<matCostID>" so
// tests can compare before/after states with a simple map diff.
type LocSnapshot map[string]float64

// MaterialLocationSnapshot loads every material_locations row scoped to the
// given project and returns the map.
func MaterialLocationSnapshot(t *testing.T, projectID uint) LocSnapshot {
	t.Helper()
	var rows []model.MaterialLocation
	if err := db.Where("project_id = ?", projectID).Find(&rows).Error; err != nil {
		t.Fatalf("read material_locations: %v", err)
	}
	out := LocSnapshot{}
	for _, r := range rows {
		out[locKey(r)] = r.Amount
	}
	return out
}

func locKey(r model.MaterialLocation) string {
	return fmt.Sprintf("%s:%d:%d", r.LocationType, r.LocationID, r.MaterialCostID)
}

// LocKey builds a snapshot key from raw fields. Useful for asserting expected
// changes without a full database round trip.
func LocKey(locationType string, locationID, materialCostID uint) string {
	return fmt.Sprintf("%s:%d:%d", locationType, locationID, materialCostID)
}

// Diff returns the set of keys that changed amount between two snapshots, with
// the *new* (or zero, if removed) value. Keys present in only one side appear
// with the value from the side they're on (or 0 for deletions).
func (before LocSnapshot) Diff(after LocSnapshot) map[string]float64 {
	changed := map[string]float64{}
	for k, v := range after {
		if before[k] != v {
			changed[k] = v
		}
	}
	for k := range before {
		if _, ok := after[k]; !ok {
			changed[k] = 0
		}
	}
	return changed
}

// Keys returns sorted keys for stable iteration in test failure messages.
func (s LocSnapshot) Keys() []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
