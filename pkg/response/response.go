package response

// Response is the unified JSON response format for all API endpoints.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success creates a success response with the given data.
func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Error creates an error response with the given business code and message.
func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// PaginatedData wraps a paginated list response.
type PaginatedData struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}
