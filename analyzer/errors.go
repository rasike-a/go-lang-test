package analyzer

import (
	"fmt"
	"time"
)

// Error codes for different types of errors
const (
	ErrCodeInvalidURL        = "INVALID_URL"
	ErrCodeHTTPError         = "HTTP_ERROR"
	ErrCodeNetworkError      = "NETWORK_ERROR"
	ErrCodeParseError        = "PARSE_ERROR"
	ErrCodeTimeoutError      = "TIMEOUT_ERROR"
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodeInternalError     = "INTERNAL_ERROR"
)

// AnalysisError represents a structured error with additional context
type AnalysisError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	URL       string    `json:"url,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code,omitempty"`
	Cause     error     `json:"-"`
}

// Error implements the error interface
func (e *AnalysisError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AnalysisError) Unwrap() error {
	return e.Cause
}

// NewAnalysisError creates a new AnalysisError
func NewAnalysisError(code, message string) *AnalysisError {
	return &AnalysisError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// WithDetails adds details to the error
func (e *AnalysisError) WithDetails(details string) *AnalysisError {
	e.Details = details
	return e
}

// WithURL adds URL context to the error
func (e *AnalysisError) WithURL(url string) *AnalysisError {
	e.URL = url
	return e
}

// WithStatusCode adds HTTP status code to the error
func (e *AnalysisError) WithStatusCode(statusCode int) *AnalysisError {
	e.StatusCode = statusCode
	return e
}

// WithCause adds the underlying error cause
func (e *AnalysisError) WithCause(cause error) *AnalysisError {
	e.Cause = cause
	return e
}

// IsAnalysisError checks if an error is an AnalysisError
func IsAnalysisError(err error) bool {
	_, ok := err.(*AnalysisError)
	return ok
}

// GetAnalysisError extracts AnalysisError from an error
func GetAnalysisError(err error) *AnalysisError {
	if ae, ok := err.(*AnalysisError); ok {
		return ae
	}
	return nil
}

// Common error constructors
func NewInvalidURLError(url string, cause error) *AnalysisError {
	return NewAnalysisError(ErrCodeInvalidURL, "Invalid URL format").
		WithURL(url).
		WithCause(cause)
}

func NewHTTPError(statusCode int, url string) *AnalysisError {
	return NewAnalysisError(ErrCodeHTTPError, fmt.Sprintf("HTTP %d: %s", statusCode, httpStatusText(statusCode))).
		WithURL(url).
		WithStatusCode(statusCode)
}

func NewNetworkError(url string, cause error) *AnalysisError {
	return NewAnalysisError(ErrCodeNetworkError, "Failed to fetch URL").
		WithURL(url).
		WithCause(cause)
}

func NewParseError(url string, cause error) *AnalysisError {
	return NewAnalysisError(ErrCodeParseError, "Failed to parse HTML content").
		WithURL(url).
		WithCause(cause)
}

func NewTimeoutError(url string, timeout time.Duration) *AnalysisError {
	return NewAnalysisError(ErrCodeTimeoutError, fmt.Sprintf("Request timed out after %v", timeout)).
		WithURL(url)
}

// httpStatusText returns HTTP status text (simplified version)
func httpStatusText(statusCode int) string {
	switch statusCode {
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return "Unknown"
	}
}
