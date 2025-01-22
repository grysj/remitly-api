package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/grysj/remitly-api/config"
)

var (
	testCtx   context.Context
	testStore *RedisStore
)

func TestMain(m *testing.M) {
	cfg := config.LoadConfig()
	testCtx = context.Background()
	redisStore, err := NewRedisStore(NewRedisStoreParams{
		RedisHost:     cfg.RedisHost,
		RedisPort:     cfg.RedisPort,
		RedisPassword: cfg.RedisPassword,
		RedisDB:       2,
	})
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	var ok bool
	testStore, ok = redisStore.DBQuerier.(*RedisStore)

	if !ok {
		log.Fatalf("testStore.DBQuerier is not a *RedisStore")
	}

	if err := testStore.CleanDB(testCtx); err != nil {
		log.Fatalf("Could not flush test database: %v", err)
	}

	code := m.Run()
	testStore.CleanDB(testCtx)
	testStore.CloseConnection()

	os.Exit(code)
}
