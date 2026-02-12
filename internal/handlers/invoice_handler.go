package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"
)

// InvoiceHandler maneja las rutas de facturas
type InvoiceHandler struct {
	service services.InvoiceService
	logger  logger.Logger
}

// NewInvoiceHandler crea un nuevo manejador de facturas
func NewInvoiceHandler(service services.InvoiceService, log logger.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		service: service,
		logger:  log,
	}
}

// Create crea una nueva factura
// POST /api/v1/invoices
func (h *InvoiceHandler) Create(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Extraer usuario del JWT
	usuarioID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("missing user_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Parsear request
	var req dto.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	// Crear factura
	invoice, err := h.service.Create(c.Request.Context(), &req, condominioID.(string), usuarioID.(string))
	if err != nil {
		h.logger.WithError(err).Error("error creating invoice")
		InternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

// GetByID obtiene una factura por ID
// GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetByID(c *gin.Context) {
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

	invoice, err := h.service.GetByID(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", id).Warn("invoice not found")
		NotFound(c, "not found")
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// List obtiene la lista de facturas con paginación
// GET /api/v1/invoices?page=1&page_size=10
func (h *InvoiceHandler) List(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	// Extraer usuario
	usuarioID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("missing user_id in context")
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

	// Listar facturas
	invoices, total, err := h.service.List(c.Request.Context(), condominioID.(string), usuarioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("error listing invoices")
		InternalServerError(c, "internal server error")
		return
	}

	response := dto.InvoicesListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     invoices,
	}

	c.JSON(http.StatusOK, response)
}

// ListByApartment obtiene facturas por apartamento
// GET /api/v1/apartments/:apartment_id/invoices?page=1&page_size=10
func (h *InvoiceHandler) ListByApartment(c *gin.Context) {
	// Extraer condominio
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	apartmentID := c.Param("apartment_id")
	if apartmentID == "" {
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

	// Listar facturas del apartamento
	invoices, total, err := h.service.ListByApartment(c.Request.Context(), apartmentID, condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error listing invoices")
		InternalServerError(c, "internal server error")
		return
	}

	response := dto.InvoicesListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     invoices,
	}

	c.JSON(http.StatusOK, response)
}

// Update actualiza una factura
// PUT /api/v1/invoices/:id
func (h *InvoiceHandler) Update(c *gin.Context) {
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
	var req dto.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	// Actualizar factura
	invoice, err := h.service.Update(c.Request.Context(), id, &req, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", id).Error("error updating invoice")
		InternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// Delete elimina una factura
// DELETE /api/v1/invoices/:id
func (h *InvoiceHandler) Delete(c *gin.Context) {
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

	// Eliminar factura
	err := h.service.Delete(c.Request.Context(), id, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("invoice_id", id).Error("error deleting invoice")
		InternalServerError(c, "internal server error")
		return
	}

	c.Status(http.StatusNoContent)
}
