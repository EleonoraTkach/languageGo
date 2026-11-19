package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	ledgerv1 "ledger/gen/ledger/v1"
	"ledger/internal/app"
	grpcTransport "ledger/internal/transport/grpc"

	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Инициализируем бизнес-логику и PostgreSQL.
	ledgerService, closeFn, err := app.NewLedger(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize ledger: %w", err)
	}
	defer closeFn()

	// Открываем TCP-порт для gRPC.
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	// Создаём gRPC-сервер.
	grpcServer := grpc.NewServer()

	// Создаём и регистрируем обработчик Ledger.
	handler := grpcTransport.NewServer(ledgerService)

	ledgerv1.RegisterLedgerServiceServer(
		grpcServer,
		handler,
	)

	log.Println("Ledger gRPC server started on :50051")

	// Запускаем сервер.
	serveErr := make(chan error, 1)

	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("gRPC server failed: %w", err)
		}

	case <-ctx.Done():
		log.Println("Shutting down Ledger gRPC server...")
		grpcServer.GracefulStop()
	}

	log.Println("Ledger gRPC server stopped")

	return nil
}
