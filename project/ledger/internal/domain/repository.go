package domain

import (
	"context"
	"time"
)

type BudgetRepository interface {
	SetBudget(ctx context.Context, budget Budget) error
	GetBudgetByCategory(ctx context.Context, category string) (Budget, error)
	ListBudgets(ctx context.Context) ([]Budget, error)
}

type TransactionRepository interface {
	AddTransaction(ctx context.Context, tx Transaction) (Transaction, error)
	ListTransactions(ctx context.Context) ([]Transaction, error)
	GetSpentByCategory(ctx context.Context, category string) (float64, error)
	GetCategories(ctx context.Context) ([]string, error)
	GetSpentByCategoryAndPeriod(
		ctx context.Context,
		category string,
		from time.Time,
		to time.Time,
	) (float64, error)
}
