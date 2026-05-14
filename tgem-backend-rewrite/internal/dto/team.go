package dto

type TeamPaginated struct {
	ID           uint     `json:"id" gorm:"primaryKey"`
	Number       string   `json:"number" gorm:"tinyText"`
	MobileNumber string   `json:"mobileNumber" gorm:"tinyText"`
	Company      string   `json:"company" gorm:"tinyText"`
	LeaderNames  []string `json:"leaderNames"`
}

type TeamPaginatedQuery struct {
	ID               uint   `json:"id"`
	LeaderID         uint   `json:"-"`
	LeaderName       string `json:"leaderName"`
	TeamNumber       string `json:"number"`
	TeamMobileNumber string `json:"mobileNumber"`
	TeamCompany      string `json:"company"`
}

type TeamMutation struct {
	ID              uint   `json:"id"`
	Number          string `json:"number"`
	MobileNumber    string `json:"mobileNumber"`
	Company         string `json:"company"`
	LeaderWorkerIDs []uint `json:"leaderIDs"`
	ProjectID       uint
}

type TeamNumberAndTeamLeaderNameQueryResult struct {
	TeamNumber     string
	TeamLeaderName string
}

type TeamSearchParameters struct {
  ProjectID uint
  Number string
  MobileNumber string
  Company string
  TeamLeaderID uint
}
