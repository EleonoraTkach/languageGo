package main

import (
	"fmt"
	"ledger/repository"
)

func main() {
	store := repository.NewStorage()

	store.Add(1500.50, "Продукты", "Магазин у дома")
	store.Add(450.00, "Транспорт", "Такси")
	_, err := store.Add(0.00, "Транспорт", "Такси")
	if err != nil {
		fmt.Println(err)
	}

	list := store.ListTransactions()

	fmt.Println("--- Список всех транзакций ---")
	for _, tx := range list {
		fmt.Printf("[%d] %s: %.2f руб. (%s)\n", tx.ID, tx.Category, tx.Amount, tx.Description)
	}
}
