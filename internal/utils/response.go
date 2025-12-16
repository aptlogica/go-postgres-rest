package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents a standard API response format
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, data interface{}, message ...string) {
	response := StandardResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusOK, response)
}

// CreatedResponse sends a created response
func CreatedResponse(c *gin.Context, data interface{}, message ...string) {
	response := StandardResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	} else {
		response.Message = "Resource created successfully"
	}

	c.JSON(http.StatusCreated, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, code, message string, details ...string) {
	errorInfo := &ErrorInfo{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 {
		errorInfo.Details = details[0]
	}

	response := StandardResponse{
		Success: false,
		Error:   errorInfo,
	}

	c.JSON(statusCode, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, message string, details ...string) {
	ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", message, details...)
}

// NotFoundResponse sends a not found response
func NotFoundResponse(c *gin.Context, resource string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("%s not found", resource))
}

// InternalErrorResponse sends an internal server error response
func InternalErrorResponse(c *gin.Context, message string, details ...string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, details...)
}

// UnauthorizedResponse sends an unauthorized response
func UnauthorizedResponse(c *gin.Context, message ...string) {
	msg := "Unauthorized access"
	if len(message) > 0 {
		msg = message[0]
	}
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", msg)
}

// ForbiddenResponse sends a forbidden response
func ForbiddenResponse(c *gin.Context, message ...string) {
	msg := "Access forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", msg)
}

// ConflictResponse sends a conflict response
func ConflictResponse(c *gin.Context, message string, details ...string) {
	ErrorResponse(c, http.StatusConflict, "CONFLICT", message, details...)
}

// TooManyRequestsResponse sends a rate limit response
func TooManyRequestsResponse(c *gin.Context, message ...string) {
	msg := "Rate limit exceeded"
	if len(message) > 0 {
		msg = message[0]
	}
	ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", msg)
}

// PaginatedSuccessResponse sends a paginated success response
func PaginatedSuccessResponse(c *gin.Context, data interface{}, pagination PaginationParams, message ...string) {
	response := CreatePaginatedResponse(data, pagination)

	standardResponse := StandardResponse{
		Success: true,
		Data:    response.Data,
		Meta: map[string]interface{}{
			"pagination": response.Pagination,
		},
	}

	if response.Meta != nil {
		standardResponse.Meta.(map[string]interface{})["page_info"] = response.Meta
	}

	if len(message) > 0 {
		standardResponse.Message = message[0]
	}

	c.JSON(http.StatusOK, standardResponse)
}
