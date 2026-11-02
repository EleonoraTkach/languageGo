package api

import (
	/*"time"*/

	m "ledger/models"
)

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	/*From     time.Time `json:"from"`
	To       time.Time `json:"to"`*/
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	/*From     time.Time `json:"from"`
	To       time.Time `json:"to"`*/
}

func (r CreateBudgetRequest) ToBudget() m.Budget {
	return m.Budget{
		Category: r.Category,
		Limit:    r.Limit,
		/*From:     r.From,
		To:       r.To,*/
	}
}

func NewBudgetResponse(b m.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
		/*From:     b.From,
		To:       b.To,*/
	}
}
