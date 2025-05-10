package models

import "time"

type Credit struct {
	ID      int
	Account int
	Sum     float64
	Paid    float64
	Debt    float64
	PayDay  time.Time
}
type CreateCreditRequest struct {
	Number string  `json:"number"`
	Sum    float64 `json:"sum"`
}
