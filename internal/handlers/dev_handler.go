package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/internal/utils"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"
)

// DevHandler maneja los endpoints exclusivos para desarrolladores
type DevHandler struct {
	condominioRepo repositories.CondominioRepository
	userRepo       repositories.UserRepository
	jwtManager     *utils.JWTManager
	logger         logger.Logger
}

// NewDevHandler crea un nuevo manejador de dev
func NewDevHandler(
	condominioRepo repositories.CondominioRepository,
	userRepo repositories.UserRepository,
	jwtManager *utils.JWTManager,
) *DevHandler {
	return &DevHandler{
		condominioRepo: condominioRepo,
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		logger:         logger.Get(),
	}
}

// ListCondominios lista todos los condominios del sistema
// GET /api/v1/dev/condominios
func (h *DevHandler) ListCondominios(c *gin.Context) {
	condominios, err := h.condominioRepo.GetAll(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("error listing condominios")
		InternalServerError(c, "error listing condominios")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  condominios,
		"total": len(condominios),
	})
}

// CreateCondominio crea un nuevo condominio
// POST /api/v1/dev/condominios
func (h *DevHandler) CreateCondominio(c *gin.Context) {
	var req dto.CreateCondominioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body for create condominio")
		BadRequest(c, err.Error())
		return
	}

	condominio := &models.Condominio{
		Nombre:             req.Nombre,
		Direccion:          req.Direccion,
		Ciudad:             req.Ciudad,
		Telefono:           req.Telefono,
		Email:              req.Email,
		NIT:                req.NIT,
		RepresentanteLegal: req.RepresentanteLegal,
		Estado:             "activo",
		PermiteSoporte:     false,
	}

	if err := h.condominioRepo.Create(c.Request.Context(), condominio); err != nil {
		h.logger.WithError(err).Error("error creating condominio")
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			Conflict(c, "condominio already exists")
			return
		}
		InternalServerError(c, "error creating condominio")
		return
	}

	c.JSON(http.StatusCreated, condominio)
}

// UpdateCondominio actualiza un condominio (dev puede editar cualquiera)
// PUT /api/v1/dev/condominios/:id
func (h *DevHandler) UpdateCondominio(c *gin.Context) {
	condominioID := c.Param("id")
	if condominioID == "" {
		BadRequest(c, "invalid request")
		return
	}

	condominio, err := h.condominioRepo.FindByID(c.Request.Context(), condominioID)
	if err != nil || condominio == nil {
		NotFound(c, "condominio not found")
		return
	}

	var req struct {
		Nombre             *string `json:"nombre"`
		Direccion          *string `json:"direccion"`
		Ciudad             *string `json:"ciudad"`
		Telefono           *string `json:"telefono"`
		Email              *string `json:"email"`
		NIT                *string `json:"nit"`
		RepresentanteLegal *string `json:"representante_legal"`
		PermiteSoporte     *bool   `json:"permite_soporte"`
		Estado             *string `json:"estado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if req.Nombre != nil {
		condominio.Nombre = *req.Nombre
	}
	if req.Direccion != nil {
		condominio.Direccion = *req.Direccion
	}
	if req.Ciudad != nil {
		condominio.Ciudad = *req.Ciudad
	}
	if req.Telefono != nil {
		condominio.Telefono = req.Telefono
	}
	if req.Email != nil {
		condominio.Email = req.Email
	}
	if req.NIT != nil {
		condominio.NIT = req.NIT
	}
	if req.RepresentanteLegal != nil {
		condominio.RepresentanteLegal = req.RepresentanteLegal
	}
	if req.PermiteSoporte != nil {
		condominio.PermiteSoporte = *req.PermiteSoporte
	}
	if req.Estado != nil {
		condominio.Estado = *req.Estado
	}

	if err := h.condominioRepo.Update(c.Request.Context(), condominio); err != nil {
		h.logger.WithError(err).Error("error updating condominio")
		InternalServerError(c, "error updating condominio")
		return
	}

	c.JSON(http.StatusOK, condominio)
}

// DeleteCondominio elimina un condominio
// DELETE /api/v1/dev/condominios/:id
func (h *DevHandler) DeleteCondominio(c *gin.Context) {
	condominioID := c.Param("id")
	if condominioID == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.condominioRepo.Delete(c.Request.Context(), condominioID); err != nil {
		h.logger.WithError(err).Error("error deleting condominio")
		InternalServerError(c, "error deleting condominio")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListAdminsByCondominio lista los administradores de un condominio
// GET /api/v1/dev/condominios/:id/admins
func (h *DevHandler) ListAdminsByCondominio(c *gin.Context) {
	condominioID := c.Param("id")
	if condominioID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Fetch all users for this condominio then filter admins
	users, _, err := h.userRepo.GetByCondominio(c.Request.Context(), condominioID, 1, 500)
	if err != nil {
		h.logger.WithError(err).Error("error listing users for condominio")
		InternalServerError(c, "error listing admins")
		return
	}

	admins := make([]*models.User, 0)
	for _, u := range users {
		role := strings.ToLower(u.Rol)
		if role == "administrador" || role == "admin" {
			admins = append(admins, u)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  admins,
		"total": len(admins),
	})
}

// Impersonate genera un token para impersonar un condominio (dev only)
// POST /api/v1/dev/impersonate
func (h *DevHandler) Impersonate(c *gin.Context) {
	callerID, _ := c.Get("user_id")

	var req dto.ImpersonateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Verify condominio exists and has soporte enabled
	condominio, err := h.condominioRepo.FindByID(c.Request.Context(), req.CondominioID)
	if err != nil || condominio == nil {
		NotFound(c, "condominio not found")
		return
	}

	if !condominio.PermiteSoporte {
		Forbidden(c, "el condominio no tiene habilitado el acceso de soporte")
		return
	}

	// Get the dev user
	devUser, err := h.userRepo.FindByID(c.Request.Context(), callerID.(string))
	if err != nil {
		h.logger.WithError(err).Error("error finding dev user")
		InternalServerError(c, "error finding user")
		return
	}

	// Generate token with the target condominio_id but keeping dev role
	token, err := h.jwtManager.GenerateToken(devUser.ID, devUser.Email, req.CondominioID, "dev", "")
	if err != nil {
		h.logger.WithError(err).Error("error generating impersonate token")
		InternalServerError(c, "error generating token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(devUser.ID)
	if err != nil {
		h.logger.WithError(err).Error("error generating impersonate refresh token")
		InternalServerError(c, "error generating refresh token")
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.jwtManager.GetExpiration()) * time.Second)

	h.logger.WithField("dev_user", devUser.ID).WithField("target_condominio", req.CondominioID).Info("Dev impersonating condominio")

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:          token,
		RefreshToken:   refreshToken,
		UserID:         devUser.ID,
		Email:          devUser.Email,
		Nombre:         devUser.Nombre,
		Apellido:       devUser.Apellido,
		CondominioID:   req.CondominioID,
		CondominioName: condominio.Nombre,
		Role:           "dev",
		IsAdmin:        true,
		ApartmentID:    "",
		Permissions:    []string{},
		ExpiresAt:      expiresAt,
	})
}
