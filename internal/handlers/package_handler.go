package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
)

// PackageHandler maneja endpoints de paquetería
type PackageHandler struct {
	service services.PackageService
}

func NewPackageHandler(service services.PackageService) *PackageHandler {
	return &PackageHandler{service: service}
}

// POST /api/v1/packages
func (h *PackageHandler) Create(c *gin.Context) {
	var req dto.CreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	pkg, err := h.service.Create(c.Request.Context(), req, condominioID.(string))
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, toPackageResponse(pkg))
}

// GET /api/v1/packages
func (h *PackageHandler) List(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	page := 1
	pageSize := 20
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

	packages, total, err := h.service.List(c.Request.Context(), condominioID.(string), page, pageSize)
	if err != nil {
		InternalServerError(c, "internal server error")
		return
	}

	responses := make([]dto.PackageResponse, len(packages))
	for i, p := range packages {
		responses[i] = toPackageResponse(p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       responses,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// GET /api/v1/apartments/:id/packages
func (h *PackageHandler) ListByApartment(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	apartmentID := c.Param("id")
	if apartmentID == "" {
		BadRequest(c, "apartment id is required")
		return
	}

	page := 1
	pageSize := 20
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

	packages, total, err := h.service.ListByApartment(
		c.Request.Context(),
		condominioID.(string),
		apartmentID,
		page,
		pageSize,
	)
	if err != nil {
		InternalServerError(c, "internal server error")
		return
	}

	responses := make([]dto.PackageResponse, len(packages))
	for i, p := range packages {
		responses[i] = toPackageResponse(p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       responses,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// PUT /api/v1/packages/:id/deliver
func (h *PackageHandler) MarkDelivered(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")
	pkg, err := h.service.MarkDelivered(c.Request.Context(), packageID, condominioID.(string))
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, toPackageResponse(pkg))
}

// POST /api/v1/packages/:id/notify
func (h *PackageHandler) Notify(c *gin.Context) {
	condominioID, ok := c.Get("condominio_id")
	if !ok {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")
	if err := h.service.Notify(c.Request.Context(), packageID, condominioID.(string)); err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func toPackageResponse(p *models.Package) dto.PackageResponse {
	return dto.PackageResponse{
		ID:             p.ID,
		ApartmentID:    p.ApartmentID,
		ApartmentLabel: buildApartmentLabel(p),
		Resident:       p.Resident,
		Carrier:        p.Carrier,
		Notes:          p.Notes,
		ReceivedAt:     p.ReceivedAt,
		DeliveredAt:    p.DeliveredAt,
		CondominioID:   p.CondominioID,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func buildApartmentLabel(p *models.Package) string {
	parts := []string{}
	if p.ApartmentTower != nil && *p.ApartmentTower != "" {
		parts = append(parts, "Torre "+*p.ApartmentTower)
	}
	if p.ApartmentFloor != nil && *p.ApartmentFloor != "" {
		parts = append(parts, "Piso "+*p.ApartmentFloor)
	}
	if p.ApartmentNumber != nil && *p.ApartmentNumber != "" {
		parts = append(parts, "Apt "+*p.ApartmentNumber)
	}
	if len(parts) == 0 {
		return "Apartamento"
	}
	return joinParts(parts)
}

func joinParts(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " • "
		}
		result += part
	}
	return result
}
