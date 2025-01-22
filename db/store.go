package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/grysj/remitly-api/parser"
	"github.com/grysj/remitly-api/util"
	"github.com/redis/go-redis/v9"
)

const bankKeyPrefix = "swiftCode:"
const iso2IndexKey = "idx:countryISO2"
const countryCode = "countryISO2:name"

func (s *RedisStore) AddBanksFromCSV(rows []parser.CsvRow) error {
	ctx := context.Background()
	pipe := s.client.TxPipeline()

	for _, row := range rows {
		bankData := Bank{
			Swift:      row.Swift,
			ISO2:       strings.ToUpper(row.ISO2),
			Name:       strings.ToUpper(row.Name),
			Type:       row.Type,
			Address:    row.Address,
			Town:       row.Town,
			Country:    row.Country,
			Timezone:   row.Timezone,
			Headquater: util.CheckIfHeadquater(row.Swift),
		}

		bankKey := bankKeyPrefix + row.Swift
		pipe.HSet(ctx, bankKey, &bankData)
		pipe.SAdd(ctx, iso2IndexKey+":"+row.ISO2, bankKey)

		if !util.CheckIfHeadquater(row.Swift) {
			pipe.SAdd(ctx, "branch:"+util.GetPrefix(row.Swift), row.Swift)
		}
		pipe.HSet(ctx, "countries", strings.ToUpper(row.ISO2), row.Country)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) AddBankToDB(bank Bank) error {
	ctx := context.Background()
	pipe := s.client.TxPipeline()

	if len(bank.ISO2) != 2 {
		return fmt.Errorf("invalid ISO2 format: must be exactly 2 letters")
	}
	if bank.Country == "" {
		return fmt.Errorf("country name cannot be empty")
	}

	formattedBank := Bank{
		Swift:      bank.Swift,
		ISO2:       strings.ToUpper(bank.ISO2),
		Name:       strings.ToUpper(bank.Name),
		Type:       bank.Type,
		Address:    bank.Address,
		Town:       bank.Town,
		Country:    bank.Country,
		Timezone:   bank.Timezone,
		Headquater: util.CheckIfHeadquater(bank.Swift),
	}

	bankKey := bankKeyPrefix + bank.Swift
	pipe.HSet(ctx, bankKey, &formattedBank)
	pipe.SAdd(ctx, iso2IndexKey+":"+formattedBank.ISO2, bankKey)
	if !util.CheckIfHeadquater(bank.Swift) {
		pipe.SAdd(ctx, "branch:"+util.GetPrefix(bank.Swift), bank.Swift)
	}
	pipe.HSet(ctx, "countries", strings.ToUpper(bank.ISO2), bank.Country)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) DeleteBankFromDB(bank DeleteBankParams) error {
	ctx := context.Background()
	pipe := s.client.TxPipeline()

	bankKey := bankKeyPrefix + bank.Swift
	bankData := &Bank{}
	err := s.client.HGetAll(ctx, bankKey).Scan(bankData)
	if err != nil {
		return fmt.Errorf("failed to get bank data: %w", err)
	}

	pipe.Del(ctx, bankKey)
	pipe.SRem(ctx, iso2IndexKey+":"+bankData.ISO2, bankKey)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetBanksByISO2(iso2 string) ([]GetBankByIsoResult, error) {
	ctx := context.Background()
	bankKeys, err := s.client.SMembers(ctx, iso2IndexKey+":"+strings.ToUpper(iso2)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get bank keys for ISO2 %s: %w", iso2, err)
	}

	if len(bankKeys) == 0 {
		return []GetBankByIsoResult{}, nil
	}

	pipe := s.client.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(bankKeys))
	for i, key := range bankKeys {
		cmds[i] = pipe.HGetAll(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank data: %w", err)
	}

	banks := make([]GetBankByIsoResult, len(bankKeys))
	for i, cmd := range cmds {
		err = cmd.Scan(&banks[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse bank data: %w", err)
		}
	}

	return banks, nil
}

func (s *RedisStore) GetBankBranches(swift string) ([]GetBranchesBySwiftResult, error) {
	ctx := context.Background()
	branchSet := "branch:" + util.GetPrefix(swift)

	exists, err := s.client.Exists(ctx, branchSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check branch set: %w", err)
	}
	if exists == 0 {
		return []GetBranchesBySwiftResult{}, nil
	}

	branchSwifts, err := s.client.SMembers(ctx, branchSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get branch swifts: %w", err)
	}

	if len(branchSwifts) == 0 {
		return []GetBranchesBySwiftResult{}, nil
	}

	pipe := s.client.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(branchSwifts))
	for i, branchSwift := range branchSwifts {
		cmds[i] = pipe.HGetAll(ctx, bankKeyPrefix+branchSwift)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch data: %w", err)
	}

	branches := make([]GetBranchesBySwiftResult, len(branchSwifts))
	for i, cmd := range cmds {
		err = cmd.Scan(&branches[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse branch data: %w", err)
		}
	}

	return branches, nil
}

func (s *RedisStore) GetBankFromSwift(swift string) (*GetBankBySwiftResult, error) {
	ctx := context.Background()
	bankKey := bankKeyPrefix + strings.ToUpper(swift)

	exists, err := s.client.Exists(ctx, bankKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check bank existence: %w", err)
	}
	if exists == 0 {
		return nil, nil
	}

	cmd := s.client.HGetAll(ctx, bankKey)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to retrieve bank data: %w", cmd.Err())
	}

	var bank GetBankBySwiftResult
	err = cmd.Scan(&bank)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bank data: %w", err)
	}

	return &bank, nil
}

func (s *RedisStore) DeleteBanksBySwiftPrefix(swiftPrefix string) error {
	ctx := context.Background()
	hqKey := bankKeyPrefix + swiftPrefix + "XXX"
	var hqBank Bank
	err := s.client.HGetAll(ctx, hqKey).Scan(&hqBank)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get headquarters info: %w", err)
	}

	pipe := s.client.Pipeline()

	if hqBank.ISO2 != "" {
		pipe.SRem(ctx, iso2IndexKey+":"+hqBank.ISO2, hqKey)
		pipe.Del(ctx, hqKey)
	}

	branchSetKey := "branch:" + swiftPrefix
	branchSwifts, err := s.client.SMembers(ctx, branchSetKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get branch members: %w", err)
	}

	for _, swift := range branchSwifts {
		bankKey := bankKeyPrefix + swift
		var branch Bank
		err := s.client.HGetAll(ctx, bankKey).Scan(&branch)
		if err == nil && branch.ISO2 != "" {
			pipe.SRem(ctx, iso2IndexKey+":"+branch.ISO2, bankKey)
		}
		pipe.Del(ctx, bankKey)
	}

	if len(branchSwifts) > 0 {
		pipe.Del(ctx, branchSetKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute deletion pipeline: %w", err)
	}

	exists, err := s.client.Exists(ctx, hqKey).Result()
	if err != nil {
		return fmt.Errorf("failed to verify cleanup: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("headquarters not properly deleted")
	}

	return nil
}

func (s *RedisStore) GetCountryNameByISO2(iso2 string) (string, error) {
	ctx := context.Background()
	countryName, err := s.client.HGet(ctx, "countries", strings.ToUpper(iso2)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to retrieve country name: %w", err)
	}
	return countryName, nil
}
