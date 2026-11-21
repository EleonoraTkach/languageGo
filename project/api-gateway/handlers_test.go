package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	api "api-gateway/internal/api"
	"api-gateway/internal/ledger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const jsonContentType = "application/json; charset=utf-8"

// mockLedger имитирует удалённый Ledger.
type mockLedger struct {
	budgets      []ledger.Budget
	transactions []ledger.Transaction
	nextID       int64
}

func newMockLedger() *mockLedger {
	return &mockLedger{
		budgets:      make([]ledger.Budget, 0),
		transactions: make([]ledger.Transaction, 0),
		nextID:       1,
	}
}

func (m *mockLedger) Ping(ctx context.Context) error {
	return nil
}

func (m *mockLedger) AddTransaction(
	ctx context.Context,
	tx ledger.Transaction,
) (ledger.Transaction, error) {
	if tx.Amount <= 0 {
		return ledger.Transaction{}, status.Error(
			codes.InvalidArgument,
			"amount must be greater than 0",
		)
	}

	if tx.Category == "" || tx.Date.IsZero() {
		return ledger.Transaction{}, status.Error(
			codes.InvalidArgument,
			"invalid transaction",
		)
	}

	for _, budget := range m.budgets {
		if budget.Category != tx.Category {
			continue
		}

		var spent float64

		for _, existing := range m.transactions {
			if existing.Category == tx.Category {
				spent += existing.Amount
			}
		}

		if spent+tx.Amount > budget.Limit {
			return ledger.Transaction{}, status.Error(
				codes.ResourceExhausted,
				"budget exceeded",
			)
		}

		break
	}

	tx.ID = m.nextID
	m.nextID++

	m.transactions = append(m.transactions, tx)

	return tx, nil
}

func (m *mockLedger) ListTransactions(
	ctx context.Context,
) ([]ledger.Transaction, error) {
	result := make([]ledger.Transaction, len(m.transactions))
	copy(result, m.transactions)

	return result, nil
}

func (m *mockLedger) SetBudget(
	ctx context.Context,
	budget ledger.Budget,
) (ledger.Budget, error) {
	if budget.Category == "" || budget.Limit <= 0 {
		return ledger.Budget{}, status.Error(
			codes.InvalidArgument,
			"invalid budget",
		)
	}

	// UPSERT: обновляем бюджет, если категория уже существует.
	for i, existing := range m.budgets {
		if existing.Category == budget.Category {
			m.budgets[i] = budget
			return budget, nil
		}
	}

	m.budgets = append(m.budgets, budget)

	return budget, nil
}

func (m *mockLedger) GetReportSummary(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (map[string]float64, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	summary := make(map[string]float64)

	for _, tx := range m.transactions {
		if tx.Date.Before(from) {
			continue
		}

		// Дата окончания включается в отчёт целиком.
		if tx.Date.After(to.AddDate(0, 0, 1).Add(-time.Nanosecond)) {
			continue
		}

		summary[tx.Category] += tx.Amount
	}

	return summary, nil
}

func (m *mockLedger) ListBudgets(
	ctx context.Context,
) ([]ledger.Budget, error) {
	result := make([]ledger.Budget, len(m.budgets))
	copy(result, m.budgets)

	return result, nil
}

func newTestRouter() http.Handler {
	client := newMockLedger()

	server := NewServer(client)
	router := NewRouter(server)

	return router.SetupRoutes()
}

func request(
	handler http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		method,
		path,
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func checkStatus(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	want int,
) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf(
			"status = %d, want %d; response: %s",
			rec.Code,
			want,
			rec.Body.String(),
		)
	}
}

func checkContentType(
	t *testing.T,
	rec *httptest.ResponseRecorder,
) {
	t.Helper()

	got := rec.Header().Get("Content-Type")

	if got != jsonContentType {
		t.Errorf(
			"Content-Type = %q, want %q",
			got,
			jsonContentType,
		)
	}
}

func checkErrorJSON(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	want string,
) {
	t.Helper()

	var response map[string]string

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf(
			"invalid error JSON: %v; body=%s",
			err,
			rec.Body.String(),
		)
	}

	if response["error"] == "" {
		t.Fatal(`response must contain a non-empty "error" field`)
	}

	if want != "" && response["error"] != want {
		t.Errorf(
			"error = %q, want %q",
			response["error"],
			want,
		)
	}
}

func TestBudgetsHTTP(t *testing.T) {
	handler := newTestRouter()

	t.Run("ok", func(t *testing.T) {
		body := `{
			"category": "еда",
			"limit": 5000,
			"from": "2026-10-01T00:00:00Z",
			"to": "2026-10-31T23:59:59Z"
		}`

		rec := request(
			handler,
			http.MethodPost,
			"/api/budgets",
			body,
		)

		checkStatus(t, rec, http.StatusCreated)
		checkContentType(t, rec)

		var response api.BudgetResponse

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		if response.Category != "еда" {
			t.Errorf("category = %q, want еда", response.Category)
		}

		if response.Limit != 5000 {
			t.Errorf("limit = %v, want 5000", response.Limit)
		}
	})

	t.Run("get_contains_budget", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodGet,
			"/api/budgets",
			"",
		)

		checkStatus(t, rec, http.StatusOK)
		checkContentType(t, rec)

		var budgets []api.BudgetResponse

		if err := json.Unmarshal(rec.Body.Bytes(), &budgets); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		found := false

		for _, budget := range budgets {
			if budget.Category == "еда" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("created budget not found")
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodPost,
			"/api/budgets",
			`{"category":`,
		)

		checkStatus(t, rec, http.StatusBadRequest)
		checkContentType(t, rec)
		checkErrorJSON(t, rec, "")
	})

	t.Run("invalid_limit", func(t *testing.T) {
		body := `{
			"category": "еда",
			"limit": 0
		}`

		rec := request(
			handler,
			http.MethodPost,
			"/api/budgets",
			body,
		)

		checkStatus(t, rec, http.StatusBadRequest)
		checkContentType(t, rec)
		checkErrorJSON(t, rec, "")
	})
}

