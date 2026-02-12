package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
)

// SpaceHandler maneja rutas de espacios comunes
type SpaceHandler struct {
	service services.SpaceService
}

// NewSpaceHandler crea un nuevo handler
func NewSpaceHandler(service services.SpaceService) *SpaceHandler {
	return &SpaceHandler{service: service}
}

// Create crea un espacio comun
// POST /api/v1/spaces
func (h *SpaceHandler) Create(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	space, err := h.service.Create(c.Request.Context(), &req, condominioID.(string))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, space)
}

// List devuelve espacios del condominio
// GET /api/v1/spaces
func (h *SpaceHandler) List(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 200 {
			pageSize = val
		}
	}

	search := strings.TrimSpace(c.Query("search"))

	res, err := h.service.List(c.Request.Context(), condominioID.(string), page, pageSize, search)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetByID obtiene un espacio
// GET /api/v1/spaces/:id
func (h *SpaceHandler) GetByID(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "invalid request")
		return
	}

	space, err := h.service.GetByID(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		NotFound(c, "not found")
		return
	}

	c.JSON(http.StatusOK, space)
}

// Update actualiza un espacio
// PUT /api/v1/spaces/:id
func (h *SpaceHandler) Update(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "invalid request")
		return
	}

	var req dto.UpdateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	space, err := h.service.Update(c.Request.Context(), id, &req, condominioID.(string))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, space)
}

// Delete elimina un espacio
// DELETE /api/v1/spaces/:id
func (h *SpaceHandler) Delete(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id, condominioID.(string)); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
