package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/grysj/remitly-api/config"
	"github.com/redis/go-redis/v9"
)

var (
	testRedis *redis.Client
	testCtx   context.Context
)

func TestMain(m *testing.M) {
	cfg := config.LoadConfig()
	testCtx = context.Background()
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	testRedis = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	_, err := testRedis.Ping(testCtx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	if err := testRedis.FlushDB(testCtx).Err(); err != nil {
		log.Fatalf("Could not flush test database: %v", err)
	}

	code := m.Run()

	if err := testRedis.FlushDB(testCtx).Err(); err != nil {
		log.Printf("Warning: Could not flush test database after tests: %v", err)
	}
	if err := testRedis.Close(); err != nil {
		log.Printf("Warning: Could not close Redis connection: %v", err)
	}

	os.Exit(code)
}
