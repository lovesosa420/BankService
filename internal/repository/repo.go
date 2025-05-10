package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	*pgxpool.Pool
}

func NewStorage() *Storage {
	connString := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	return &Storage{pool}
}

func (s *Storage) Close() {
	s.Pool.Close()
}
