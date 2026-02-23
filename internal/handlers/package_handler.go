package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
)

// PackageHandler maneja endpoints de paquetería
type PackageHandler struct {
	service      services.PackageService
	notifService services.NotificationService
}

func NewPackageHandler(service services.PackageService, notifService services.NotificationService) *PackageHandler {
	return &PackageHandler{service: service, notifService: notifService}
}

// POST /api/v1/packages — empleado + admin
func (h *PackageHandler) Create(c *gin.Context) {
	role := c.GetString("role")
	if !isEmpleadoOrAdmin(role) {
		Forbidden(c, "only employees and admins can create packages")
		return
	}

	var req dto.CreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	pkg, err := h.service.Create(c.Request.Context(), req, condominioID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Notificar a los residentes del apartamento (best-effort)
	go h.notifService.NotifyPackage(
		c.Request.Context(),
		condominioID,
		req.ApartmentID,
		pkg.ID,
		req.Carrier,
		buildApartmentLabel(pkg),
	)

	c.JSON(http.StatusCreated, toPackageResponse(pkg))
}

// GET /api/v1/packages — todos (scoped por rol)
func (h *PackageHandler) List(c *gin.Context) {
	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	role := c.GetString("role")
	jwtApartmentID := c.GetString("apartment_id")

	f := dto.PackageFilter{
		Status:   c.Query("status"),
		Carrier:  c.Query("carrier"),
		Search:   c.Query("search"),
		Page:     parsePage(c),
		PageSize: parsePageSize(c),
	}

	// Residentes solo pueden ver su propio apartamento — forzado desde JWT
	if strings.EqualFold(role, "residente") {
		f.ApartmentID = jwtApartmentID
	} else {
		f.ApartmentID = c.Query("apartment_id")
	}

	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			f.DateFrom = t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			f.DateTo = t
		}
	}

	packages, total, err := h.service.ListWithFilters(c.Request.Context(), condominioID, f)
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
		"page":       f.Page,
		"pageSize":   f.PageSize,
		"total":      total,
		"totalPages": (total + f.PageSize - 1) / f.PageSize,
	})
}

// GET /api/v1/packages/:id — todos
func (h *PackageHandler) GetByID(c *gin.Context) {
	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")
	pkg, err := h.service.GetByID(c.Request.Context(), packageID, condominioID)
	if err != nil {
		NotFound(c, "package not found")
		return
	}

	// Residentes solo pueden ver paquetes de su apartamento
	role := c.GetString("role")
	if strings.EqualFold(role, "residente") {
		jwtApartmentID := c.GetString("apartment_id")
		if pkg.ApartmentID != jwtApartmentID {
			Forbidden(c, "access denied")
			return
		}
	}

	c.JSON(http.StatusOK, toPackageResponse(pkg))
}

// PUT /api/v1/packages/:id — empleado + admin
func (h *PackageHandler) Update(c *gin.Context) {
	role := c.GetString("role")
	if !isEmpleadoOrAdmin(role) {
		Forbidden(c, "only employees and admins can update packages")
		return
	}

	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")

	var req dto.UpdatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	pkg, err := h.service.Update(c.Request.Context(), packageID, condominioID, role, req)
	if err != nil {
		if err.Error() == "package not found" {
			NotFound(c, "package not found")
			return
		}
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, toPackageResponse(pkg))
}

// DELETE /api/v1/packages/:id — solo admin
func (h *PackageHandler) Delete(c *gin.Context) {
	role := c.GetString("role")
	if !isAdmin(role) {
		Forbidden(c, "only admins can delete packages")
		return
	}

	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), packageID, condominioID); err != nil {
		if err.Error() == "package not found" {
			NotFound(c, "package not found")
			return
		}
		InternalServerError(c, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GET /api/v1/apartments/:id/packages
func (h *PackageHandler) ListByApartment(c *gin.Context) {
	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	apartmentID := c.Param("id")
	if apartmentID == "" {
		BadRequest(c, "apartment id is required")
		return
	}

	// Residentes solo pueden ver su propio apartamento
	role := c.GetString("role")
	if strings.EqualFold(role, "residente") {
		jwtApartmentID := c.GetString("apartment_id")
		if apartmentID != jwtApartmentID {
			Forbidden(c, "access denied")
			return
		}
	}

	f := dto.PackageFilter{
		ApartmentID: apartmentID,
		Page:        parsePage(c),
		PageSize:    parsePageSize(c),
	}

	packages, total, err := h.service.ListWithFilters(c.Request.Context(), condominioID, f)
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
		"page":       f.Page,
		"pageSize":   f.PageSize,
		"total":      total,
		"totalPages": (total + f.PageSize - 1) / f.PageSize,
	})
}

// PUT /api/v1/packages/:id/deliver — empleado + admin
func (h *PackageHandler) MarkDelivered(c *gin.Context) {
	role := c.GetString("role")
	if !isEmpleadoOrAdmin(role) {
		Forbidden(c, "only employees and admins can mark packages as delivered")
		return
	}

	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")
	pkg, err := h.service.MarkDelivered(c.Request.Context(), packageID, condominioID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, toPackageResponse(pkg))
}

// POST /api/v1/packages/:id/notify
func (h *PackageHandler) Notify(c *gin.Context) {
	condominioID := c.GetString("condominio_id")
	if condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	packageID := c.Param("id")
	if err := h.service.Notify(c.Request.Context(), packageID, condominioID); err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func isEmpleadoOrAdmin(role string) bool {
	r := strings.ToLower(role)
	return r == "empleado" || r == "administrador" || r == "admin"
}

func isAdmin(role string) bool {
	r := strings.ToLower(role)
	return r == "administrador" || r == "admin"
}

func parsePage(c *gin.Context) int {
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 1
}

func parsePageSize(c *gin.Context) int {
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			return parsed
		}
	}
	return 20
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
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " • "
		}
		result += part
	}
	return result
}
