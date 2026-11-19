package service

import (
	"context"

	d "ledger/internal/domain"
)

type Service interface {
	SetBudget(ctx context.Context, b d.Budget) error
	GetBudget(ctx context.Context, category string) (d.Budget, error)
	ListBudgets(ctx context.Context) ([]d.Budget, error)

	AddTransaction(ctx context.Context, tx d.Transaction) (d.Transaction, error)
	ListTransactions(ctx context.Context) ([]d.Transaction, error)
}
