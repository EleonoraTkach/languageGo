package main

import (
	"log"
	"net/http"

	rep "ledger/repository"
)

func main() {
	storage := rep.NewStorage()
	server := NewServer(storage)
	router := NewRouter(server)

	log.Println("Gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router.SetupRoutes()))
}
