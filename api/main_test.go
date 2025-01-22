package api

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/grysj/remitly-api/config"
	"github.com/grysj/remitly-api/db"
)

var (
	testServer *Server
	testCtx    context.Context
	password   string
)

func TestMain(m *testing.M) {
	log.Printf("Starting test setup...")
	cfg := config.LoadConfig()
	testCtx = context.Background()
	testRedis, err := db.NewRedisStore(db.NewRedisStoreParams{
		RedisHost:     cfg.RedisHost,
		RedisPort:     cfg.RedisPort,
		RedisPassword: cfg.RedisPassword,
		RedisDB:       1,
	})
	password = cfg.ApiPassword
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	testServer, err = NewServer(testRedis, *cfg)
	if err != nil {
		log.Fatalf("Could not create test server: %v", err)
	}

	code := m.Run()
	err = testServer.store.CleanDB(testCtx)
	if err != nil {

	}
	err = testServer.store.CloseConnection()
	if err != nil {
		log.Fatalf("Could not close connection: %v", err)
	}
	os.Exit(code)
}
