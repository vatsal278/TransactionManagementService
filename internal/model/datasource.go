package model

import "time"

// ColumnUpdate represents a column update in a database table
type ColumnUpdate struct {
	UpdateSet string
}

// Transaction represents a single transaction for a user's account
type Transaction struct {
	UserId        string    `json:"-"` // User ID associated with the transaction (not included in JSON response)
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

// Schema represents the database schema for the transactions table
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
