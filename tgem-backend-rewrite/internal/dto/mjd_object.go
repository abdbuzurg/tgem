package dto

import "backend-v2/model"

type MJDObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Name             string `json:"name" gorm:"tinyText"`
	Status           string `json:"status" gorm:"tinyText"`
	Model            string `json:"model" gorm:"tinyText"`
	AmountStores     uint   `json:"amountStores"`
	AmountEntrances  uint   `json:"amountEntrances"`
	HasBasement      bool   `json:"hasBasement"`
}

type MJDObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Name             string   `json:"name" gorm:"tinyText"`
	Status           string   `json:"status" gorm:"tinyText"`
	Model            string   `json:"model" gorm:"tinyText"`
	AmountStores     uint     `json:"amountStores"`
	AmountEntrances  uint     `json:"amountEntrances"`
	HasBasement      bool     `json:"hasBasement"`
	Supervisors      []string `json:"supervisors"`
	Teams            []string `json:"teams"`
	TPNames          []string `json:"tpNames"`
}

type MJDObjectCreate struct {
	BaseInfo              model.Object     `json:"baseInfo"`
	DetailedInfo          model.MJD_Object `json:"detailedInfo"`
	Supervisors           []uint           `json:"supervisors"`
	Teams                 []uint           `json:"teams"`
	NourashedByTPObjectID []uint           `json:"nourashedByTP"`
}

type MJDObjectImportData struct {
	Object            model.Object
	MJD               model.MJD_Object
	ObjectSupervisors model.ObjectSupervisors
	ObjectTeam        model.ObjectTeams
	NourashedByTP     model.TPNourashesObjects
}

type MJDObjectSearchParameters struct {
  ProjectID uint
  ObjectName string 
  SupervisorWorkerID uint
  TeamID uint
  TPObjectID uint
}
