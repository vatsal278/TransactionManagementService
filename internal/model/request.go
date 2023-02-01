package model

import "time"

type UpdateTransaction struct {
	AccountNumber   int     `json:"account_number" validate:"required"`
	Amount          float64 `json:"amount" validate:"required"`
	TransactionType string  `json:"transaction_type" validate:"required,oneof=debit credit"`
}
type SessionStruct struct {
	UserId string
	Cookie string
}
type UserDetails struct {
	Name      string
	Email     string
	Company   string
	LastLogin time.Time
}
