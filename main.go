package main

import (
	"fmt"
	"log"

	"github.com/grysj/remitly-api/api"
	"github.com/grysj/remitly-api/config"
	"github.com/grysj/remitly-api/db"
	"github.com/grysj/remitly-api/parser"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	conn := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	parsed, err := parser.ParseCSV(cfg.CsvPath)
	if err != nil {
		log.Fatal("cannot parse file:", err)
	}

	err = db.AddBanksToRedis(conn, parsed)
	if err != nil {
		log.Fatal("cannot init db:", err)
	}

	server, err := api.NewServer(conn, *cfg)
	if err != nil {
		log.Fatal("cannot configure server:", err)
	}

	err = server.StartServer("8080")

	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
