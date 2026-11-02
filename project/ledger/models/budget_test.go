package models

import (
	"testing"
	/*"time"*/)

func TestBudgetValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
	}{
		{
			name: "valid budget",
			budget: Budget{
				Category: "еда",
				Limit:    5000,
				/*From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
				To:       time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC),*/
			},
			wantErr: false,
		},
		{
			name: "zero limit",
			budget: Budget{
				Category: "еда",
				Limit:    0,
				/*From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
				To:       time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),*/
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			budget: Budget{
				Category: "еда",
				Limit:    -1,
				/*From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
				To:       time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),*/
			},
			wantErr: true,
		},
		{
			name: "empty category",
			budget: Budget{
				Limit: 5000,
				/*From:  time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
				To:    time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),*/
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.budget.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
