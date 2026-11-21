package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	api "api-gateway/internal/api"
	"api-gateway/internal/ledger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ledger ledger.Client
}

func NewServer(client ledger.Client) *Server {
	return &Server{
		ledger: client,
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

func (s *Server) createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	tx := ledger.Transaction{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}

	createdTx, err := s.ledger.AddTransaction(r.Context(), tx)
	if err != nil {
		log.Printf("add transaction: %v", err)
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, api.TransactionResponse{
		ID:          strconv.FormatInt(createdTx.ID, 10),
		Amount:      createdTx.Amount,
		Category:    createdTx.Category,
		Description: createdTx.Description,
		Date:        createdTx.Date,
	})
}

func (s *Server) listTransactionsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	transactions, err := s.ledger.ListTransactions(r.Context())
	if err != nil {
		log.Printf("list transactions: %v", err)
		writeLedgerError(w, err)
		return
	}

	response := make([]api.TransactionResponse, 0, len(transactions))

	for _, tx := range transactions {
		response = append(response, api.TransactionResponse{
			ID:          strconv.FormatInt(tx.ID, 10),
			Amount:      tx.Amount,
			Category:    tx.Category,
			Description: tx.Description,
			Date:        tx.Date,
		})
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

	budget := ledger.Budget{
		Category: req.Category,
		Limit:    req.Limit,
	}

	createdBudget, err := s.ledger.SetBudget(r.Context(), budget)
	if err != nil {
		log.Printf("set budget: %v", err)
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, api.BudgetResponse{
		Category: createdBudget.Category,
		Limit:    createdBudget.Limit,
	})
}

func (s *Server) listBudgetsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	budgets, err := s.ledger.ListBudgets(r.Context())
	if err != nil {
		log.Printf("list budget: %v", err)
		writeLedgerError(w, err)
		return
	}

	response := make([]api.BudgetResponse, 0, len(budgets))

	for _, budget := range budgets {
		response = append(response, api.BudgetResponse{
			Category: budget.Category,
			Limit:    budget.Limit,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func writeLedgerError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": status.Convert(err).Message(),
		})

	case codes.ResourceExhausted, codes.FailedPrecondition:
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "budget exceeded",
		})

	case codes.NotFound:
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": status.Convert(err).Message(),
		})

	case codes.DeadlineExceeded:
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{
			"error": "request timeout",
		})

	case codes.Canceled:
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{
			"error": "request canceled",
		})

	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
	}
}

func (s *Server) reportSummaryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	fromValue := r.URL.Query().Get("from")
	if fromValue == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "query parameter 'from' is required",
		})
		return
	}

	toValue := r.URL.Query().Get("to")
	if toValue == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "query parameter 'to' is required",
		})
		return
	}

	const dateLayout = "2006-01-02"

	from, err := time.Parse(dateLayout, fromValue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid 'from' date: expected YYYY-MM-DD",
		})
		return
	}

	to, err := time.Parse(dateLayout, toValue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid 'to' date: expected YYYY-MM-DD",
		})
		return
	}

	if from.After(to) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "'from' date must not be after 'to' date",
		})
		return
	}

	summary, err := s.ledger.GetReportSummary(r.Context(), from, to)
	if err != nil {
		log.Printf("failed to get report summary: %v", err)
		writeLedgerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