func TestTransactionsHTTP(t *testing.T) {
	handler := newTestRouter()

	t.Run("create_budget", func(t *testing.T) {
		body := `{
			"category": "еда",
			"limit": 5000
		}`

		rec := request(
			handler,
			http.MethodPost,
			"/api/budgets",
			body,
		)

		checkStatus(t, rec, http.StatusCreated)
		checkContentType(t, rec)
	})

	t.Run("ok", func(t *testing.T) {
		body := `{
			"amount": 3000,
			"category": "еда",
			"description": "продукты",
			"date": "2026-10-10T00:00:00Z"
		}`

		rec := request(
			handler,
			http.MethodPost,
			"/api/transactions",
			body,
		)

		checkStatus(t, rec, http.StatusCreated)
		checkContentType(t, rec)

		var response api.TransactionResponse

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		if response.Category != "еда" {
			t.Errorf("category = %q, want еда", response.Category)
		}

		if response.Amount != 3000 {
			t.Errorf("amount = %v, want 3000", response.Amount)
		}

		if response.ID == "" {
			t.Error("transaction ID must not be empty")
		}
	})

	t.Run("get_contains_transaction", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodGet,
			"/api/transactions",
			"",
		)

		checkStatus(t, rec, http.StatusOK)
		checkContentType(t, rec)

		var transactions []api.TransactionResponse

		if err := json.Unmarshal(rec.Body.Bytes(), &transactions); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		found := false

		for _, tx := range transactions {
			if tx.Category == "еда" && tx.Amount == 3000 {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("created transaction not found")
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		body := `{
			"amount": 2500,
			"category": "еда",
			"description": "ресторан",
			"date": "2026-10-15T00:00:00Z"
		}`

		rec := request(
			handler,
			http.MethodPost,
			"/api/transactions",
			body,
		)

		checkStatus(t, rec, http.StatusConflict)
		checkContentType(t, rec)
		checkErrorJSON(t, rec, "budget exceeded")
	})

	t.Run("bad_json", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodPost,
			"/api/transactions",
			`{"amount":`,
		)

		checkStatus(t, rec, http.StatusBadRequest)
		checkContentType(t, rec)
		checkErrorJSON(t, rec, "")
	})

}

func TestReportSummaryHTTP(t *testing.T) {
	handler := newTestRouter()

	// Подготавливаем данные для отчёта.
	createTransaction := func(amount float64, category, date string) {
		t.Helper()

		body := `{
			"amount": ` + strconv.FormatFloat(amount, 'f', 2, 64) + `,
			"category": "` + category + `",
			"description": "test",
			"date": "` + date + `"
		}`

		rec := request(
			handler,
			http.MethodPost,
			"/api/transactions",
			body,
		)

		if rec.Code != http.StatusCreated {
			t.Fatalf("failed to create transaction: %s", rec.Body.String())
		}
	}

	createTransaction(100, "еда", "2026-11-10T12:00:00Z")
	createTransaction(200, "еда", "2026-11-15T12:00:00Z")
	createTransaction(50, "транспорт", "2026-11-12T12:00:00Z")
	createTransaction(300, "еда", "2026-12-01T12:00:00Z")

	t.Run("ok", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodGet,
			"/api/reports/summary?from=2026-11-01&to=2026-11-30",
			"",
		)

		checkStatus(t, rec, http.StatusOK)
		checkContentType(t, rec)

		var summary map[string]float64
		if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		if summary["еда"] != 300 {
			t.Errorf("еда = %v, want 300", summary["еда"])
		}

		if summary["транспорт"] != 50 {
			t.Errorf("транспорт = %v, want 50", summary["транспорт"])
		}
	})

	t.Run("missing_from", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodGet,
			"/api/reports/summary?to=2026-11-30",
			"",
		)

		checkStatus(t, rec, http.StatusBadRequest)
		checkErrorJSON(t, rec, "")
	})

	t.Run("invalid_date", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodGet,
			"/api/reports/summary?from=wrong&to=2026-11-30",
			"",
		)

		checkStatus(t, rec, http.StatusBadRequest)
		checkErrorJSON(t, rec, "")
	})

	t.Run("from_after_to", func(t *testing.T) {
		rec := request(
			handler,
			http.MethodGet,
			"/api/reports/summary?from=2026-12-01&to=2026-11-01",
			"",
		)

		checkStatus(t, rec, http.StatusBadRequest)
		checkErrorJSON(t, rec, "")
	})
}
