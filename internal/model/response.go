package model

import "time"

type CacheResponse struct {
	Status      int
	Response    string
	ContentType string
}
type PaginatedResponse struct {
	Response   []GetTransaction
	Pagination Paginate
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
type Paginate struct {
	CurrentPage int
	NextPage    int
	TotalPage   int
}
