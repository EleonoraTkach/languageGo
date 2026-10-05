package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	m "ledger/models"
	"ledger/repository"
)

func runBudgetScenario() (*repository.Storage, error) {
	storage := repository.NewStorage()

	storage.SetBudget(m.Budget{
		Category: "еда",
		Limit:    5000,
		From:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC),
	})

	fmt.Println("Бюджет: 5000")

	file, err := os.Open("budgets.json")
	if err != nil {
		fmt.Println("Ошибка открытия budgets.json:", err)
	} else {
		defer file.Close()

		reader := bufio.NewReader(file)

		if err := storage.LoadBudgets(reader); err != nil {
			fmt.Println("Ошибка загрузки бюджетов:", err)
		} else {
			fmt.Println("Бюджеты из файла успешно загружены")
		}
	}

	fmt.Println("Бюджеты успешно загружены")
	budgets := storage.ListBudgets()

	fmt.Println("Количество бюджетов:", len(budgets))
	fmt.Println("Список бюджетов:")

	for _, budget := range budgets {
		fmt.Printf(
			"Категория: %s, Лимит: %.2f, Период: %s - %s\n",
			budget.Category,
			budget.Limit,
			budget.From.Format("02.01.2006"),
			budget.To.Format("02.01.2006"),
		)
	}

	err = storage.AddTransaction(m.Transaction{
		Amount:      3000,
		Category:    "еда",
		Description: "продукты",
		Date:        time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC),
	})

	if err != nil {
		fmt.Println("Первая транзакция: ошибка:", err)
	} else {
		fmt.Println("Первая транзакция: успешно добавлена")
	}

	err = storage.AddTransaction(m.Transaction{
		Amount:      2500,
		Category:    "еда",
		Description: "ресторан",
		Date:        time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC),
	})

	if err != nil {
		fmt.Println("Вторая транзакция: ошибка:", err)
	} else {
		fmt.Println("Вторая транзакция: успешно добавлена")
	}

	err = storage.AddTransaction(m.Transaction{
		Amount:      2000,
		Category:    "еда",
		Description: "продукты в ноябре",
		Date:        time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC),
	})

	if err != nil {
		fmt.Println("Транзакция вне периода: ошибка:", err)
	} else {
		fmt.Println("Транзакция вне периода: успешно добавлена")
	}

	err = storage.AddTransaction(m.Transaction{
		Amount:      4000,
		Category:    "развлечения",
		Description: "кино",
		Date:        time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC),
	})

	if err != nil {
		fmt.Println("Транзакция без бюджета: ошибка:", err)
	} else {
		fmt.Println("Транзакция без бюджета: успешно добавлена")
	}

	transactions := storage.ListTransactions()

	fmt.Println("Количество сохранённых транзакций:", len(transactions))

	fmt.Println("Список транзакций:")
	for _, tx := range transactions {
		fmt.Printf(
			"Категория: %s, Сумма: %.2f, Описание: %s, Дата: %s\n",
			tx.Category,
			tx.Amount,
			tx.Description,
			tx.Date.Format("02.01.2006 15:04"),
		)
	}

	return storage, nil
}

func CheckValid(v m.Validatable) error {
	return v.Validate()
}

func main() {
	_, err := runBudgetScenario()

	if err != nil {
		fmt.Println("Ошибка выполнения:", err)
	}

	transaction := m.Transaction{
		ID:          1,
		Amount:      100,
		Category:    "",
		Description: "Lunch",
		Date:        time.Now(),
	}

	budget := m.Budget{
		Category: "Food",
		Limit:    -1,
		From:     time.Now(),
		To:       time.Now().AddDate(0, 1, 0),
	}

	if err := CheckValid(transaction); err != nil {
		fmt.Println("Transaction:", err)
	} else {
		fmt.Println("Transaction is valid")
	}

	if err := CheckValid(budget); err != nil {
		fmt.Println("Budget:", err)
	} else {
		fmt.Println("Budget is valid")
	}
}
