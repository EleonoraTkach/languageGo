package models

import (
	"testing"
	"time"
)

func TestTransactionValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{
			name: "valid transaction",
			tx: Transaction{
				Amount:      450,
				Category:    "еда",
				Description: "ланч",
				Date:        time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "zero amount",
			tx: Transaction{
				Amount:   0,
				Category: "еда",
				Date:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			tx: Transaction{
				Amount:   -100,
				Category: "еда",
				Date:     time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty category",
			tx: Transaction{
				Amount: 100,
				Date:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
