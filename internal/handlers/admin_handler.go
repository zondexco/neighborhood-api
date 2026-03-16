package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/logger"
)

// AdminHandler maneja las rutas de administración
type AdminHandler struct {
	userRepository          repositories.UserRepository
	apartmentRepository     repositories.ApartmentRepository
	communicationRepository repositories.CommunicationRepository
	condominioRepository    repositories.CondominioRepository
	logger                  logger.Logger
}

// NewAdminHandler crea un nuevo manejador de admin
func NewAdminHandler(
	userRepo repositories.UserRepository,
	apartmentRepo repositories.ApartmentRepository,
	communicationRepo repositories.CommunicationRepository,
	condominioRepo repositories.CondominioRepository,
	log logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		userRepository:          userRepo,
		apartmentRepository:     apartmentRepo,
		communicationRepository: communicationRepo,
		condominioRepository:    condominioRepo,
		logger:                  log,
	}
}

// ListUsers obtiene la lista de usuarios del condominio
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	page := 1
	pageSize := 100

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 500 {
			pageSize = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 500 {
			pageSize = parsed
		}
	}

	// Obtener todos los usuarios del condominio de la base de datos
	users, total, err := h.userRepository.GetByCondominio(c.Request.Context(), condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("condominio_id", condominioID).Error("error fetching users")
		InternalServerError(c, "error fetching users")
		return
	}

	// Obtener apartamentos para mapear labels (best-effort)
	aptMap := map[string]*models.Apartment{}
	if apartments, _, aptErr := h.apartmentRepository.GetByCondominio(c.Request.Context(), condominioID.(string), 1, 1000); aptErr == nil {
		for _, apt := range apartments {
			aptMap[apt.ID] = apt
		}
	} else {
		h.logger.WithError(aptErr).WithField("condominio_id", condominioID).Warn("could not preload apartments for user listing")
	}

	formatApartment := func(a *models.Apartment) string {
		if a == nil {
			return ""
		}
		parts := []string{}
		if a.Torre != nil && *a.Torre != "" {
			parts = append(parts, "Torre "+*a.Torre)
		}
		if a.Numero != "" {
			parts = append(parts, "Apt "+a.Numero)
		}
		if len(parts) == 0 {
			return "Apartamento"
		}
		return strings.Join(parts, " · ")
	}

	formatted := make([]gin.H, 0, len(users))
	for _, u := range users {
		var aptLabel any = nil
		if u.ApartmentID != nil {
			if apt, ok := aptMap[*u.ApartmentID]; ok {
				aptLabel = formatApartment(apt)
			}
			if aptLabel == nil {
				aptLabel = *u.ApartmentID
			}
		}

		item := gin.H{
			"id":               u.ID,
			"nombres":          u.Nombre,
			"nombre":           u.Nombre,
			"apellidos":        u.Apellido,
			"apellido":         u.Apellido,
			"email":            u.Email,
			"telefono":         u.Telefono,
			"rol":              u.Rol,
			"estado":           u.Estado,
			"condominio_id":    u.CondominioID,
			"tipo_documento":   u.TipoDoc,
			"numero_documento": u.NumeroDoc,
			"apartamento_id":   u.ApartmentID,
			"apartamento":      aptLabel,
			"created_at":       u.CreatedAt,
			"updated_at":       u.UpdatedAt,
		}

		if u.Celular != nil {
			item["celular"] = *u.Celular
		}

		formatted = append(formatted, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       formatted,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// GetUserByID obtiene un usuario por ID
// GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Obtener usuario por ID de la base de datos
	user, err := h.userRepository.FindByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Warn("user not found")
		NotFound(c, "user not found")
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListApartments obtiene la lista de apartamentos del condominio
// GET /api/v1/admin/apartments
func (h *AdminHandler) ListApartments(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	page := 1
	pageSize := 100

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 500 {
			pageSize = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 500 {
			pageSize = parsed
		}
	}

	// Obtener todos los apartamentos del condominio de la base de datos
	apartments, total, err := h.apartmentRepository.GetByCondominio(c.Request.Context(), condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("condominio_id", condominioID).Error("error fetching apartments")
		InternalServerError(c, "error fetching apartments")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       apartments,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// GetApartmentByID obtiene un apartamento por ID
// GET /api/v1/admin/apartments/:id
func (h *AdminHandler) GetApartmentByID(c *gin.Context) {
	apartmentID := c.Param("id")
	if apartmentID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Obtener apartamento por ID de la base de datos
	apartment, err := h.apartmentRepository.FindByID(c.Request.Context(), apartmentID)
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Warn("apartment not found")
		NotFound(c, "apartment not found")
		return
	}

	c.JSON(http.StatusOK, apartment)
}

// GetCondominio obtiene la información del condominio
// GET /api/v1/admin/condominio
func (h *AdminHandler) GetCondominio(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Obtener información del condominio de la base de datos
	condominio, err := h.condominioRepository.FindByID(c.Request.Context(), condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("condominio_id", condominioID).Warn("condominio not found")
		NotFound(c, "condominio not found")
		return
	}

	c.JSON(http.StatusOK, condominio)
}

// GetStats obtiene estadísticas del condominio
// GET /api/v1/admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	condID := condominioID.(string)
	ctx := c.Request.Context()

	// Usar COUNT(*) directas en vez de cargar filas completas
	_, totalUsers, err := h.userRepository.GetByCondominio(ctx, condID, 1, 1)
	if err != nil {
		h.logger.WithError(err).Error("error fetching user count for stats")
		InternalServerError(c, "error fetching users")
		return
	}

	_, totalApartments, err := h.apartmentRepository.GetByCondominio(ctx, condID, 1, 1)
	if err != nil {
		h.logger.WithError(err).Error("error fetching apartment count for stats")
		InternalServerError(c, "error fetching apartments")
		return
	}

	_, totalCommunications, err := h.communicationRepository.GetByCondominio(ctx, condID, 1, 1)
	if err != nil {
		h.logger.WithError(err).Error("error fetching communication count for stats")
		InternalServerError(c, "error fetching communications")
		return
	}

	stats := map[string]interface{}{
		"total_usuarios":      totalUsers,
		"total_apartamentos":  totalApartments,
		"comunicaciones_mes":  totalCommunications,
		"facturas_pendientes": 0, // TODO: Implement when invoice data is available
		"reservas_activas":    0, // TODO: Implement when reservation data is available
	}

	c.JSON(http.StatusOK, stats)
}
