// Package errors provides centralized error handling for JARVIS
package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Database errors
	ErrCodeDatabase        ErrorCode = "DATABASE_ERROR"
	ErrCodeDatabaseTimeout ErrorCode = "DATABASE_TIMEOUT"

	// LLM errors
	ErrCodeLLM        ErrorCode = "LLM_ERROR"
	ErrCodeLLMTimeout ErrorCode = "LLM_TIMEOUT"
	ErrCodeLLMQuota   ErrorCode = "LLM_QUOTA_EXCEEDED"

	// Proxy errors
	ErrCodeProxy           ErrorCode = "PROXY_ERROR"
	ErrCodeProxyTimeout    ErrorCode = "PROXY_TIMEOUT"
	ErrCodeTargetUnreach   ErrorCode = "TARGET_UNREACHABLE"

	// Validation errors
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidSpec      ErrorCode = "INVALID_SPEC"
	ErrCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrCodeInvalidResponse  ErrorCode = "INVALID_RESPONSE"

	// Configuration errors
	ErrCodeConfig         ErrorCode = "CONFIG_ERROR"
	ErrCodeInvalidConfig  ErrorCode = "INVALID_CONFIG"
	ErrCodeMissingConfig  ErrorCode = "MISSING_CONFIG"

	// Certificate errors
	ErrCodeCertificate     ErrorCode = "CERTIFICATE_ERROR"
	ErrCodeCertExpired     ErrorCode = "CERTIFICATE_EXPIRED"
	ErrCodeCertInvalid     ErrorCode = "CERTIFICATE_INVALID"

	// Generic errors
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
)

// JarvisError represents a structured error with context
type JarvisError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"-"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *JarvisError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *JarvisError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *JarvisError) WithContext(key string, value interface{}) *JarvisError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// MarshalJSON implements json.Marshaler
func (e *JarvisError) MarshalJSON() ([]byte, error) {
	type Alias JarvisError
	return json.Marshal(&struct {
		*Alias
		CauseMsg string `json:"cause,omitempty"`
	}{
		Alias:    (*Alias)(e),
		CauseMsg: fmt.Sprintf("%v", e.Cause),
	})
}

// New creates a new JarvisError
func New(code ErrorCode, message string) *JarvisError {
	return &JarvisError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with a JarvisError
func Wrap(code ErrorCode, message string, cause error) *JarvisError {
	return &JarvisError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	if jarvisErr, ok := err.(*JarvisError); ok {
		return jarvisErr.Code == code
	}
	return false
}

// GetContext retrieves context from an error
func GetContext(err error, key string) (interface{}, bool) {
	if jarvisErr, ok := err.(*JarvisError); ok {
		val, exists := jarvisErr.Context[key]
		return val, exists
	}
	return nil, false
}

// Predefined error constructors for common cases

// NewDatabaseError creates a database error
func NewDatabaseError(message string, cause error) *JarvisError {
	return Wrap(ErrCodeDatabase, message, cause)
}

// NewLLMError creates an LLM error
func NewLLMError(message string, cause error) *JarvisError {
	return Wrap(ErrCodeLLM, message, cause)
}

// NewProxyError creates a proxy error
func NewProxyError(message string, cause error) *JarvisError {
	return Wrap(ErrCodeProxy, message, cause)
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *JarvisError {
	return Wrap(ErrCodeValidation, message, cause)
}

// NewConfigError creates a configuration error
func NewConfigError(message string, cause error) *JarvisError {
	return Wrap(ErrCodeConfig, message, cause)
}

// NewCertificateError creates a certificate error
func NewCertificateError(message string, cause error) *JarvisError {
	return Wrap(ErrCodeCertificate, message, cause)
}
