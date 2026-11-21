package grpc

import (
	"context"
	"fmt"
	"time"

	ledgerv1 "api-gateway/gen/ledger/v1"
	"api-gateway/internal/ledger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	conn   *grpc.ClientConn
	client ledgerv1.LedgerServiceClient
}

func NewClient(target string) (*Client, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	return &Client{
		conn:   conn,
		client: ledgerv1.NewLedgerServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &ledgerv1.PingRequest{})
	return err
}

func (c *Client) AddTransaction(
	ctx context.Context,
	tx ledger.Transaction,
) (ledger.Transaction, error) {
	response, err := c.client.AddTransaction(ctx, &ledgerv1.AddTransactionRequest{
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        timestamppb.New(tx.Date),
	})
	if err != nil {
		return ledger.Transaction{}, err
	}

	return ledger.Transaction{
		ID:          response.GetId(),
		Amount:      response.GetAmount(),
		Category:    response.GetCategory(),
		Description: response.GetDescription(),
		Date:        response.GetDate().AsTime(),
	}, nil
}

func (c *Client) ListTransactions(
	ctx context.Context,
) ([]ledger.Transaction, error) {
	response, err := c.client.ListTransactions(
		ctx,
		&ledgerv1.ListTransactionsRequest{},
	)
	if err != nil {
		return nil, err
	}

	transactions := make([]ledger.Transaction, 0, len(response.GetTransactions()))

	for _, tx := range response.GetTransactions() {
		item := ledger.Transaction{
			ID:          tx.GetId(),
			Amount:      tx.GetAmount(),
			Category:    tx.GetCategory(),
			Description: tx.GetDescription(),
		}

		if tx.GetDate() != nil {
			item.Date = tx.GetDate().AsTime()
		}

		transactions = append(transactions, item)
	}

	return transactions, nil
}

func (c *Client) SetBudget(
	ctx context.Context,
	budget ledger.Budget,
) (ledger.Budget, error) {
	response, err := c.client.SetBudget(ctx, &ledgerv1.SetBudgetRequest{
		Category: budget.Category,
		Limit:    budget.Limit,
	})
	if err != nil {
		return ledger.Budget{}, err
	}

	return ledger.Budget{
		Category: response.GetCategory(),
		Limit:    response.GetLimit(),
	}, nil
}

func (c *Client) ListBudgets(
	ctx context.Context,
) ([]ledger.Budget, error) {
	response, err := c.client.ListBudgets(
		ctx,
		&ledgerv1.ListBudgetsRequest{},
	)
	if err != nil {
		return nil, err
	}

	budgets := make([]ledger.Budget, 0, len(response.GetBudgets()))

	for _, budget := range response.GetBudgets() {
		budgets = append(budgets, ledger.Budget{
			Category: budget.GetCategory(),
			Limit:    budget.GetLimit(),
		})
	}

	return budgets, nil
}

func (c *Client) GetReportSummary(
	ctx context.Context,
	from time.Time,
	to time.Time,
) (map[string]float64, error) {
	response, err := c.client.GetReportSummary(
		ctx,
		&ledgerv1.GetReportSummaryRequest{
			From: timestamppb.New(from),
			To:   timestamppb.New(to),
		},
	)
	if err != nil {
		return nil, err
	}

	return response.GetSummary(), nil
}
