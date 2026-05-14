package model

import "time"

type UserAction struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	ActionURL           string    `json:"actionURL"`
	ActionType          string    `json:"actionType"`
	ActionID            uint      `json:"actionID"`
	ActionStatus        bool      `json:"actionStatus"`
	ActionStatusMessage string    `json:"actionStatusMessage"`
	HTTPMethod          string    `json:"httpMethod" gorm:"column:http_method"`
	RequestIP           string    `json:"requestIP" gorm:"column:request_ip"`
	UserID              uint      `json:"userID"`
	ProjectID           uint      `json:"projectID"`
	DateOfAction        time.Time `json:"dateOfAction"`
}
