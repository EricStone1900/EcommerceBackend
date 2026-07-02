package errors

import "fmt"

// Error represents a business error with a code, user-facing message, and HTTP status code.
type Error struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

// Business error codes
const (
	CodeSuccess            = 0
	CodeInvalidRequest     = 1001
	CodeValidationError    = 1002
	CodeUnauthorized       = 2001
	CodeTokenExpired       = 2002
	CodeForbidden          = 2003
	CodeUserNotFound       = 3001
	CodeEmailAlreadyExists = 3002
	CodeInvalidCredentials = 3003
	CodeInternalError      = 5001
)

// Sentinel errors — each carries the correct HTTP status code and message.
var (
	ErrInvalidRequest     = &Error{Code: CodeInvalidRequest, Message: "invalid request", HTTPCode: 400}
	ErrValidationError    = &Error{Code: CodeValidationError, Message: "validation failed", HTTPCode: 400}
	ErrUnauthorized       = &Error{Code: CodeUnauthorized, Message: "unauthorized", HTTPCode: 401}
	ErrTokenExpired       = &Error{Code: CodeTokenExpired, Message: "token expired", HTTPCode: 401}
	ErrForbidden          = &Error{Code: CodeForbidden, Message: "forbidden", HTTPCode: 403}
	ErrUserNotFound       = &Error{Code: CodeUserNotFound, Message: "user not found", HTTPCode: 404}
	ErrEmailAlreadyExists = &Error{Code: CodeEmailAlreadyExists, Message: "email already registered", HTTPCode: 409}
	ErrInvalidCredentials = &Error{Code: CodeInvalidCredentials, Message: "invalid email or password", HTTPCode: 401}
	ErrInternalError      = &Error{Code: CodeInternalError, Message: "internal server error", HTTPCode: 500}
)

// NewValidationError creates a validation error with a field-specific message.
func NewValidationError(field, reason string) *Error {
	return &Error{
		Code:     CodeValidationError,
		Message:  fmt.Sprintf("validation failed: %s %s", field, reason),
		HTTPCode: 400,
	}
}
