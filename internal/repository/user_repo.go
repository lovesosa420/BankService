package repository

import (
	"BankService/internal/domain/models"
	"context"
	"errors"
	"fmt"
)

func (s *Storage) RegisterUser(ctx context.Context, user models.UserRegisterRequest) error {
	if _, err := s.GetUserData(ctx, user.Email); err == nil {
		return errors.New("user with such email already exists")
	}
	query := "INSERT INTO person (login, name, password) VALUES ($1, $2, $3)"
	_, err := s.Exec(ctx, query, user.Email, user.Name, user.Password)
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	return nil
}

func (s *Storage) GetUserData(ctx context.Context, login string) (models.User, error) {
	query := "SELECT * FROM person WHERE login = $1"
	row := s.QueryRow(ctx, query, login)
	var data models.User
	err := row.Scan(&data.Email, &data.Name, &data.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("error getting user data: %w", err)
	}
	return data, nil
}
