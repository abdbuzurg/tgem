package utils

import "backend-v2/model"

func RequestedMaterials(materials []model.MaterialLocation, requestedAmount float64) []model.MaterialLocation {
	result := []model.MaterialLocation{}
	index := 0
	for requestedAmount > 0 {
		if materials[index].Amount <= requestedAmount {
			requestedAmount -= materials[index].Amount
			result = append(result, materials[index])
		} else {
			result = append(result, model.MaterialLocation{
				MaterialCostID: materials[index].MaterialCostID,
				ID:             0,
				LocationID:     0,
				LocationType:   "warehouse",
				Amount:         requestedAmount,
			})
			materials[index].Amount -= requestedAmount
			requestedAmount = 0
		}

		index++
	}

	return result
}
