package dto

import "backend-v2/model"

type MaterialDataAndCost struct {
	Details model.Material     `json:"details"`
	Cost    model.MaterialCost `json:"cost"`
}
