package dto

type PieChartData struct {
	ID    uint   `json:"id"`
	Value float64   `json:"value"`
	Label string `json:"label"`
}

type InvoiceMaterialStats struct {
  Amount float64
  InvoiceType string
}

type LocationMaterialStats struct {
  Amount float64
  LocationType string
}
