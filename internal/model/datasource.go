package model

type DsResponse struct {
	Data string
}

type PingDs struct {
	Data string
}

type ColumnUpdate struct {
	UpdateSet string
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
