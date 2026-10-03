package models

import (
	"time"
)

type Budget struct {
	Category string    `json:"category"`
	Limit    float64   `json:"limit"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
}
