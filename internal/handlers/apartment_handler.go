package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/repositories"
	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"
)

// ApartmentHandler maneja las rutas de apartamentos
type ApartmentHandler struct {
	service  services.ApartmentService
	userRepo repositories.UserRepository
	logger   logger.Logger
}

// NewApartmentHandler crea un nuevo manejador
func NewApartmentHandler(service services.ApartmentService, userRepo repositories.UserRepository, log logger.Logger) *ApartmentHandler {
	return &ApartmentHandler{
		service:  service,
		userRepo: userRepo,
		logger:   log,
	}
}

// Create crea un nuevo apartamento
// POST /api/v1/apartments
func (h *ApartmentHandler) Create(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Parsear request
	var req dto.ApartmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	// Crear apartamento
	apartment, err := h.service.Create(c.Request.Context(), &req, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).Error("error creating apartment")
		InternalServerError(c, "error creating apartment")
		return
	}

	c.JSON(http.StatusCreated, apartment)
}

// GetByID obtiene un apartamento por ID
// GET /api/v1/apartments/:id
func (h *ApartmentHandler) GetByID(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "invalid request")
		return
	}

	apartment, err := h.service.GetByID(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", id).Warn("apartment not found")
		NotFound(c, "apartment not found")
		return
	}

	c.JSON(http.StatusOK, apartment)
}

// List obtiene la lista de apartamentos con paginación
// GET /api/v1/apartments?page=1&page_size=10
func (h *ApartmentHandler) List(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Parsear paginación
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	// Empleados y admins ven todos los apartamentos del condominio
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr == "empleado" || roleStr == "administrador" || roleStr == "admin" || roleStr == "dev" {
		userIDStr = ""
	}

	result, err := h.service.List(c.Request.Context(), condominioID.(string), userIDStr, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("error listing apartments")
		InternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update actualiza un apartamento
// PUT /api/v1/apartments/:id
func (h *ApartmentHandler) Update(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Parsear request
	var req dto.ApartmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	apartment, err := h.service.Update(c.Request.Context(), id, &req, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", id).Warn("error updating apartment")
		NotFound(c, "apartment not found")
		return
	}

	c.JSON(http.StatusOK, apartment)
}

// Delete elimina un apartamento
// DELETE /api/v1/apartments/:id
func (h *ApartmentHandler) Delete(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "invalid request")
		return
	}

	err := h.service.Delete(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", id).Warn("error deleting apartment")
		NotFound(c, "apartment not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "apartment deleted successfully"})
}

// ListMembers obtiene los miembros (usuarios) de un apartamento
// GET /api/v1/apartments/:id/members
func (h *ApartmentHandler) ListMembers(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	apartmentID := c.Param("id")
	if apartmentID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Verificar que el apartamento pertenece al condominio
	_, err := h.service.GetByID(c.Request.Context(), apartmentID, condominioID.(string))
	if err != nil {
		NotFound(c, "apartment not found")
		return
	}

	users, err := h.userRepo.GetByApartment(c.Request.Context(), apartmentID)
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error listing apartment members")
		InternalServerError(c, "internal server error")
		return
	}

	type MemberDTO struct {
		ID       string `json:"id"`
		Nombre   string `json:"nombre"`
		Apellido string `json:"apellido"`
		Email    string `json:"email"`
		Rol      string `json:"rol"`
	}

	members := make([]MemberDTO, len(users))
	for i, u := range users {
		members[i] = MemberDTO{
			ID:       u.ID,
			Nombre:   u.Nombre,
			Apellido: u.Apellido,
			Email:    u.Email,
			Rol:      u.Rol,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": members, "total": len(members)})
}
