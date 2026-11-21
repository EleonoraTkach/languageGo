package ledger

import (
	"context"
	"time"
)

type Transaction struct {
	ID          int64
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

type Budget struct {
	Category string
	Limit    float64
}

type Client interface {
	Ping(ctx context.Context) error
	AddTransaction(ctx context.Context, tx Transaction) (Transaction, error)
	ListTransactions(ctx context.Context) ([]Transaction, error)
	SetBudget(ctx context.Context, budget Budget) (Budget, error)
	ListBudgets(ctx context.Context) ([]Budget, error)

	GetReportSummary(
		ctx context.Context,
		from time.Time,
		to time.Time,
	) (map[string]float64, error)
}
