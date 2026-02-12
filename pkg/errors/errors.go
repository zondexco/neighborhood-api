package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType define el tipo de error
type ErrorType string

const (
	ValidationError      ErrorType = "VALIDATION_ERROR"
	NotFoundError        ErrorType = "NOT_FOUND_ERROR"
	UnauthorizedError    ErrorType = "UNAUTHORIZED_ERROR"
	ForbiddenError       ErrorType = "FORBIDDEN_ERROR"
	ConflictError        ErrorType = "CONFLICT_ERROR"
	InternalError        ErrorType = "INTERNAL_ERROR"
	BadRequestError      ErrorType = "BAD_REQUEST_ERROR"
	DatabaseError        ErrorType = "DATABASE_ERROR"
	ExternalServiceError ErrorType = "EXTERNAL_SERVICE_ERROR"

	// Errores específicos de autenticación
	InvalidEmailOrPINError ErrorType = "INVALID_EMAIL_OR_PIN_ERROR"
	UserNotActiveError     ErrorType = "USER_NOT_ACTIVE_ERROR"
)

// CustomError estructura personalizada de error
type CustomError struct {
	Type    ErrorType
	Message string
	Details map[string]interface{}
	Err     error
}

// Error implementa la interfaz error
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// New crea un nuevo error personalizado
func New(errorType ErrorType, message string) *CustomError {
	return &CustomError{
		Type:    errorType,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithError añade el error original
func (e *CustomError) WithError(err error) *CustomError {
	e.Err = err
	return e
}

// WithDetail añade detalles del error
func (e *CustomError) WithDetail(key string, value interface{}) *CustomError {
	e.Details[key] = value
	return e
}

// WithDetails añade múltiples detalles
func (e *CustomError) WithDetails(details map[string]interface{}) *CustomError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// GetHTTPStatusCode retorna el código HTTP apropiado
func (e *CustomError) GetHTTPStatusCode() int {
	switch e.Type {
	case ValidationError, BadRequestError:
		return http.StatusBadRequest
	case NotFoundError:
		return http.StatusNotFound
	case UnauthorizedError, InvalidEmailOrPINError:
		return http.StatusUnauthorized
	case ForbiddenError, UserNotActiveError:
		return http.StatusForbidden
	case ConflictError:
		return http.StatusConflict
	case DatabaseError, ExternalServiceError, InternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ErrorResponse estructura para respuestas de error
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Code    string                 `json:"code"`
}

// ToErrorResponse convierte CustomError a ErrorResponse
func (e *CustomError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Error:   "error",
		Message: e.Message,
		Details: e.Details,
		Code:    string(e.Type),
	}
}

// Is comprueba si un error es del tipo especificado
func Is(err error, errorType ErrorType) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.Type == errorType
	}
	return false
}

// Helper functions para crear errores comunes

func ValidationErrorf(message string, args ...interface{}) *CustomError {
	return New(ValidationError, fmt.Sprintf(message, args...))
}

func NotFoundErrorf(message string, args ...interface{}) *CustomError {
	return New(NotFoundError, fmt.Sprintf(message, args...))
}

func UnauthorizedErrorf(message string, args ...interface{}) *CustomError {
	return New(UnauthorizedError, fmt.Sprintf(message, args...))
}

func ForbiddenErrorf(message string, args ...interface{}) *CustomError {
	return New(ForbiddenError, fmt.Sprintf(message, args...))
}

func ConflictErrorf(message string, args ...interface{}) *CustomError {
	return New(ConflictError, fmt.Sprintf(message, args...))
}

func InternalErrorf(message string, args ...interface{}) *CustomError {
	return New(InternalError, fmt.Sprintf(message, args...))
}

func DatabaseErrorf(message string, args ...interface{}) *CustomError {
	return New(DatabaseError, fmt.Sprintf(message, args...))
}

func InvalidEmailOrPINErrorf(message string, args ...interface{}) *CustomError {
	return New(InvalidEmailOrPINError, fmt.Sprintf(message, args...))
}

func UserNotActiveErrorf(message string, args ...interface{}) *CustomError {
	return New(UserNotActiveError, fmt.Sprintf(message, args...))
}
