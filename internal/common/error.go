package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// AppError represents a custom application error
type AppError struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	HTTPCode  int           `json:"-"`
}

// ErrorDetail represents individual error details
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// StandardErrorResponse represents the complete error response
type StandardErrorResponse struct {
	Success   bool          `json:"success"`
	Error     ErrorResponse `json:"error"`
	Timestamp time.Time     `json:"timestamp"`
}

// ErrorResponse represents the error part of the response
type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s", e.Code, e.Message)
}

// ToJSON converts AppError to JSON string
func (e *AppError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// NewAppError creates a new AppError
func NewAppError(code, message string, httpCode int) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		HTTPCode:  httpCode,
		Timestamp: time.Now(),
		Details:   make([]ErrorDetail, 0),
	}
}

// WithDetail adds a single error detail
func (e *AppError) WithDetail(field, message string) *AppError {
	e.Details = append(e.Details, ErrorDetail{
		Field:   field,
		Message: message,
	})
	return e
}

// WithDetails adds multiple error details
func (e *AppError) WithDetails(details []ErrorDetail) *AppError {
	e.Details = append(e.Details, details...)
	return e
}

// Common error constructors
func ErrInvalidInput(message string) *AppError {
	return NewAppError("INVALID_INPUT", message, http.StatusBadRequest)
}

func ErrNotFound(resource string) *AppError {
	return NewAppError("NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func ErrUnauthorized(message string) *AppError {
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized)
}

func ErrForbidden(message string) *AppError {
	return NewAppError("FORBIDDEN", message, http.StatusForbidden)
}

func ErrInternalServer(message string) *AppError {
	return NewAppError("INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError)
}

func ErrConflict(message string) *AppError {
	return NewAppError("CONFLICT", message, http.StatusConflict)
}

func ErrValidation(field, message string) *AppError {
	return NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest).
		WithDetail(field, message)
}

// NewStandardErrorResponse creates a new standard error response
func NewStandardErrorResponse(err *AppError) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Error: ErrorResponse{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
		Timestamp: err.Timestamp,
	}
}

// IsAppError checks if error is an AppError
func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// WrapError wraps a standard error into AppError
func WrapError(err error, code, message string, httpCode int) *AppError {
	appErr := NewAppError(code, message, httpCode)
	appErr.WithDetail("original_error", err.Error())
	return appErr
}

// HTTP Error Handler - untuk digunakan di HTTP handlers
func HandleError(w http.ResponseWriter, err error) {
	var appErr *AppError

	if customErr, ok := IsAppError(err); ok {
		appErr = customErr
	} else {
		// Wrap unknown errors as internal server error
		appErr = WrapError(err, "INTERNAL_SERVER_ERROR", "An unexpected error occurred", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPCode)

	response := NewStandardErrorResponse(appErr)
	json.NewEncoder(w).Encode(response)
}

// Validation helpers
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

func ErrMultipleValidation(errors []ErrorDetail) *AppError {
	appErr := NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest)
	appErr.WithDetails(errors)
	return appErr
}

// Validator instance
var Validate *validator.Validate

func init() {
	Validate = validator.New()
	// Register custom tag name function to use JSON tags
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom tag slug validation
	Validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slug := fl.Field().String()
		// Slug hanya lowercase, angka, dan tanda hubung (tanpa spasi/simbol)
		matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
		return matched
	})
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) *AppError {
	logrus.Infof("Validating struct: %T", s)
	err := Validate.Struct(s)
	if err == nil {
		return nil
	}

	logrus.Errorf("Validation error: %v", err)

	var errorDetails []ErrorDetail

	// Handle validator.ValidationErrors
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			errorDetails = append(errorDetails, ErrorDetail{
				Field:   fieldError.Field(),
				Message: getValidationMessage(fieldError),
			})
		}
	} else {
		// Handle other validation errors
		errorDetails = append(errorDetails, ErrorDetail{
			Field:   "unknown",
			Message: err.Error(),
		})
	}

	logrus.Errorf("Validation errors: %+v", errorDetails)
	return ErrMultipleValidation(errorDetails)
}

// getValidationMessage returns user-friendly validation messages
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", fe.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", fe.Field())
	case "numeric":
		return fmt.Sprintf("%s must contain only numeric characters", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fe.Field(), fe.Param())
	case "unique":
		return fmt.Sprintf("%s must contain unique values", fe.Field())
	case "dive":
		return fmt.Sprintf("%s contains invalid nested values", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Custom validation functions
func RegisterCustomValidation(tag string, fn validator.Func) error {
	return Validate.RegisterValidation(tag, fn)
}

// Common custom validators
func init() {
	// Register custom password validation
	Validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		// At least 8 characters, contains uppercase, lowercase, number
		if len(password) < 8 {
			return false
		}

		hasUpper := false
		hasLower := false
		hasNumber := false

		for _, char := range password {
			switch {
			case 'A' <= char && char <= 'Z':
				hasUpper = true
			case 'a' <= char && char <= 'z':
				hasLower = true
			case '0' <= char && char <= '9':
				hasNumber = true
			}
		}

		return hasUpper && hasLower && hasNumber
	})

	// Register phone number validation (Indonesian format)
	Validate.RegisterValidation("phone_id", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Simple Indonesian phone validation: starts with +62, 08, or 62
		if strings.HasPrefix(phone, "+62") || strings.HasPrefix(phone, "62") || strings.HasPrefix(phone, "08") {
			return len(phone) >= 10 && len(phone) <= 15
		}
		return false
	})
}
