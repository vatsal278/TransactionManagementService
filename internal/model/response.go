package model

type CacheResponse struct {
	Status      int
	Response    string
	ContentType string
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
