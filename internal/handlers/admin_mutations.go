package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/utils"
)

var allowedEstadoGeneral = map[string]string{
	"activo":     "activo",
	"inactivo":   "inactivo",
	"pendiente":  "pendiente",
	"suspendido": "suspendido",
	"cancelado":  "cancelado",
}

// CreateUser registra un usuario administrativo/empleado/usuario
// POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Nombres       string  `json:"nombres" binding:"required"`
		Apellidos     string  `json:"apellidos" binding:"required"`
		Email         string  `json:"email" binding:"required"`
		Telefono      *string `json:"telefono"`
		ApartamentoID *string `json:"apartamento_id"`
		Rol           string  `json:"rol"`
		Estado        string  `json:"estado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body for create user")
		BadRequest(c, err.Error())
		return
	}

	rol := normalizeRole(req.Rol)
	estado := normalizeEstadoGeneral(req.Estado)

	pinPlain, err := generatePIN()
	if err != nil {
		h.logger.WithError(err).Error("failed generating pin for user")
		InternalServerError(c, "failed generating PIN")
		return
	}

	pinHash, err := utils.HashPIN(pinPlain)
	if err != nil {
		h.logger.WithError(err).Error("failed hashing pin for user")
		InternalServerError(c, "failed generating PIN")
		return
	}

	user := &models.User{
		Nombre:       req.Nombres,
		Apellido:     req.Apellidos,
		Email:        req.Email,
		Telefono:     stringOrEmpty(req.Telefono),
		CondominioID: condominioID.(string),
		ApartmentID:  req.ApartamentoID,
		PIN:          pinHash,
		Rol:          rol,
		Estado:       estado,
		TipoDoc:      "",
		NumeroDoc:    "",
	}

	if err := h.userRepository.Create(c.Request.Context(), user); err != nil {
		h.logger.WithError(err).Error("error creating admin user")
		InternalServerError(c, "error creating user")
		return
	}

	response := gin.H{
		"user":         user,
		"pin_temporal": pinPlain,
	}

	if req.Telefono != nil {
		response["telefono"] = *req.Telefono
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateUser actualiza un usuario
// PUT /api/v1/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequest(c, "invalid request")
		return
	}

	var req struct {
		Nombres       *string `json:"nombres"`
		Apellidos     *string `json:"apellidos"`
		Email         *string `json:"email"`
		Telefono      *string `json:"telefono"`
		ApartamentoID *string `json:"apartamento_id"`
		Rol           *string `json:"rol"`
		Estado        *string `json:"estado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body for update user")
		BadRequest(c, err.Error())
		return
	}

	user, err := h.userRepository.FindByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Warn("user not found")
		NotFound(c, "user not found")
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
	if req.ApartamentoID != nil {
		if strings.TrimSpace(*req.ApartamentoID) == "" {
			user.ApartmentID = nil
		} else {
			user.ApartmentID = req.ApartamentoID
		}
	}
	if req.Rol != nil {
		user.Rol = normalizeRole(*req.Rol)
	}
	if req.Estado != nil {
		user.Estado = normalizeEstadoGeneral(*req.Estado)
	}

	if err := h.userRepository.Update(c.Request.Context(), user); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("error updating user")
		InternalServerError(c, "error updating user")
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser elimina un usuario
// DELETE /api/v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.userRepository.Delete(c.Request.Context(), userID); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("error deleting user")
		InternalServerError(c, "error deleting user")
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateApartment crea un nuevo apartamento
// POST /api/v1/admin/apartments
func (h *AdminHandler) CreateApartment(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Numero        string   `json:"numero" binding:"required"`
		Bloque        *string  `json:"bloque" binding:"omitempty"`
		Torre         *string  `json:"torre" binding:"omitempty"`
		Estado        string   `json:"estado" binding:"omitempty"`
		Piso          *int     `json:"piso" binding:"omitempty"`
		Area          *float64 `json:"area" binding:"omitempty"`
		PropietarioID *string  `json:"propietario_id" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	req.Estado = normalizeEstadoGeneral(req.Estado)

	if req.Piso == nil {
		zeroP := 0
		req.Piso = &zeroP
	}

	// Si no se envía área, usar 0 para evitar restricciones de DB
	if req.Area == nil {
		zero := 0.0
		req.Area = &zero
	}
	areaComun := 0.0

	apartment := &models.Apartment{
		ID:            uuid.New().String(),
		CondominioID:  condominioID.(string),
		Numero:        req.Numero,
		Bloque:        req.Bloque,
		Torre:         req.Torre,
		Piso:          toStringPtr(req.Piso),
		PropietarioID: req.PropietarioID,
		AreaPrivada:   req.Area,
		AreaComun:     &areaComun,
		Estado:        req.Estado,
	}

	if err := h.apartmentRepository.Create(c.Request.Context(), apartment); err != nil {
		h.logger.WithError(err).Error("error creating apartment")
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			Conflict(c, "apartment already exists for this condominio")
			return
		}
		InternalServerError(c, "error creating apartment")
		return
	}

	c.JSON(http.StatusCreated, apartment)
}

// UpdateApartment actualiza un apartamento
// PUT /api/v1/admin/apartments/:id
func (h *AdminHandler) UpdateApartment(c *gin.Context) {
	// Extraer condominio del JWT
	_, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	apartmentID := c.Param("id")
	if apartmentID == "" {
		BadRequest(c, "invalid request")
		return
	}

	var req struct {
		Numero        *string  `json:"numero"`
		Bloque        *string  `json:"bloque"`
		Torre         *string  `json:"torre"`
		Area          *float64 `json:"area"`
		Estado        *string  `json:"estado"`
		Piso          *int     `json:"piso"`
		PropietarioID *string  `json:"propietario_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	existing, err := h.apartmentRepository.FindByID(c.Request.Context(), apartmentID)
	if err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Warn("apartment not found")
		NotFound(c, "apartment not found")
		return
	}

	if req.Numero != nil {
		existing.Numero = *req.Numero
	}
	if req.Bloque != nil {
		existing.Bloque = req.Bloque
	}
	if req.Torre != nil {
		existing.Torre = req.Torre
	}
	if req.Area != nil {
		existing.AreaPrivada = req.Area
	}
	if req.Estado != nil {
		existing.Estado = normalizeEstadoGeneral(*req.Estado)
	}
	if req.Piso != nil {
		existing.Piso = toStringPtr(req.Piso)
	}
	if req.PropietarioID != nil {
		existing.PropietarioID = req.PropietarioID
	}

	if err := h.apartmentRepository.Update(c.Request.Context(), existing); err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error updating apartment")
		InternalServerError(c, "error updating apartment")
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteApartment elimina un apartamento
// DELETE /api/v1/admin/apartments/:id
func (h *AdminHandler) DeleteApartment(c *gin.Context) {
	// Extraer condominio del JWT
	_, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	apartmentID := c.Param("id")
	if apartmentID == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.apartmentRepository.Delete(c.Request.Context(), apartmentID); err != nil {
		h.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error deleting apartment")
		InternalServerError(c, "error deleting apartment")
		return
	}

	c.Status(http.StatusNoContent)
}

