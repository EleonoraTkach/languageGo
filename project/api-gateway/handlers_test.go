package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rep "ledger/repository"
)

const jsonContentType = "application/json; charset=utf-8"

func newTestRouter() http.Handler {
	storage := rep.NewStorage()
	storage.Reset() // Если метода Reset() пока нет — добавь его в repository.Storage.

	server := NewServer(storage)
	router := NewRouter(server)
	return router.SetupRoutes()
}

func request(
	handler http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func checkStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
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

func checkContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	got := rec.Header().Get("Content-Type")
	if got != jsonContentType {
		t.Errorf("Content-Type = %q, want %q", got, jsonContentType)
	}
}

func checkErrorJSON(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid error JSON: %v; body=%s", err, rec.Body.String())
	}

	if response["error"] == "" {
		t.Fatal(`response must contain a non-empty "error" field`)
	}

	if want != "" && response["error"] != want {
		t.Errorf("error = %q, want %q", response["error"], want)
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

		rec := request(handler, http.MethodPost, "/api/budgets", body)

		checkStatus(t, rec, http.StatusCreated)
		checkContentType(t, rec)

		var response map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		if response["category"] != "еда" {
			t.Errorf("category = %v, want еда", response["category"])
		}
		if response["limit"] != float64(5000) {
			t.Errorf("limit = %v, want 5000", response["limit"])
		}
	})

	t.Run("get_contains_budget", func(t *testing.T) {
		rec := request(handler, http.MethodGet, "/api/budgets", "")

		checkStatus(t, rec, http.StatusOK)
		checkContentType(t, rec)

		var budgets []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &budgets); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		found := false
		for _, budget := range budgets {
			if budget["category"] == "еда" {
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
			"limit": 0,
			"from": "2026-10-01T00:00:00Z",
			"to": "2026-10-31T23:59:59Z"
		}`

		rec := request(handler, http.MethodPost, "/api/budgets", body)

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
			"limit": 5000,
			"from": "2026-10-01T00:00:00Z",
			"to": "2026-10-31T23:59:59Z"
		}`

		rec := request(handler, http.MethodPost, "/api/budgets", body)

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

		rec := request(handler, http.MethodPost, "/api/transactions", body)

		checkStatus(t, rec, http.StatusCreated)
		checkContentType(t, rec)

		var response map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		if response["category"] != "еда" {
			t.Errorf("category = %v, want еда", response["category"])
		}
		if response["amount"] != float64(3000) {
			t.Errorf("amount = %v, want 3000", response["amount"])
		}
	})

	t.Run("get_contains_transaction", func(t *testing.T) {
		rec := request(handler, http.MethodGet, "/api/transactions", "")

		checkStatus(t, rec, http.StatusOK)
		checkContentType(t, rec)

		var transactions []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &transactions); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		found := false
		for _, tx := range transactions {
			if tx["category"] == "еда" && tx["amount"] == float64(3000) {
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

		rec := request(handler, http.MethodPost, "/api/transactions", body)

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
