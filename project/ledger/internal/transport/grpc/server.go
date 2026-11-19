package grpc

import (
	"context"
	"errors"

	ledgerv1 "ledger/gen/ledger/v1"
	d "ledger/internal/domain"
	"ledger/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	ledgerv1.UnimplementedLedgerServiceServer
	service service.Service
}

func NewServer(svc service.Service) *Server {
	return &Server{
		service: svc,
	}
}

func (s *Server) Ping(
	ctx context.Context,
	req *ledgerv1.PingRequest,
) (*ledgerv1.PingResponse, error) {
	return &ledgerv1.PingResponse{
		Status: "pong",
	}, nil
}

func (s *Server) AddTransaction(
	ctx context.Context,
	req *ledgerv1.AddTransactionRequest,
) (*ledgerv1.TransactionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.Date == nil || !req.Date.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid transaction date")
	}

	tx := d.Transaction{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date.AsTime(),
	}

	createdTx, err := s.service.AddTransaction(ctx, tx)

	if err != nil {
		return nil, mapError(err)
	}

	return &ledgerv1.TransactionResponse{
		Id:          int64(createdTx.ID),
		Amount:      createdTx.Amount,
		Category:    createdTx.Category,
		Description: createdTx.Description,
		Date:        timestamppb.New(createdTx.Date),
	}, nil
}

func (s *Server) ListTransactions(
	ctx context.Context,
	req *ledgerv1.ListTransactionsRequest,
) (*ledgerv1.ListTransactionsResponse, error) {
	transactions, err := s.service.ListTransactions(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	response := &ledgerv1.ListTransactionsResponse{
		Transactions: make([]*ledgerv1.TransactionResponse, 0, len(transactions)),
	}

	for _, tx := range transactions {
		response.Transactions = append(response.Transactions, &ledgerv1.TransactionResponse{
			Id:          int64(tx.ID),
			Amount:      tx.Amount,
			Category:    tx.Category,
			Description: tx.Description,
			Date:        timestamppb.New(tx.Date),
		})
	}

	return response, nil
}

func (s *Server) SetBudget(
	ctx context.Context,
	req *ledgerv1.SetBudgetRequest,
) (*ledgerv1.BudgetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	budget := d.Budget{
		Category: req.Category,
		Limit:    req.Limit,
	}

	if err := s.service.SetBudget(ctx, budget); err != nil {
		return nil, mapError(err)
	}

	return &ledgerv1.BudgetResponse{
		Category: budget.Category,
		Limit:    budget.Limit,
	}, nil
}

func (s *Server) ListBudgets(
	ctx context.Context,
	req *ledgerv1.ListBudgetsRequest,
) (*ledgerv1.ListBudgetsResponse, error) {
	budgets, err := s.service.ListBudgets(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	response := &ledgerv1.ListBudgetsResponse{
		Budgets: make([]*ledgerv1.BudgetResponse, 0, len(budgets)),
	}

	for _, budget := range budgets {
		response.Budgets = append(response.Budgets, &ledgerv1.BudgetResponse{
			Category: budget.Category,
			Limit:    budget.Limit,
		})
	}

	return response, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, d.ErrBudgetExceeded):
		return status.Error(codes.ResourceExhausted, err.Error())

	case errors.Is(err, d.ErrBudgetNotFound):
		return status.Error(codes.NotFound, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
