package api

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
	testServer *Server
	testRedis  *redis.Client
	testCtx    context.Context
)

func TestMain(m *testing.M) {
	log.Printf("Starting test setup...")
	cfg := config.LoadConfig()
	testCtx = context.Background()
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	testRedis = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.RedisPassword,
		DB:       1,
	})

	pong, err := testRedis.Ping(testCtx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Printf("Redis connection successful: %s", pong)

	if err := testRedis.FlushDB(testCtx).Err(); err != nil {
		log.Fatalf("Could not flush test database: %v", err)
	}
	log.Printf("Database flushed successfully")

	testServer, err = NewServer(testRedis, *cfg)
	if err != nil {
		log.Fatalf("Could not create test server: %v", err)
	}
	log.Printf("Server created successfully")

	mux := testServer.router
	if mux == nil {
		log.Fatal("Server router is nil")
	}
	log.Printf("Server router initialized")

	log.Printf("Starting tests...")
	code := m.Run()

	log.Printf("Cleaning up...")
	if err := testRedis.FlushDB(testCtx).Err(); err != nil {
		log.Printf("Warning: Could not flush test database after tests: %v", err)
	}
	if err := testRedis.Close(); err != nil {
		log.Printf("Warning: Could not close Redis connection: %v", err)
	}
	os.Exit(code)
}
