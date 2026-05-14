package model

type ObjectSupervisors struct {
	ID                 uint `json:"id" gorm:"primaryKey"`
	SupervisorWorkerID uint `json:"supervisorWorkerID"`
	ObjectID           uint `json:"objectID"`
}
