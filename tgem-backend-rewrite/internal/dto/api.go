package dto

type APIRequestFormat[T any] struct {
	RequestURL string `json:"requestURL"`
	Data       T      `json:"data"`
}
