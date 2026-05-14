package dto

import "backend-v2/model"

type STVTObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Type             string `json:"type" gorm:"tinyText"`
	Name             string `json:"name" gorm:"tinyText"`
	Status           string `json:"status" gorm:"tinyText"`
	VoltageClass     string `json:"voltageClass" gorm:"tinyText"`
	TTCoefficient    string `json:"ttCoefficient" gorm:"tinyText"`
}

type STVTObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Type             string   `json:"type" gorm:"tinyText"`
	Name             string   `json:"name" gorm:"tinyText"`
	Status           string   `json:"status" gorm:"tinyText"`
	VoltageClass     string   `json:"voltageClass" gorm:"tinyText"`
	TTCoefficient    string   `json:"ttCoefficient" gorm:"tinyText"`
	Supervisors      []string `json:"supervisors"`
	Teams            []string `json:"teams"`
}

type STVTObjectCreate struct {
	BaseInfo     model.Object      `json:"baseInfo"`
	DetailedInfo model.STVT_Object `json:"detailedInfo"`
	Supervisors  []uint            `json:"supervisors"`
	Teams        []uint            `json:"teams"`
}

type STVTObjectSearchParameters struct {
	ProjectID          uint
	ObjectName         string
	SupervisorWorkerID uint
	TeamID             uint
}

type STVTObjectImportData struct {
	Object            model.Object
	STVT               model.STVT_Object
	ObjectSupervisors model.ObjectSupervisors
	ObjectTeam        model.ObjectTeams
}
