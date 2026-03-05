package handlers

import (
	"crypto/rand"
	"fmt"
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

// CreateAdminForCondominio crea un administrador en un condominio específico
// POST /api/v1/dev/condominios/:id/admins
func (h *DevHandler) CreateAdminForCondominio(c *gin.Context) {
	condominioID := c.Param("id")
	if condominioID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Verify condominio exists
	condo, err := h.condominioRepo.FindByID(c.Request.Context(), condominioID)
	if err != nil || condo == nil {
		NotFound(c, "condominio not found")
		return
	}

	var req struct {
		Nombres   string  `json:"nombres" binding:"required"`
		Apellidos string  `json:"apellidos" binding:"required"`
		Email     string  `json:"email" binding:"required"`
		Telefono  *string `json:"telefono"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Generate a random 6-digit PIN
	pinPlain, err := devGeneratePIN()
	if err != nil {
		h.logger.WithError(err).Error("failed generating pin")
		InternalServerError(c, "failed generating PIN")
		return
	}

	pinHash, err := utils.HashPIN(pinPlain)
	if err != nil {
		h.logger.WithError(err).Error("failed hashing pin")
		InternalServerError(c, "failed generating PIN")
		return
	}

	telefono := ""
	if req.Telefono != nil {
		telefono = *req.Telefono
	}

	user := &models.User{
		Nombre:       req.Nombres,
		Apellido:     req.Apellidos,
		Email:        req.Email,
		Telefono:     telefono,
		CondominioID: condominioID,
		PIN:          pinHash,
		Rol:          "administrador",
		Estado:       "activo",
		TipoDoc:      "",
		NumeroDoc:    "",
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		h.logger.WithError(err).Error("error creating admin for condominio")
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			Conflict(c, "ya existe un usuario con ese email")
			return
		}
		InternalServerError(c, "error creating admin")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"pin_temporal": pinPlain,
	})
}

// UpdateAdminInCondominio actualiza un administrador de un condominio
// PUT /api/v1/dev/condominios/:id/admins/:userId
func (h *DevHandler) UpdateAdminInCondominio(c *gin.Context) {
	condominioID := c.Param("id")
	userID := c.Param("userId")
	if condominioID == "" || userID == "" {
		BadRequest(c, "invalid request")
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		NotFound(c, "user not found")
		return
	}

	// Ensure user belongs to this condominio
	if user.CondominioID != condominioID {
		Forbidden(c, "el usuario no pertenece a este condominio")
		return
	}

	var req struct {
		Nombres   *string `json:"nombres"`
		Apellidos *string `json:"apellidos"`
		Email     *string `json:"email"`
		Telefono  *string `json:"telefono"`
		Estado    *string `json:"estado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if req.Nombres != nil {
		user.Nombre = *req.Nombres
	}
	if req.Apellidos != nil {
		user.Apellido = *req.Apellidos
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Telefono != nil {
		user.Telefono = *req.Telefono
	}
	if req.Estado != nil {
		estado := strings.TrimSpace(strings.ToLower(*req.Estado))
		switch estado {
		case "activo", "inactivo", "suspendido":
			user.Estado = estado
		default:
			user.Estado = "activo"
		}
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		h.logger.WithError(err).Error("error updating admin")
		InternalServerError(c, "error updating admin")
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteAdminFromCondominio elimina un administrador de un condominio
// DELETE /api/v1/dev/condominios/:id/admins/:userId
func (h *DevHandler) DeleteAdminFromCondominio(c *gin.Context) {
	condominioID := c.Param("id")
	userID := c.Param("userId")
	if condominioID == "" || userID == "" {
		BadRequest(c, "invalid request")
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		NotFound(c, "user not found")
		return
	}

	if user.CondominioID != condominioID {
		Forbidden(c, "el usuario no pertenece a este condominio")
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), userID); err != nil {
		h.logger.WithError(err).Error("error deleting admin")
		InternalServerError(c, "error deleting admin")
		return
	}

	c.Status(http.StatusNoContent)
}

func devGeneratePIN() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	num := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return fmt.Sprintf("%06d", num%1000000), nil
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
