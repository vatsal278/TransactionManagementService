package model

type CacheResponse struct {
	Status      int
	Response    string
	ContentType string
}
type PaginatedResponse struct {
	Response   []Transaction
	Pagination Paginate
}

type Paginate struct {
	CurrentPage int
	NextPage    int
	TotalPage   int
}
