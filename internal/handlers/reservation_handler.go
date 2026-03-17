package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
	pkgerrors "neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"
)

// ReservationHandler maneja las rutas de reservas
type ReservationHandler struct {
	service      services.ReservationService
	notifService services.NotificationService
	logger       logger.Logger
}

// NewReservationHandler crea un nuevo manejador de reservas
func NewReservationHandler(service services.ReservationService, notifService services.NotificationService, log logger.Logger) *ReservationHandler {
	return &ReservationHandler{
		service:      service,
		notifService: notifService,
		logger:       log,
	}
}

// Create crea una nueva reserva
// POST /api/v1/reservations
func (h *ReservationHandler) Create(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Extraer usuario_id del JWT
	usuarioID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("missing user_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Parsear request
	var req dto.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	// Crear reserva
	reservation, err := h.service.Create(c.Request.Context(), &req, condominioID.(string), usuarioID.(string))
	if err != nil {
		var customErr *pkgerrors.CustomError
		if errors.As(err, &customErr) {
			c.JSON(customErr.GetHTTPStatusCode(), gin.H{"message": customErr.Message, "code": string(customErr.Type)})
		} else {
			h.logger.WithError(err).Error("error creating reservation")
			InternalServerError(c, "internal server error")
		}
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// GetByID obtiene una reserva por ID
// GET /api/v1/reservations/:id
func (h *ReservationHandler) GetByID(c *gin.Context) {
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

	reservation, err := h.service.GetByID(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("reservation_id", id).Warn("reservation not found")
		NotFound(c, "not found")
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// List obtiene la lista de reservas con paginación
// GET /api/v1/reservations?page=1&page_size=10
func (h *ReservationHandler) List(c *gin.Context) {
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
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	// Listar reservas
	reservations, total, err := h.service.List(c.Request.Context(), condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("error listing reservations")
		InternalServerError(c, "internal server error")
		return
	}

	response := dto.ReservationsListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     reservations,
	}

	c.JSON(http.StatusOK, response)
}

// ListByUsuario obtiene reservas del usuario actual
// GET /api/v1/users/:user_id/reservations?page=1&page_size=10
func (h *ReservationHandler) ListByUsuario(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	usuarioID := c.Param("user_id")
	if usuarioID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Parsear paginación
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	// Listar reservas del usuario
	reservations, total, err := h.service.ListByUsuario(c.Request.Context(), usuarioID, condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", usuarioID).Error("error listing reservations")
		InternalServerError(c, "internal server error")
		return
	}

	response := dto.ReservationsListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     reservations,
	}

	c.JSON(http.StatusOK, response)
}

// ListByEspacio obtiene reservas de un espacio común
// GET /api/v1/espacios/:espacio_id/reservations?page=1&page_size=10
func (h *ReservationHandler) ListByEspacio(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	espacioID := c.Param("espacio_id")
	if espacioID == "" {
		BadRequest(c, "invalid request")
		return
	}

	// Parsear paginación
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	// Listar reservas del espacio
	reservations, total, err := h.service.ListByEspacio(c.Request.Context(), espacioID, condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("espacio_id", espacioID).Error("error listing reservations")
		InternalServerError(c, "internal server error")
		return
	}

	response := dto.ReservationsListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     reservations,
	}

	c.JSON(http.StatusOK, response)
}

// ListByApartment obtiene reservas de todos los miembros de un apartamento
// GET /api/v1/apartments/:id/reservations
func (h *ReservationHandler) ListByApartment(c *gin.Context) {
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

	page := 1
	pageSize := 50

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	reservations, total, err := h.service.ListByApartment(c.Request.Context(), apartmentID, condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error listing reservations by apartment")
		InternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, dto.ReservationsListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     reservations,
	})
}

// Update actualiza una reserva
// PUT /api/v1/reservations/:id
func (h *ReservationHandler) Update(c *gin.Context) {
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

	// Parsear request de actualización
	var req dto.UpdateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	// Actualizar reserva
	reservation, err := h.service.Update(c.Request.Context(), id, &req, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("reservation_id", id).Error("error updating reservation")
		InternalServerError(c, "internal server error")
		return
	}

	// Notificar al dueño si cambió el estado (best-effort, en goroutine).
	// Usamos context.WithoutCancel para que no se cancele al enviar la respuesta.
	if req.Estado != nil && *req.Estado != "" {
		condID := condominioID.(string)
		newEstado := *req.Estado
		resID := reservation.ID
		userID := reservation.UsuarioID
		espacioID := reservation.EspacioID
		detachedCtx := context.WithoutCancel(c.Request.Context())
		go func() {
			espacioNombre, _ := h.service.GetEspacioNombre(detachedCtx, espacioID, condID)
			_ = h.notifService.NotifyReservationStatus(detachedCtx, condID, userID, resID, espacioNombre, newEstado)
		}()
	}

	c.JSON(http.StatusOK, reservation)
}

// Delete elimina una reserva
// DELETE /api/v1/reservations/:id
func (h *ReservationHandler) Delete(c *gin.Context) {
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

	// Eliminar reserva
	err := h.service.Delete(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("reservation_id", id).Error("error deleting reservation")
		InternalServerError(c, "internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}
