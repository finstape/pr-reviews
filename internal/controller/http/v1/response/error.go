package response

import "github.com/finstape/pr-reviews/internal/entity"

// ErrorResponse -.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail -.
type ErrorDetail struct {
	Code    entity.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

// NewErrorResponse -.
func NewErrorResponse(code entity.ErrorCode, message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}

