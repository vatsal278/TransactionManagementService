package model

// UpdateTransaction is the model for updating transactions
type UpdateTransaction struct {
	AccountNumber   int     `json:"account_number" validate:"required"`
	Amount          float64 `json:"amount" validate:"required"`
	TransactionType string  `json:"transaction_type" validate:"required,oneof=debit credit"`
}

// SessionStruct is the model for user sessions
type SessionStruct struct {
	UserId string
	Cookie string
}

// NewTransaction is the model for creating new transactions
type NewTransaction struct {
	UserId        string  `json:"-"`
	AccountNumber int     `json:"account_number"`
	Amount        float64 `json:"amount"`
	TransferTo    int     `json:"transfer_to"`
	Status        string  `json:"status" validate:"required,oneof=approved rejected"`
	Type          string  `json:"type" validate:"required,oneof=credit debit"`
	Comment       string  `json:"comment"`
}
