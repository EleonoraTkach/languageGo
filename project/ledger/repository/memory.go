package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	m "ledger/models"
	"time"
)

type Storage struct {
	transactions []m.Transaction
	budgets      map[string][]m.Budget
}

/*
Требование по чеклисту (в задании пункт 2) добавить по умолчанию, в дальшейшем УДАЛИТЬ
*/
func NewStorage() *Storage {
	return &Storage{
		transactions: make([]m.Transaction, 0),
		budgets: map[string][]m.Budget{
			"еда": {
				{
					Category: "еда",
					Limit:    5000,
					From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					To:       time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC),
				},
			},
			"транспорт": {
				{
					Category: "транспорт",
					Limit:    3000,
					From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					To:       time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC),
				},
			},
		},
	}
}

/*
1. Транзакция должна быть больше 0
2. Если для транзакции нет категории и периода бюджета, то добавляем
3. Если для транзакции есть категория и период бюджета, но сумма превышает бюджет, то выбрасываем ошибку
*/
func (s *Storage) AddTransaction(tx m.Transaction) error {
	if tx.Amount <= 0 {
		return errors.New("transaction amount must be greater than zero")
	}

	budgets, exists := s.budgets[tx.Category]

	if !exists {
		s.transactions = append(s.transactions, tx)
		return nil
	}

	for _, budget := range budgets {
		if tx.Date.Before(budget.From) || tx.Date.After(budget.To) {
			continue
		}

		var total float64

		for _, existingTx := range s.transactions {
			if existingTx.Category != tx.Category {
				continue
			}

			if existingTx.Date.Before(budget.From) ||
				existingTx.Date.After(budget.To) {
				continue
			}

			total += existingTx.Amount
		}

		if total+tx.Amount > budget.Limit {
			return errors.New("budget exceeded")
		}
	}

	s.transactions = append(s.transactions, tx)

	return nil
}

func (s *Storage) ListTransactions() []m.Transaction {
	copiedTransactions := make([]m.Transaction, len(s.transactions))
	copy(copiedTransactions, s.transactions)

	return copiedTransactions
}

func (s *Storage) ListBudgets() []m.Budget {
	var budgets []m.Budget

	for _, categoryBudgets := range s.budgets {
		budgets = append(budgets, categoryBudgets...)
	}

	return budgets
}

func (s *Storage) SetBudget(b m.Budget) {
	budgets := s.budgets[b.Category]

	for i, budget := range budgets {
		if budget.From.Equal(b.From) && budget.To.Equal(b.To) {
			budgets[i] = b
			s.budgets[b.Category] = budgets
			return
		}
	}

	s.budgets[b.Category] = append(budgets, b)
}

func (s *Storage) LoadBudgets(r io.Reader) error {
	var budgets []m.Budget

	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("failed to parse budgets: %w", err)
	}

	for _, budget := range budgets {
		s.SetBudget(budget)
	}

	return nil
}
