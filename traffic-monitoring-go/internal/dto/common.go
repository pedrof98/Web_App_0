package dto

// Success is a generic response envelope for successful operations
type Success[T any] struct {
	Data	T		`json:"data"`
	Meta	*MetaInfo	`json:"meta,omitempty"`
}


// error is a standardized error response format
type Error struct {
	Error struct {
		Code	string			`json:"code"`
		Message	string			`json:"message"`
		Fields	map[string]string	`json:"fields,omitempty"`
	} `json:"error"`
}

// MetaInfo contains additional metadata about a response
type MetaInfo struct {
	// Pagination info
	Page		int	`json:"page,omitempty"`
	PageSize	int	`json:"page_size,omitempty"`
	TotalItems	int64	`json:"total_items,omitempty"`
	TotalPages	int	`json:"total_pages,omitempty"`
}


// PaginationQuery represents common pagination parameters
type PaginationQuery struct {
	Page		int	`form:"page" binding:"omitempty,min=1" default:"1"`
	PageSize	int	`form:"page_size" binding:"omitempty,min=1,max=100" default:"50"`
}


// CalculationPagination calculates pagination values
func CalculatePagination(page, pageSize int, total int64) *MetaInfo {
	totalPages := (int(total) + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &MetaInfo{
		Page:		page,
		PageSize:	pageSize,
		TotalItems:	total,
		TotalPages:	totalPages,
	}
}


