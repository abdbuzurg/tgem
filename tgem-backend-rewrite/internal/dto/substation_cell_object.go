package dto

import "backend-v2/model"

type SubstationCellObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Type             string `json:"type" gorm:"tinyText"`
	Name             string `json:"name" gorm:"tinyText"`
	Status           string `json:"status" gorm:"tinyText"`
}

type SubstationCellObjectSearchParameters struct {
	ProjectID          uint
	ObjectName         string
	SupervisorWorkerID uint
	TeamID             uint
  SubstationObjectID uint
}

type SubstationCellObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Type             string   `json:"type" gorm:"tinyText"`
	Name             string   `json:"name" gorm:"tinyText"`
	Status           string   `json:"status" gorm:"tinyText"`
	Supervisors      []string `json:"supervisors"`
	Teams            []string `json:"teams"`
	SubstationName   string   `json:"substationName"`
}

type SubstationCellObjectCreate struct {
	BaseInfo           model.Object               `json:"baseInfo"`
	DetailedInfo       model.SubstationCellObject `json:"detailedInfo"`
	Supervisors        []uint                     `json:"supervisors"`
	Teams              []uint                     `json:"teams"`
	SubstationObjectID uint                       `json:"substationObjectID"`
}

type SubstationCellObjectImportData struct {
	Object            model.Object
	SubstationCell    model.SubstationCellObject
	ObjectSupervisors model.ObjectSupervisors
	ObjectTeam        model.ObjectTeams
	Nourashes         model.SubstationCellNourashesSubstationObject
}
