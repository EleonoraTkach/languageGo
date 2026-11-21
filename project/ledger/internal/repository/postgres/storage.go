package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	d "ledger/internal/domain"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(database *sql.DB) *Storage {
	return &Storage{
		db: database,
	}
}

/* Бюджеты */

func (s *Storage) SetBudget(
	ctx context.Context,
	b d.Budget,
) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO budgets(category, limit_amount)
		 VALUES($1, $2)
		 ON CONFLICT(category)
		 DO UPDATE SET limit_amount = EXCLUDED.limit_amount`,
		b.Category,
		b.Limit,
	)

	if err != nil {
		return fmt.Errorf("failed to set budget: %w", err)
	}

	return nil
}

func (s *Storage) GetBudgetByCategory(
	ctx context.Context,
	category string,
) (d.Budget, error) {
	var b d.Budget

	err := s.db.QueryRowContext(
		ctx,
		`SELECT category, limit_amount
		 FROM budgets
		 WHERE category = $1`,
		category,
	).Scan(&b.Category, &b.Limit)

	if errors.Is(err, sql.ErrNoRows) {
		return d.Budget{}, d.ErrBudgetNotFound
	}

	if err != nil {
		return d.Budget{}, fmt.Errorf("failed to get budget: %w", err)
	}

	return b, nil
}

func (s *Storage) ListBudgets(
	ctx context.Context,
) ([]d.Budget, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT category, limit_amount
		 FROM budgets
		 ORDER BY category`,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []d.Budget

	for rows.Next() {
		var b d.Budget

		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}

		budgets = append(budgets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate budgets: %w", err)
	}

	return budgets, nil
}

/* Транзакции */

func (s *Storage) AddTransaction(
	ctx context.Context,
	tx d.Transaction,
) (d.Transaction, error) {
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO expenses(amount, category, description, date)
		 VALUES($1, $2, $3, $4)
		 RETURNING id`,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.Date,
	).Scan(&tx.ID)

	if err != nil {
		return d.Transaction{}, fmt.Errorf("failed to add transaction: %w", err)
	}

	return tx, nil
}

func (s *Storage) ListTransactions(
	ctx context.Context,
) ([]d.Transaction, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, amount, category, description, date
		 FROM expenses
		 ORDER BY date DESC, id DESC`,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []d.Transaction

	for rows.Next() {
		var tx d.Transaction

		if err := rows.Scan(
			&tx.ID,
			&tx.Amount,
			&tx.Category,
			&tx.Description,
			&tx.Date,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate transactions: %w", err)
	}

	return transactions, nil
}

func (s *Storage) GetSpentByCategory(
	ctx context.Context,
	category string,
) (float64, error) {
	var spent float64

	err := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE category = $1`,
		category,
	).Scan(&spent)

	if err != nil {
		return 0, fmt.Errorf("failed to calculate expenses: %w", err)
	}

	return spent, nil
}

func (s *Storage) GetCategories(
	ctx context.Context,
) ([]string, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT DISTINCT category
         FROM expenses
         ORDER BY category`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []string

	for rows.Next() {
		var category string

		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate categories: %w", err)
	}

	return categories, nil
}

func (s *Storage) GetSpentByCategoryAndPeriod(
	ctx context.Context,
	category string,
	from time.Time,
	to time.Time,
) (float64, error) {
	var spent float64

	err := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(amount), 0)
         FROM expenses
         WHERE category = $1
           AND date >= $2
           AND date < $3`,
		category,
		from,
		to.AddDate(0, 0, 1),
	).Scan(&spent)

	if err != nil {
		return 0, fmt.Errorf(
			"failed to calculate expenses for category %q: %w",
			category,
			err,
		)
	}

	return spent, nil
}
