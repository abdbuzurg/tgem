package dto

import "backend-v2/model"

type SubstationObjectPaginatedQuery struct {
	ObjectID             uint   `json:"objectID"`
	ObjectDetailedID     uint   `json:"objectDetailedID"`
	Name                 string `json:"name"`
	Status               string `json:"status"`
	VoltageClass         string `json:"voltageClass"`
	NumberOfTransformers string `json:"numberOfTransformers"`
}

type SubstationObjectPaginated struct {
	ObjectID             uint     `json:"objectID"`
	ObjectDetailedID     uint     `json:"objectDetailedID"`
	Name                 string   `json:"name"`
	Status               string   `json:"status"`
	VoltageClass         string   `json:"voltageClass"`
	NumberOfTransformers string   `json:"numberOfTransformers"`
	Supervisors          []string `json:"supervisors"`
	Teams                []string `json:"teams"`
}

type SubstationObjectCreate struct {
	BaseInfo     model.Object            `json:"baseInfo"`
	DetailedInfo model.Substation_Object `json:"detailedInfo"`
	Supervisors  []uint                  `json:"supervisors"`
	Teams        []uint                  `json:"teams"`
}

type SubstationObjectImportData struct {
	Object            model.Object
	Substation        model.Substation_Object
	ObjectSupervisors model.ObjectSupervisors
	ObjectTeam        model.ObjectTeams
}

type SubstationObjectSearchParameters struct {
	ProjectID          uint
	ObjectName         string
	SupervisorWorkerID uint
	TeamID             uint
}
