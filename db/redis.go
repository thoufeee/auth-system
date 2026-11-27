package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// redis connection
func ConnectRedis() (*redis.Client, error) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")

	dbStr := os.Getenv("REDIS_DB")

	db := 0

	if dbStr != "" {
		parsedDB, err := strconv.Atoi(dbStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
		}
		db = parsedDB
	}

	client := redis.NewClient(&redis.Options{
		Addr:         host + ":" + port,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return client, nil
}
