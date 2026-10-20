package api

import (
	"time"

	m "ledger/models"
)

type CreateTransactionRequest struct {
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type TransactionResponse struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (r CreateTransactionRequest) ToTransaction() m.Transaction {
	return m.Transaction{
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        r.Date,
	}
}

func NewTransactionResponse(t m.Transaction) TransactionResponse {
	return TransactionResponse{
		Amount:      t.Amount,
		Category:    t.Category,
		Description: t.Description,
		Date:        t.Date,
	}
}
