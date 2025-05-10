package services

import (
	"BankService/internal/domain/models"
	"BankService/internal/repository"
	"context"
	"log"
	"time"
)

func RunScheduler() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			storage := repository.NewStorage()
			checkDebts(storage)
			storage.Close()
		}
	}
}

func checkDebts(storage *repository.Storage) {
	query := "SELECT * FROM credit"
	rows, err := storage.Query(context.Background(), query)
	if err != nil {
		log.Printf("error getting credit data: %w", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var credit models.Credit
		err = rows.Scan(&credit.ID, &credit.Account, &credit.Sum, &credit.Paid, &credit.Debt, &credit.PayDay)
		if err != nil {
			log.Printf("error parsing credit data: %w", err)
		}
		now := time.Now()
		if now.After(credit.PayDay) {
			if !reduceBalance(credit, storage) {
				increaseDebt(credit, storage)
			}
		}
	}
}

func reduceBalance(credit models.Credit, storage *repository.Storage) bool {
	query := "SELECT balance FROM account WHERE id = $1"
	row := storage.QueryRow(context.Background(), query, credit.Account)
	var balance float64
	err := row.Scan(&balance)
	if err != nil {
		log.Printf("error parsing account data: %w", err)
		return false
	}
	if balance >= credit.Sum/12 {
		query = "UPDATE account SET balance = balance - $1 WHERE id = $2"
		_, err = storage.Exec(context.Background(), query, credit.Sum/12, credit.Account)
		if err != nil {
			log.Printf("error reducing account balance: %w", err)
		}
		query = "UPDATE credit SET paid = paid + $1 WHERE id = $2"
		_, err = storage.Exec(context.Background(), query, credit.Sum/12, credit.ID)
		if err != nil {
			log.Printf("error increasing credit debt balance: %w", err)
		}
		return true
	}
	return false
}

func increaseDebt(credit models.Credit, storage *repository.Storage) {
	query := "UPDATE credit SET debt = debt + $1 WHERE id = $2"
	_, err := storage.Exec(context.Background(), query, credit.Sum/12+0.1*credit.Sum, credit.ID)
	if err != nil {
		log.Printf("error increasing credit debt balance: %w", err)
	}
}
