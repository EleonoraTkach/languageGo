package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

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

func (s *LedgerService) GetReportSummary(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (map[string]float64, error) {
	// Если запрос уже отменён, не начинаем расчёт.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Получаем все категории.
	categories, err := s.transactions.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(categories) == 0 {
		return map[string]float64{}, nil
	}

	// Для каждой категории будет отдельный результат и ошибка.
	results := make([]float64, len(categories))
	errs := make([]error, len(categories))

	// Позволяем остановить остальные запросы,
	// если один из них завершился ошибкой.
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Heartbeat.
	heartbeatDone := make(chan struct{})
	heartbeatStopped := make(chan struct{})

	go func() {
		defer close(heartbeatStopped)

		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Report calculation is in progress...")

			case <-workCtx.Done():
				return

			case <-heartbeatDone:
				return
			}
		}
	}()

	var wg sync.WaitGroup

	// Запускаем отдельную горутину для каждой категории.
	for i, category := range categories {
		wg.Add(1)

		go func(i int, category string) {
			defer wg.Done()

			if err := workCtx.Err(); err != nil {
				errs[i] = err
				return
			}

			total, err := s.transactions.GetSpentByCategoryAndPeriod(
				workCtx,
				category,
				from,
				to,
			)

			if err != nil {
				errs[i] = err
				cancel()
				return
			}

			results[i] = total
		}(i, category)
	}

	// Ждём завершения всех вычислений.
	wg.Wait()

	// Останавливаем heartbeat и дожидаемся его завершения.
	close(heartbeatDone)
	<-heartbeatStopped

	// Проверяем отмену исходного HTTP/gRPC-запроса.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Сначала ищем настоящую ошибку, а не отмену
	// соседнего запроса, вызванную этой ошибкой.
	for _, err := range errs {
		if err != nil &&
			!errors.Is(err, context.Canceled) &&
			!errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("failed to calculate report: %w", err)
		}
	}

	for _, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("report calculation interrupted: %w", err)
		}
	}

	// Собираем итоговую map после завершения всех горутин.
	summary := make(map[string]float64, len(categories))

	for i, category := range categories {
		summary[category] = results[i]
	}

	return summary, nil
}
