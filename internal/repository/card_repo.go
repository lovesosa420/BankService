package repository

import (
	"BankService/internal/domain/models"
	"context"
	"fmt"
)

func (s *Storage) CreateCard(ctx context.Context, numAccount, numCard, date, cvv string) error {
	query := "SELECT id FROM account WHERE number = $1"
	row := s.QueryRow(ctx, query, numAccount)
	var idAccount int
	err := row.Scan(&idAccount)
	if err != nil {
		return fmt.Errorf("error getting account data: %w", err)
	}
	query = "INSERT INTO card (account, number, date, cvv) VALUES ($1, $2, $3, $4)"
	_, err = s.Exec(ctx, query, idAccount, numCard, date, cvv)
	if err != nil {
		return fmt.Errorf("create card failed: %w", err)
	}
	return nil
}

func (s *Storage) GetCards(ctx context.Context, numAccount string) ([]models.GetCardResponse, error) {
	query := "SELECT id FROM account WHERE number = $1"
	row := s.QueryRow(ctx, query, numAccount)
	var idAccount int
	err := row.Scan(&idAccount)
	if err != nil {
		return nil, fmt.Errorf("error getting account data: %w", err)
	}
	query = "SELECT number, date, cvv FROM card WHERE account = $1"
	rows, err := s.Query(ctx, query, idAccount)
	if err != nil {
		return nil, fmt.Errorf("error getting card data: %w", err)
	}
	defer rows.Close()
	var cardlist []models.GetCardResponse
	for rows.Next() {
		card := models.GetCardResponse{}
		err = rows.Scan(&card.Number, &card.Date, &card.CVV)
		if err != nil {
			return nil, fmt.Errorf("error getting card data: %w", err)
		}
		cardlist = append(cardlist, card)
	}
	return cardlist, nil
}
