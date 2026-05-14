package dto

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
  ProjectID uint `json:"projectID"`
}

type LoginResponse struct {
  Token string `json:"token"`
  Admin bool `json:"admin"`
}
