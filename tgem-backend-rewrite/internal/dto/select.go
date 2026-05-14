package dto

type DataForSelect[T uint | string] struct {
	Label string `json:"label"`
	Value T      `json:"value"`
}
