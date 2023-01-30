package model

type CacheResponse struct {
	Status      int
	Response    string
	ContentType string
}
type PaginatedResponse struct {
	Response   []GetTransaction
	Pagination Paginate
}

type Paginate struct {
	CurrentPage int
	Limit       int
	TotalCount  int
	TotalPage   int
}
