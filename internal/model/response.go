package model

import "time"

type CacheResponse struct {
	Status      int
	Response    string
	ContentType string
}

type GetTransaction struct {
	AccountNumber int       `json:"account_number"`
	TransactionId string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	TransferTo    int       `json:"transfer_to"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	Comment       string    `json:"comment"`
}
type PaginatedResponse struct {
	Response   []Transaction `json:"response"`
	Pagination Paginate      `json:"pagination"`
}
type Paginate struct {
	CurrentPage int `json:"current_page"`
	NextPage    int `json:"next_page"`
	TotalPage   int `json:"total_page"`
}
