package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// CreateCommunication crea un nuevo comunicado con campos extendidos
// POST /api/v1/admin/communications
func (h *AdminHandler) CreateCommunication(c *gin.Context) {
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
		Titulo             string   `json:"titulo" binding:"required"`
		Contenido          string   `json:"contenido" binding:"required"`
		ProgramadoPara     *string  `json:"programado_para"`
		RolesDestino       []string `json:"roles_destino"`
		Icono              string   `json:"icono"`
		PermiteComentarios bool     `json:"permite_comentarios"`
		Publicado          *bool    `json:"publicado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		BadRequest(c, err.Error())
		return
	}

	icono := req.Icono
	if icono == "" {
		icono = "megaphone"
	}

	publicado := true
	if req.Publicado != nil {
		publicado = *req.Publicado
	}

	communication := &models.Communication{
		Titulo:             req.Titulo,
		Contenido:          req.Contenido,
		Autor:              userID.(string),
		CondominioID:       condominioID.(string),
		Icono:              icono,
		PermiteComentarios: req.PermiteComentarios,
		Publicado:          publicado,
	}

	if len(req.RolesDestino) > 0 {
		communication.RolesDestino = req.RolesDestino
	}

	if req.ProgramadoPara != nil && *req.ProgramadoPara != "" {
		t, err := time.Parse(time.RFC3339, *req.ProgramadoPara)
		if err != nil {
			BadRequest(c, "invalid programado_para format, use RFC3339")
			return
		}
		communication.ProgramadoPara = &t
	}

	if err := h.communicationRepository.Create(c.Request.Context(), communication); err != nil {
		h.logger.WithError(err).Error("error creating admin communication")
		InternalServerError(c, "error creating communication")
		return
	}

	c.JSON(http.StatusCreated, communication)
}

// UpdateCommunication actualiza un comunicado existente
// PUT /api/v1/admin/communications/:id
func (h *AdminHandler) UpdateCommunication(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	commID := c.Param("id")
	if commID == "" {
		BadRequest(c, "invalid request")
		return
	}

	existing, err := h.communicationRepository.FindByID(c.Request.Context(), commID, condominioID.(string))
	if err != nil {
		h.logger.WithError(err).WithField("communication_id", commID).Warn("communication not found")
		NotFound(c, "communication not found")
		return
	}

	var req struct {
		Titulo             *string  `json:"titulo"`
		Contenido          *string  `json:"contenido"`
		ProgramadoPara     *string  `json:"programado_para"`
		RolesDestino       []string `json:"roles_destino"`
		Icono              *string  `json:"icono"`
		PermiteComentarios *bool    `json:"permite_comentarios"`
		Publicado          *bool    `json:"publicado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if req.Titulo != nil {
		existing.Titulo = *req.Titulo
	}
	if req.Contenido != nil {
		existing.Contenido = *req.Contenido
	}
	if req.Icono != nil {
		existing.Icono = *req.Icono
	}
	if req.PermiteComentarios != nil {
		existing.PermiteComentarios = *req.PermiteComentarios
	}
	if req.Publicado != nil {
		existing.Publicado = *req.Publicado
	}
	if req.RolesDestino != nil {
		existing.RolesDestino = req.RolesDestino
	}
	if req.ProgramadoPara != nil {
		if *req.ProgramadoPara == "" {
			existing.ProgramadoPara = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.ProgramadoPara)
			if err != nil {
				BadRequest(c, "invalid programado_para format")
				return
			}
			existing.ProgramadoPara = &t
		}
	}

	if err := h.communicationRepository.Update(c.Request.Context(), existing); err != nil {
		h.logger.WithError(err).Error("error updating communication")
		InternalServerError(c, "error updating communication")
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteCommunication elimina un comunicado
// DELETE /api/v1/admin/communications/:id
func (h *AdminHandler) DeleteCommunication(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	commID := c.Param("id")
	if commID == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.communicationRepository.Delete(c.Request.Context(), commID, condominioID.(string)); err != nil {
		h.logger.WithError(err).Error("error deleting communication")
		InternalServerError(c, "error deleting communication")
		return
	}

	c.Status(http.StatusNoContent)
}

// ListCommunicationsAdmin lista TODOS los comunicados del condominio (sin filtro de visibilidad)
// GET /api/v1/admin/communications
func (h *AdminHandler) ListCommunicationsAdmin(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 200 {
			pageSize = parsed
		}
	}

	communications, total, err := h.communicationRepository.GetByCondominio(c.Request.Context(), condominioID.(string), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("error fetching communications")
		InternalServerError(c, "error fetching communications")
		return
	}

	formatted := make([]gin.H, 0, len(communications))
	for _, comm := range communications {
		formatted = append(formatted, gin.H{
			"id":                  comm.ID,
			"titulo":              comm.Titulo,
			"contenido":           comm.Contenido,
			"fecha":               comm.Fecha,
			"autor":               comm.Autor,
			"condominio_id":       comm.CondominioID,
			"programado_para":     comm.ProgramadoPara,
			"roles_destino":       comm.RolesDestino,
			"icono":               comm.Icono,
			"permite_comentarios": comm.PermiteComentarios,
			"publicado":           comm.Publicado,
			"created_at":          comm.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       formatted,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// UpdateCondominio actualiza la información del condominio
// PUT /api/v1/admin/condominio
func (h *AdminHandler) UpdateCondominio(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	condominio, err := h.condominioRepository.FindByID(c.Request.Context(), condominioID.(string))
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

	if err := h.condominioRepository.Update(c.Request.Context(), condominio); err != nil {
		h.logger.WithError(err).Error("error updating condominio")
		InternalServerError(c, "error updating condominio")
		return
	}

	c.JSON(http.StatusOK, condominio)
}
