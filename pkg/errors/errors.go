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
	CodeProductNotFound    = 4004
	CodeFileTooLarge       = 6001
	CodeInvalidFileType    = 6002
	CodeFileNotFound       = 6003
	CodeFileUploadFailed   = 6004
	CodePushTokenNotFound  = 7001
	CodePushSendFailed     = 7002
	CodeInvalidPlatform    = 7003
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
	ErrProductNotFound    = &Error{Code: CodeProductNotFound, Message: "product not found", HTTPCode: 404}
	ErrFileTooLarge       = &Error{Code: CodeFileTooLarge, Message: "file size exceeds limit", HTTPCode: 413}
	ErrInvalidFileType    = &Error{Code: CodeInvalidFileType, Message: "invalid file type or extension", HTTPCode: 400}
	ErrFileNotFound       = &Error{Code: CodeFileNotFound, Message: "file not found", HTTPCode: 404}
	ErrFileUploadFailed   = &Error{Code: CodeFileUploadFailed, Message: "file upload failed", HTTPCode: 500}
	ErrPushTokenNotFound  = &Error{Code: CodePushTokenNotFound, Message: "push token not found", HTTPCode: 404}
	ErrPushSendFailed     = &Error{Code: CodePushSendFailed, Message: "push send failed", HTTPCode: 500}
	ErrInvalidPlatform    = &Error{Code: CodeInvalidPlatform, Message: "invalid platform, only 'ios' is supported", HTTPCode: 400}
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
