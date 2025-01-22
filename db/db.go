package db

import (
	"context"
	"fmt"

	"github.com/grysj/remitly-api/parser"
	"github.com/redis/go-redis/v9"
)

type DBQuerier interface {
	AddBanksFromCSV(rows []parser.CsvRow) error
	AddBankToDB(bank Bank) error
	DeleteBankFromDB(bank DeleteBankParams) error
	GetBanksByISO2(iso2 string) ([]GetBankByIsoResult, error)
	GetBankBranches(swift string) ([]GetBranchesBySwiftResult, error)
	DeleteBanksBySwiftPrefix(swiftPrefix string) error
	GetCountryNameByISO2(iso2 string) (string, error)
	GetBankFromSwift(swift string) (*GetBankBySwiftResult, error)
	CleanDB(ctx context.Context) error
	CloseConnection() error
}

type Store struct {
	DBQuerier
}

type RedisStore struct {
	client *redis.Client
}

type NewRedisStoreParams struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

func NewRedisStore(cfg NewRedisStoreParams) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Store{
		DBQuerier: &RedisStore{client: client},
	}, nil
}

func (r *RedisStore) CleanDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
func (r *RedisStore) CloseConnection() error {
	return r.client.Close()
}
