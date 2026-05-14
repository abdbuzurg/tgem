package dto

import "time"

type UserActionView struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	ActionURL           string    `json:"actionURL"`
	ActionType          string    `json:"actionType"`
	ActionID            uint      `json:"actionID"`
	ActionStatus        bool      `json:"actionStatus"`
	ActionStatusMessage string    `json:"actionStatusMessage"`
	HTTPMethod          string    `json:"httpMethod"`
	RequestIP           string    `json:"requestIP"`
	UserID              uint      `json:"userID"`
	ProjectID           uint      `json:"projectID"`
	Username            string    `json:"username"`
	DateOfAction        time.Time `json:"dateOfAction"`
}

// UserActionFilter holds optional filter values parsed from query parameters
// by the GetPaginated handler. A zero value means "no filter on this field".
// DateFrom / DateTo are inclusive bounds on date_of_action.
type UserActionFilter struct {
	Page       int
	Limit      int
	UserID     uint
	ProjectID  uint
	ActionType string
	HTTPMethod string
	HasStatus  bool
	Status     bool
	DateFrom   time.Time
	DateTo     time.Time
}

// UserActionFilterUserOption is one row in the dropdown that powers the
// audit-log page's "filter by user" search box. The frontend matches the
// admin's typed text against username OR worker_name.
type UserActionFilterUserOption struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	WorkerName string `json:"workerName"`
}
