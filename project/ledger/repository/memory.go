package repository

import (
	"errors"
	m "ledger/models"
	"time"
)

type Storage struct {
	transactions []m.Transaction
}

func NewStorage() *Storage {
	return &Storage{
		transactions: make([]m.Transaction, 0),
	}
}

func (s *Storage) Add(amount float64, category, description string) (m.Transaction, error) {

	if amount <= 0 {
		return m.Transaction{}, errors.New("сумма транзакции должна быть больше нуля")
	}

	newTx := m.Transaction{
		ID:          len(s.transactions) + 1,
		Amount:      amount,
		Category:    category,
		Description: description,
		Date:        time.Now(),
	}

	s.transactions = append(s.transactions, newTx)

	return newTx, nil
}

func (s *Storage) ListTransactions() []m.Transaction {
	copiedTransactions := make([]m.Transaction, len(s.transactions))
	copy(copiedTransactions, s.transactions)

	return copiedTransactions
}
