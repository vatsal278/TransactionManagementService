package model

// CacheResponse is the structure for the response of the cache middleware
type CacheResponse struct {
	Status      int    // Status code of the cached response
	Response    string // Response body of the cached response
	ContentType string // Content type of the cached response
}

// PaginatedResponse is the structure for the paginated response of transactions
type PaginatedResponse struct {
	Response   []Transaction // List of transactions for the current page
	Pagination Paginate      // Pagination information for the paginated response
}

// Paginate is the structure for pagination information
type Paginate struct {
	CurrentPage int `json:"current_page"` // The current page number
	NextPage    int `json:"next_page"`    // The next page number
	TotalPage   int `json:"total_page"`   // The total number of pages
}
