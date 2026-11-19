package service

import (
	"context"
	"errors"
	"fmt"

	d "ledger/internal/domain"
)

type LedgerService struct {
	budgets      d.BudgetRepository
	transactions d.TransactionRepository
}

func NewLedgerService(
	budgets d.BudgetRepository,
	transactions d.TransactionRepository,
) *LedgerService {
	return &LedgerService{
		budgets:      budgets,
		transactions: transactions,
	}
}

func (s *LedgerService) AddTransaction(
	ctx context.Context,
	tx d.Transaction,
) (d.Transaction, error) {
	if err := tx.Validate(); err != nil {
		return d.Transaction{}, err
	}

	budget, err := s.budgets.GetBudgetByCategory(ctx, tx.Category)
	if err != nil {
		if errors.Is(err, d.ErrBudgetNotFound) {
			return s.transactions.AddTransaction(ctx, tx)
		}

		return d.Transaction{}, fmt.Errorf("failed to get budget: %w", err)
	}

	spent, err := s.transactions.GetSpentByCategory(ctx, tx.Category)
	if err != nil {
		return d.Transaction{}, fmt.Errorf("failed to calculate expenses: %w", err)
	}

	if spent+tx.Amount > budget.Limit {
		return d.Transaction{}, d.ErrBudgetExceeded
	}

	return s.transactions.AddTransaction(ctx, tx)
}

func (s *LedgerService) SetBudget(
	ctx context.Context,
	b d.Budget,
) error {
	if err := b.Validate(); err != nil {
		return err
	}

	if err := s.budgets.SetBudget(ctx, b); err != nil {
		return fmt.Errorf("failed to upsert budget: %w", err)
	}

	return nil
}

func (s *LedgerService) GetBudget(
	ctx context.Context,
	category string,
) (d.Budget, error) {
	return s.budgets.GetBudgetByCategory(ctx, category)
}

func (s *LedgerService) ListBudgets(
	ctx context.Context,
) ([]d.Budget, error) {
	return s.budgets.ListBudgets(ctx)
}

func (s *LedgerService) ListTransactions(
	ctx context.Context,
) ([]d.Transaction, error) {
	return s.transactions.ListTransactions(ctx)
}
