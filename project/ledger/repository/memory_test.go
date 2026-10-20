package repository

import (
	"testing"
	"time"

	m "ledger/models"
)

func TestAddTransactionBudgetExceeded(t *testing.T) {
	storage := NewStorage()
	storage.Reset()

	budget := m.Budget{
		Category: "еда",
		Limit:    5000,
		From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC),
	}

	if err := storage.SetBudget(budget); err != nil {
		t.Fatalf("SetBudget() error = %v", err)
	}

	first := m.Transaction{
		Amount:      3000,
		Category:    "еда",
		Description: "продукты",
		Date:        time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC),
	}

	if err := storage.AddTransaction(first); err != nil {
		t.Fatalf("first AddTransaction() error = %v", err)
	}

	second := m.Transaction{
		Amount:      2500,
		Category:    "еда",
		Description: "ресторан",
		Date:        time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC),
	}

	if err := storage.AddTransaction(second); err == nil {
		t.Fatal("expected budget exceeded error, got nil")
	}

	transactions := storage.ListTransactions()

	if got, want := len(transactions), 1; got != want {
		t.Fatalf("ListTransactions() length = %d, want %d", got, want)
	}

	if transactions[0].Amount != 3000 {
		t.Errorf("saved transaction amount = %v, want 3000", transactions[0].Amount)
	}
}
