package repository

import (
	"BankService/internal/domain/models"
	"context"
	"fmt"
	"time"
)

func (s *Storage) CreateCredit(ctx context.Context, request models.CreateCreditRequest) error {
	query := "SELECT id FROM account WHERE number = $1"
	row := s.QueryRow(ctx, query, request.Number)
	var numAccount int
	err := row.Scan(&numAccount)
	if err != nil {
		return fmt.Errorf("error getting account data: %w", err)
	}
	now := time.Now()
	payDay := time.Date(now.Year(), now.Month()+1, now.Day(), 0, 0, 0, 0, &time.Location{}).Format("2006-01-02")
	query = "INSERT INTO credit (account, sum, paid, debt, pay_day) VALUES ($1, $2, 0, 0, $3)"
	_, err = s.Exec(ctx, query, numAccount, request.Sum, payDay)
	if err != nil {
		return fmt.Errorf("create credit failed: %w", err)
	}
	return nil
}
