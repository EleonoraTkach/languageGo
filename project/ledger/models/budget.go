package models

import (
	"errors"
	"time"
)

type Budget struct {
	Category string    `json:"category"`
	Limit    float64   `json:"limit"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
}

func (b Budget) Validate() error {
	if b.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}

	if b.Category == "" {
		return errors.New("category must not be empty")
	}

	if b.From.IsZero() {
		return errors.New("from date is required")
	}

	if b.To.IsZero() {
		return errors.New("to date is required")
	}

	return nil
}
