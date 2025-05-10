package models

import "time"

type Card struct {
	Account int
	Number  string `json:"number"`
	Date    string `json:"date"`
	CVV     string `json:"cvv"`
}

type CardRequest struct {
	Number string `json:"number"`
}
type CreateCardResponse struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	CVV    string `json:"cvv"`
}

type GetCardResponse struct {
	Number string    `json:"number"`
	Date   time.Time `json:"date"`
	CVV    string    `json:"cvv"`
}
