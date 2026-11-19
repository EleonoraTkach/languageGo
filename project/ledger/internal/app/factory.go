package app

import (
	"context"
	"fmt"

	"ledger/db"
	"ledger/internal/repository/postgres"
	"ledger/internal/service"
)

func NewLedger(ctx context.Context) (service.Service, func(), error) {
	database, err := db.NewDB(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	storage := postgres.NewStorage(database)

	ledgerService := service.NewLedgerService(
		storage,
		storage,
	)

	closeFn := func() {
		_ = database.Close()
	}

	return ledgerService, closeFn, nil
}
