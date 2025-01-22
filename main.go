package main

import (
	"log"

	"github.com/grysj/remitly-api/api"
	"github.com/grysj/remitly-api/config"
	"github.com/grysj/remitly-api/db"
	"github.com/grysj/remitly-api/parser"
)

func main() {
	cfg := config.LoadConfig()

	store, err := db.NewRedisStore(db.NewRedisStoreParams{
		RedisDB:       0,
		RedisHost:     cfg.RedisHost,
		RedisPort:     cfg.RedisPort,
		RedisPassword: cfg.RedisPassword,
	})
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	parsed, err := parser.ParseCSV(cfg.CsvPath)
	if err != nil {
		log.Fatalf("cannot parse file: %v", err)
	}

	if err := store.AddBanksFromCSV(parsed); err != nil {
		log.Fatalf("cannot init db: %v", err)
	}

	server, err := api.NewServer(store, *cfg)
	if err != nil {
		log.Fatalf("cannot configure server: %v", err)
	}

	if err := server.StartServer("8080"); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}
