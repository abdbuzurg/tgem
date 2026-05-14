package model

type SerialNumberLocation struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	SerialNumberID uint   `json:"serialNumberID"`
	ProjectID      uint   `json:"projectID"`
	LocationID     uint   `json:"locationID"`
	LocationType   string `json:"locationType"`
}
