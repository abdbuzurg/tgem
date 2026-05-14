package dto

import "backend-v2/model"

type TPObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	Model            string `json:"model"`
	VoltageClass     string `json:"voltageClass"`
}

type TPObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	Model            string   `json:"model"`
	VoltageClass     string   `json:"voltageClass"`
	Supervisors      []string `json:"supervisors"`
	Teams            []string `json:"teams"`
}

type TPObjectCreate struct {
	BaseInfo     model.Object    `json:"baseInfo"`
	DetailedInfo model.TP_Object `json:"detailedInfo"`
	Supervisors  []uint          `json:"supervisors"`
	Teams        []uint          `json:"teams"`
}

type TPObjectSearchParameters struct {
	ProjectID          uint
	ObjectName         string
	SupervisorWorkerID uint
	TeamID             uint
}

type TPObjectImportData struct {
	Object            model.Object
	TP                model.TP_Object
	ObjectSupervisors model.ObjectSupervisors
	ObjectTeam        model.ObjectTeams
	NourashedByTP     model.TPNourashesObjects
}
