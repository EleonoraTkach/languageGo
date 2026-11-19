package domain

import (
	"errors"
)

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (b Budget) Validate() error {
	if b.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}

	if b.Category == "" {
		return errors.New("category must not be empty")
	}

	return nil
}
