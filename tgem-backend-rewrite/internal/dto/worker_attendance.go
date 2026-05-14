package dto

import "time"

type WorkerAttendancePaginated struct {
	ID              uint      `json:"id"`
	WorkerName      string    `json:"workerName"`
	CompanyWorkerID string    `json:"companyWorkerID"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
}
