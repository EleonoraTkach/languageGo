package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	mu     sync.Mutex
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()

	if Client != nil {
		return nil
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbNumber := 0
	if value := os.Getenv("REDIS_DB"); value != "" {
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		dbNumber = n
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           dbNumber,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return err
	}

	Client = client

	log.Println("Успешное подключение к Redis установлено!")

	return nil
}
