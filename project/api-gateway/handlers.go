package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	api "api-gateway/internal/api"
	rep "ledger/repository"
)

type Server struct {
	storage *rep.Storage
}

func NewServer(storage *rep.Storage) *Server {
	return &Server{
		storage: storage,
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

func notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "not implemented",
	})
}

func (s *Server) createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	tx := req.ToTransaction()

	if err := tx.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := s.storage.AddTransaction(tx); err != nil {
		if errors.Is(err, rep.ErrBudgetExceeded) {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "budget exceeded",
			})
			return
		}

		log.Printf("add transaction: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusCreated, api.NewTransactionResponse(tx))
}

func (s *Server) listTransactionsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	transactions, err := s.storage.ListTransactions()
	if err != nil {
		http.Error(
			w,
			"Не удалось получить список транзакций",
			http.StatusInternalServerError,
		)
		return
	}

	response := make([]api.TransactionResponse, 0, len(transactions))
	for _, tx := range transactions {
		response = append(response, api.NewTransactionResponse(tx))
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) createBudgetHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req api.CreateBudgetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	budget := req.ToBudget()

	if err := budget.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := s.storage.SetBudget(budget); err != nil {
		log.Printf("set budget: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusCreated, api.NewBudgetResponse(budget))
}

func (s *Server) listBudgetsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	budgets, err := s.storage.ListBudgets()

	if err != nil {
		http.Error(
			w,
			"Не удалось получить список бюджетов",
			http.StatusInternalServerError,
		)
		return
	}

	response := make([]api.BudgetResponse, 0, len(budgets))
	for _, budget := range budgets {
		response = append(response, api.NewBudgetResponse(budget))
	}

	writeJSON(w, http.StatusOK, response)
}
