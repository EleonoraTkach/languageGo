package main

import (
	"log"
	"net/http"
	"os"

	grpcclient "api-gateway/internal/grpc"
)

func main() {
	ledgerAddr := os.Getenv("LEDGER_GRPC_ADDR")
	if ledgerAddr == "" {
		ledgerAddr = "localhost:50051"
	}

	ledgerClient, err := grpcclient.NewClient(ledgerAddr)
	if err != nil {
		log.Fatalf("failed to initialize ledger client: %v", err)
	}
	defer ledgerClient.Close()

	server := NewServer(ledgerClient)
	router := NewRouter(server)

	log.Println("Gateway listening on :8080")

	if err := http.ListenAndServe(":8080", router.SetupRoutes()); err != nil {
		log.Fatalf("Gateway server failed: %v", err)
	}
}
