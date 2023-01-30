package model

type UpdateTransaction struct {
	AccountNumber   int
	Amount          int
	TransactionType string
}
type SessionStruct struct {
	UserId string
	Cookie string
}
