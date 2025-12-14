package utils

// PaginationParams represents query parameters for pagination
type PaginationParams struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// PaginationMeta represents pagination metadata in responses
type PaginationMeta struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	TotalRecords int64 `json:"total_records"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

// DefaultPaginationParams returns default pagination parameters (page 1, size 10)
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 10,
	}
}

// Validate ensures pagination params are within acceptable ranges
func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// GetOffset calculates the database offset based on page and page_size
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size (for database queries)
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}

// NewPaginationMeta creates pagination metadata for responses
func NewPaginationMeta(page, pageSize int, totalRecords int64) *PaginationMeta {
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationMeta{
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
	}
}
