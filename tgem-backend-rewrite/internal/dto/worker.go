package dto

type WorkerInformationForSearch struct {
	Name              []string `json:"name" gorm:"type:text[] serializer:json"`
	JobTitleInCompany []string `json:"jobTitleInCompany" gorm:"type:text[] serializer:json"`
	CompanyWorkerID   []string `json:"companyWorkerID" gorm:"type:text[] serializer:json"`
	JobTitleInProject []string `json:"jobTitleInProject" gorm:"type:text[] serializer:json"`
	MobileNumber      []string `json:"mobileNumber" gorm:"type:text[] serializer:json"`
}

type WorkerSearchParameters struct {
	ProjectID         uint
	Name              string
	JobTitleInCompany string
	JobTitleInProject string
	CompanyWorkerID   string
	MobileNumber      string
}
