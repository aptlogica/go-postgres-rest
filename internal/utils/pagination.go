package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationParams struct {
	Page   int   `json:"page"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{}      `json:"data"`
	Pagination PaginationParams `json:"pagination"`
	Meta       *ResponseMeta    `json:"meta,omitempty"`
}

type ResponseMeta struct {
	TotalPages   int  `json:"total_pages,omitempty"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
	NextPage     *int `json:"next_page,omitempty"`
	PreviousPage *int `json:"previous_page,omitempty"`
}

// GetPaginationParams extracts pagination parameters from request
func GetPaginationParams(c *gin.Context) PaginationParams {
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// CreatePaginatedResponse creates a paginated response with metadata
func CreatePaginatedResponse(data interface{}, pagination PaginationParams) PaginatedResponse {
	response := PaginatedResponse{
		Data:       data,
		Pagination: pagination,
	}

	// Add metadata if total is provided
	if pagination.Total > 0 {
		totalPages := int((pagination.Total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

		meta := &ResponseMeta{
			TotalPages:  totalPages,
			HasNext:     pagination.Page < totalPages,
			HasPrevious: pagination.Page > 1,
		}

		if meta.HasNext {
			nextPage := pagination.Page + 1
			meta.NextPage = &nextPage
		}

		if meta.HasPrevious {
			prevPage := pagination.Page - 1
			meta.PreviousPage = &prevPage
		}

		response.Meta = meta
	}

	return response
}

// CalculateTotalPages calculates total pages based on total records and limit
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(page, limit int) error {
	if page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}

	if limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}

	if limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}

	return nil
}

// GetLimitOffset extracts limit and offset from pagination params
func GetLimitOffset(params PaginationParams) (int, int) {
	return params.Limit, params.Offset
}
