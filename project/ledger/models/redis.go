package models

type CategorySummary struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type ReportSummary struct {
	From       string            `json:"from"`
	To         string            `json:"to"`
	Total      float64           `json:"total"`
	ByCategory []CategorySummary `json:"by_category"`
}
