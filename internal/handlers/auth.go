package handlers

import (
	"net/http"

	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuthHandler maneja los endpoints de autenticación
type AuthHandler struct {
	AuthService services.AuthService
	log         logger.Logger
}

// NewAuthHandler crea un nuevo AuthHandler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
		log:         logger.Get(),
	}
}

// Login godoc
// @Summary Login
// @Description Login con email y PIN - obtiene condominio_id automáticamente
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request (email y PIN)"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	// Bindir request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid login request format")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   true,
			Message: "invalid request format",
			Code:    "BAD_REQUEST",
		})
		return
	}

	// Validar email y PIN
	if req.Email == "" || req.PIN == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   true,
			Message: "email and PIN are required",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Llamar al servicio de autenticación
	// El servicio se encargará de encontrar el usuario, validar el PIN y obtener su condominio_id
	response, err := h.AuthService.Login(c.Request.Context(), &req, "")
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.log.WithField("user_id", response.UserID).Info("Login successful")
	c.JSON(http.StatusOK, response)
}

// Refresh godoc
// @Summary Refresh Token
// @Description Generar nuevo par de tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid refresh token request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   true,
			Message: "invalid request format",
			Code:    "BAD_REQUEST",
		})
		return
	}

	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   true,
			Message: "refresh_token is required",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	response, err := h.AuthService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.log.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, response)
}

// ChangePIN godoc
// @Summary Change PIN
// @Description Cambiar el PIN del usuario autenticado
// @Tags auth
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body dto.ChangePINRequest true "Change PIN request"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/change-pin [post]
func (h *AuthHandler) ChangePIN(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.log.Warn("Missing user_id in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   true,
			Message: "unauthorized",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req dto.ChangePINRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid change PIN request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   true,
			Message: "invalid request format",
			Code:    "BAD_REQUEST",
		})
		return
	}

	err := h.AuthService.ChangePIN(c.Request.Context(), userID.(string), req.CurrentPin, req.NewPin)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.log.WithField("user_id", userID).Info("PIN changed successfully")
	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "PIN changed successfully",
	})
}

// Logout godoc
// @Summary Logout
// @Description Cerrar sesión del usuario
// @Tags auth
// @Security Bearer
// @Success 200 {object} dto.SuccessResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.log.Warn("Missing user_id in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   true,
			Message: "unauthorized",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	err := h.AuthService.Logout(c.Request.Context(), userID.(string))
	if err != nil {
		h.log.WithError(err).Error("Error during logout")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   true,
			Message: "error during logout",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	h.log.WithField("user_id", userID).Info("User logged out")
	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "logged out successfully",
	})
}

// ========== Private Methods ==========

// handleAuthError maneja errores de autenticación
func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	if customErr, ok := err.(*errors.CustomError); ok {
		c.JSON(customErr.GetHTTPStatusCode(), dto.ErrorResponse{
			Error:   true,
			Message: customErr.Message,
			Code:    string(customErr.Type),
		})
		return
	}

	h.log.WithError(err).Error("Unknown error in auth handler")
	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error:   true,
		Message: "internal server error",
		Code:    "INTERNAL_ERROR",
	})
}

// Health godoc
// @Summary Health Check
// @Description Verificar que el servidor está funcionando
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "API is running",
	})
}
