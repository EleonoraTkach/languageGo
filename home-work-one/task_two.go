package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	username := os.Getenv("USERNAME")
	fmt.Printf("Имя пользователя: %s\n", username)

	args := os.Args[1:]
	fmt.Printf("Аргументы CLI (%d): %v\n", len(args), args)

	fmt.Printf("Версия Go: %s\n", runtime.Version())
}
