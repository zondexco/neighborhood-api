package handlers

import (
	"net/http"

	"neighborhood-api/pkg/dto"

	"github.com/gin-gonic/gin"
)

// ErrorResponse envía una respuesta de error estándar
func ErrorResponse(c *gin.Context, statusCode int, message, code string) {
	c.JSON(statusCode, dto.ErrorResponse{
		Error:   true,
		Message: message,
		Code:    code,
	})
}

// BadRequest envía error 400
func BadRequest(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message, "BAD_REQUEST_ERROR")
}

// Unauthorized envía error 401
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, "UNAUTHORIZED_ERROR")
}

// NotFound envía error 404
func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message, "NOT_FOUND_ERROR")
}

// InternalServerError envía error 500
func InternalServerError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message, "INTERNAL_ERROR")
}

// Conflict envía error 409
func Conflict(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusConflict, message, "CONFLICT_ERROR")
}

// Forbidden envía error 403
func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, "FORBIDDEN_ERROR")
}

// InvalidEmailOrPIN envía error 401 para credenciales inválidas
func InvalidEmailOrPIN(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, "INVALID_EMAIL_OR_PIN_ERROR")
}

// UserNotActive envía error 403 cuando el usuario no está activo
func UserNotActive(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, "USER_NOT_ACTIVE_ERROR")
}