func toStringPtr(value *int) *string {
	if value == nil {
		return nil
	}
	str := strconv.Itoa(*value)
	return &str
}

func normalizeEstadoGeneral(value string) string {
	val := strings.TrimSpace(strings.ToLower(value))
	switch val {
	case "disponible", "ocupado", "mantenimiento", "activo":
		return "activo"
	case "inactivo":
		return "inactivo"
	case "pendiente":
		return "pendiente"
	case "suspendido":
		return "suspendido"
	case "cancelado":
		return "cancelado"
	default:
		return "activo"
	}
}

func normalizeRole(value string) string {
	val := strings.TrimSpace(strings.ToLower(value))
	switch val {
	case "admin", "administrador", "administrator":
		return "administrador"
	case "empleado", "staff":
		return "empleado"
	case "usuario", "residente", "user", "resident":
		return "residente"
	default:
		return "residente"
	}
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func generatePIN() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	num := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return fmt.Sprintf("%06d", num%1000000), nil
}

// CreateCommunication crea un nuevo comunicado
// POST /api/v1/admin/communications
func (h *AdminHandler) CreateCommunication(c *gin.Context) {
	// Extraer condominio del JWT
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		h.logger.Warn("missing condominio_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	userID, userExists := c.Get("user_id")
	if !userExists {
		h.logger.Warn("missing user_id in context")
		Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Titulo    string `json:"titulo" binding:"required"`
		Contenido string `json:"contenido" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	communication := &models.Communication{
		Titulo:       req.Titulo,
		Contenido:    req.Contenido,
		Autor:        userID.(string),
		CondominioID: condominioID.(string),
	}

	if err := h.communicationRepository.Create(c.Request.Context(), communication); err != nil {
		h.logger.WithError(err).Error("error creating admin communication")
		InternalServerError(c, "error creating communication")
		return
	}

	c.JSON(http.StatusCreated, communication)
}
