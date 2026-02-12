package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
)

// CommunicationHandler maneja las peticiones relacionadas con comunicados
type CommunicationHandler struct {
	service services.CommunicationService
}

// NewCommunicationHandler crea un nuevo manejador
func NewCommunicationHandler(service services.CommunicationService) *CommunicationHandler {
	return &CommunicationHandler{
		service: service,
	}
}

// CreateCommunication crea un nuevo comunicado
// @Summary Create a new communication
// @Description Crea un nuevo comunicado
// @Tags Communications
// @Accept json
// @Produce json
// @Param body body dto.CreateCommunicationRequest true "Communication data"
// @Success 201 {object} dto.CommunicationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/communications [post]
func (h *CommunicationHandler) CreateCommunication(c *gin.Context) {
	var request dto.CreateCommunicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	// Obtener condominioID del contexto JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	// Obtener usuarioID del contexto JWT (será el autor)
	usuarioID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	communication, err := h.service.Create(c.Request.Context(), request, condominioID.(string), usuarioID.(string))
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, toCommunicationResponse(communication))
}

// GetCommunication obtiene un comunicado por ID
func (h *CommunicationHandler) GetCommunication(c *gin.Context) {
	communicationID := c.Param("id")

	// Obtener condominioID del contexto JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	communication, err := h.service.GetByID(c.Request.Context(), communicationID, condominioID.(string))
	if err != nil {
		if errors.Is(err, errors.New("communication not found")) {
			NotFound(c, "not found")
		} else {
			InternalServerError(c, "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, toCommunicationResponse(communication))
}

// ListCommunications obtiene comunicados con paginación
func (h *CommunicationHandler) ListCommunications(c *gin.Context) {
	// Obtener condominioID del contexto JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	// Obtener parámetros de paginación
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	communications, total, err := h.service.List(c.Request.Context(), condominioID.(string), page, pageSize)
	if err != nil {
		InternalServerError(c, "internal server error")
		return
	}

	responses := make([]dto.CommunicationResponse, len(communications))
	for i, comm := range communications {
		responses[i] = toCommunicationResponse(comm)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       responses,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// UpdateCommunication actualiza un comunicado
func (h *CommunicationHandler) UpdateCommunication(c *gin.Context) {
	communicationID := c.Param("id")

	var request dto.UpdateCommunicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	// Obtener condominioID del contexto JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	communication, err := h.service.Update(c.Request.Context(), request, communicationID, condominioID.(string))
	if err != nil {
		if errors.Is(err, errors.New("communication not found")) {
			NotFound(c, "not found")
		} else {
			BadRequest(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, toCommunicationResponse(communication))
}

// DeleteCommunication elimina un comunicado
func (h *CommunicationHandler) DeleteCommunication(c *gin.Context) {
	communicationID := c.Param("id")

	// Obtener condominioID del contexto JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	err := h.service.Delete(c.Request.Context(), communicationID, condominioID.(string))
	if err != nil {
		if errors.Is(err, errors.New("communication not found")) {
			NotFound(c, "not found")
		} else {
			InternalServerError(c, "internal server error")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// toCommunicationResponse convierte un modelo a respuesta
func toCommunicationResponse(c *models.Communication) dto.CommunicationResponse {
	return dto.CommunicationResponse{
		ID:           c.ID,
		Titulo:       c.Titulo,
		Contenido:    c.Contenido,
		Fecha:        c.Fecha,
		Autor:        c.Autor,
		CondominioID: c.CondominioID,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
