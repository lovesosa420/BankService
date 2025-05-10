package repository

import (
	"BankService/internal/domain/models"
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
)

func (s *Storage) CreateAccount(ctx context.Context, account models.Account) (string, error) {
	var number strings.Builder
	for {
		for i := 0; i < 24; i++ {
			num := rand.Intn(10)
			number.WriteString(fmt.Sprintf("%d", num))
		}
		if s.GetAccount(ctx, number.String()) == nil {
			number.Reset()
			continue
		}
		account.Number = number.String()
		break
	}
	query := "INSERT INTO account (person, number, balance, type) VALUES ($1, $2, $3, $4)"
	_, err := s.Exec(ctx, query, account.Person, account.Number, account.Balance, account.Type)
	if err != nil {
		return "", fmt.Errorf("create account failed: %w", err)
	}
	return account.Number, nil
}

func (s *Storage) UpdateAccountBalance(ctx context.Context, user, number string, sum float64) (float64, error) {
	tx, err := s.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("update account balance failed: %w", err)
	}
	var currentBalance float64
	query := "SELECT balance FROM account WHERE person = $1 AND number = $2"
	err = tx.QueryRow(ctx, query, user, number).Scan(&currentBalance)
	if err != nil {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("reading account balance failed: %w", err)
	}
	if sum < 0 {
		if math.Abs(sum) > currentBalance {
			tx.Rollback(ctx)
			return 0, fmt.Errorf("account balance not enough")
		}
	}
	query = "UPDATE account SET balance = balance + $1 WHERE person = $2 AND number = $3"
	_, err = tx.Exec(ctx, query, sum, user, number)
	if err != nil {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("update account balance failed: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("update account balance failed: %w", err)
	}
	return currentBalance + sum, nil
}

func (s *Storage) TransferBetweenAccounts(ctx context.Context, user, from, to string, sum float64) (float64, error) {
	tx, err := s.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("transfer between accounts failed: %w", err)
	}
	var currentBalance float64
	query := "SELECT balance FROM account WHERE person = $1 AND number = $2"
	err = tx.QueryRow(ctx, query, user, from).Scan(&currentBalance)
	if err != nil {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("reading account balance failed: %w", err)
	}
	if currentBalance-sum < 0 {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("account balance not enough")
	}
	query = "UPDATE account SET balance = balance - $1 WHERE person = $2 AND number = $3"
	_, err = tx.Exec(ctx, query, sum, user, from)
	if err != nil {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("update sending account balance failed: %w", err)
	}
	if s.GetAccount(ctx, to) != nil {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("receiving account not exist")
	}
	query = "UPDATE account SET balance = balance + $1 WHERE number = $2"
	_, err = tx.Exec(ctx, query, sum, to)
	if err != nil {
		tx.Rollback(ctx)
		return 0, fmt.Errorf("update receiving account balance failed: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("transfer between accounts failed: %w", err)
	}
	return currentBalance - sum, nil
}

func (s *Storage) GetAccount(ctx context.Context, number string) error {
	query := "SELECT id FROM account WHERE number = $1"
	row := s.QueryRow(ctx, query, number)
	var account models.Account
	err := row.Scan(&account.ID)
	if err != nil {
		return fmt.Errorf("error getting account data: %w", err)
	}
	return nil
}
