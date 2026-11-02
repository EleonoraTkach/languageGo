package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"database/sql"
	"ledger/db"
	"ledger/internal/cache"
	m "ledger/models"
	"log"

	"github.com/redis/go-redis/v9"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

type Storage struct{}

func NewStorage() *Storage {
	if err := cache.Init(); err != nil {
		log.Printf("Предупреждение: Redis недоступен: %v", err)
	}

	return &Storage{}
}

func (s *Storage) AddTransaction(tx m.Transaction) error {
	// 1. Валидация транзакции
	if err := tx.Validate(); err != nil {
		return err
	}

	// 2. Получаем бюджет категории
	var limit float64

	err := db.DB.QueryRow(
		`SELECT limit_amount FROM budgets WHERE category = $1`,
		tx.Category,
	).Scan(&limit)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to get budget: %w", err)
	}

	if err == nil {
		// 3. Считаем уже потраченную сумму
		var spent float64

		err = db.DB.QueryRow(
			`SELECT COALESCE(SUM(amount), 0)
			 FROM expenses
			 WHERE category = $1`,
			tx.Category,
		).Scan(&spent)

		if err != nil {
			return fmt.Errorf("failed to calculate expenses: %w", err)
		}

		// 4. Проверяем бюджет
		if spent+tx.Amount > limit {
			return ErrBudgetExceeded
		}
	}

	// 5. Добавляем транзакцию
	var id int

	err = db.DB.QueryRow(
		`INSERT INTO expenses(amount, category, description, date)
		 VALUES($1, $2, $3, $4)
		 RETURNING id`,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.Date,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to add transaction: %w", err)
	}

	return nil
}

func (s *Storage) ListTransactions() ([]m.Transaction, error) {
	rows, err := db.DB.Query(
		`SELECT id, amount, category, description, date
		 FROM expenses
		 ORDER BY date DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []m.Transaction

	for rows.Next() {
		var tx m.Transaction

		err := rows.Scan(
			&tx.ID,
			&tx.Amount,
			&tx.Category,
			&tx.Description,
			&tx.Date,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate transactions: %w", err)
	}

	return transactions, nil
}

func (s *Storage) SetBudget(b m.Budget) error {
	// 1. Валидация бюджета
	if err := b.Validate(); err != nil {
		return err
	}

	// 2. UPSERT бюджета
	_, err := db.DB.Exec(
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

func (s *Storage) ListBudgets() ([]m.Budget, error) {
	rows, err := db.DB.Query(
		`SELECT category, limit_amount
		 FROM budgets
		 ORDER BY category`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}
	defer rows.Close()

	var budgets []m.Budget

	for rows.Next() {
		var b m.Budget

		err := rows.Scan(
			&b.Category,
			&b.Limit,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}

		budgets = append(budgets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate budgets: %w", err)
	}

	return budgets, nil
}

func (s *Storage) LoadBudgets(r io.Reader) error {
	var budgets []m.Budget

	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("failed to parse budgets: %w", err)
	}

	for _, budget := range budgets {
		if err := s.SetBudget(budget); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) Reset() error {
	_, err := db.DB.Exec(`DELETE FROM expenses`)
	if err != nil {
		return fmt.Errorf("failed to reset expenses: %w", err)
	}

	_, err = db.DB.Exec(`DELETE FROM budgets`)
	if err != nil {
		return fmt.Errorf("failed to reset budgets: %w", err)
	}

	return nil
}

func (s *Storage) GetReportSummary(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (m.ReportSummary, error) {

	key := fmt.Sprintf(
		"report:summary:%s:%s",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	// 1. Проверяем кеш Redis
	if cache.Client != nil {
		data, err := cache.Client.Get(ctx, key).Bytes()

		if err == nil {
			var report m.ReportSummary

			if err := json.Unmarshal(data, &report); err == nil {
				log.Println("Redis cache HIT:", key)
				return report, nil
			}

			log.Printf("Ошибка чтения кеша Redis: %v", err)
		} else if !errors.Is(err, redis.Nil) {
			log.Printf("Redis недоступен: %v", err)
		}
	}

	// 2. Если кеша нет, получаем данные из PostgreSQL
	rows, err := db.DB.QueryContext(
		ctx,
		`SELECT category, COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE date >= $1 AND date < $2
		 GROUP BY category
		 ORDER BY category`,
		from,
		to.AddDate(0, 0, 1),
	)

	if err != nil {
		return m.ReportSummary{}, fmt.Errorf(
			"failed to get report: %w", err,
		)
	}
	defer rows.Close()

	// 3. Формируем отчёт
	report := m.ReportSummary{
		From:       from.Format("2006-01-02"),
		To:         to.Format("2006-01-02"),
		ByCategory: make([]m.CategorySummary, 0),
	}

	for rows.Next() {
		var item m.CategorySummary

		if err := rows.Scan(&item.Category, &item.Amount); err != nil {
			return m.ReportSummary{}, err
		}

		report.Total += item.Amount
		report.ByCategory = append(report.ByCategory, item)
	}

	if err := rows.Err(); err != nil {
		return m.ReportSummary{}, err
	}

	if cache.Client != nil {
		data, err := json.Marshal(report)

		if err == nil {
			err = cache.Client.Set(
				ctx,
				key,
				data,
				30*time.Second,
			).Err()

			if err != nil {
				log.Printf("Не удалось сохранить отчёт в Redis: %v", err)
			} else {
				log.Println("Redis cache MISS, report saved:", key)
			}
		}
	}

	return report, nil
}
