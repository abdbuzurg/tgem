package dto

import "backend-v2/model"

type UserPaginated struct {
	ID                 uint     `json:"id"`
	Username           string   `json:"username"`
	WorkerName         string   `json:"workerName"`
	WorkerMobileNumber string   `json:"workerMobileNumber"`
	WorkerJobTitle     string   `json:"workerJobTitle"`
	RoleName           string   `json:"roleName"`
	AccessToProjects   []string `json:"accessToProjects"`
}

type NewUserData struct {
	UserData model.User `json:"userData"`
	Projects []uint     `json:"projects"`
}
