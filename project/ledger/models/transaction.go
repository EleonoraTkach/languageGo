package models

import (
	"errors"
	"time"
)

type Transaction struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if t.Category == "" {
		return errors.New("category must not be empty")
	}

	if t.Date.IsZero() {
		return errors.New("date must not be empty")
	}

	return nil
}
