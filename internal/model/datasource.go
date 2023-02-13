package model

import "time"

type ColumnUpdate struct {
	UpdateSet string
}

type Transaction struct {
	UserId        string    `json:"-"`
	AccountNumber int       `json:"account_number"`
	TransactionId string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	TransferTo    int       `json:"transfer_to"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status" validate:"required,oneof=approved rejected"`
	Type          string    `json:"type" validate:"required,oneof=credit debit"`
	Comment       string    `json:"comment"`
}

const Schema = `
		(
		    transaction_id VARCHAR(255) NOT NULL PRIMARY KEY,
    account_number INT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    transfer_to VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    status VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    comment VARCHAR(255)
		);
`
