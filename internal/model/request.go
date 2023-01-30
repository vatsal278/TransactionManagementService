package model

type UpdateTransaction struct {
	AccountNumber   int
	Amount          float64
	TransactionType string
}
type SessionStruct struct {
	UserId string
	Cookie string
}
