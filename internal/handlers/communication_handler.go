package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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

// ListCommunications obtiene comunicados visibles para el usuario autenticado
func (h *CommunicationHandler) ListCommunications(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}
	userRole, _ := c.Get("role")

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

	communications, total, err := h.service.ListVisible(
		c.Request.Context(),
		condominioID.(string),
		userID.(string),
		userRole.(string),
		page,
		pageSize,
	)
	if err != nil {
		InternalServerError(c, "internal server error")
		return
	}

	responses := make([]gin.H, 0, len(communications))
	for _, comm := range communications {
		responses = append(responses, gin.H{
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
			"leido":               comm.Leido,
			"num_comentarios":     comm.NumComentarios,
			"created_at":          comm.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       responses,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// GetCommunication obtiene un comunicado por ID
func (h *CommunicationHandler) GetCommunication(c *gin.Context) {
	communicationID := c.Param("id")
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	communication, err := h.service.GetByID(c.Request.Context(), communicationID, condominioID.(string))
	if err != nil {
		NotFound(c, "not found")
		return
	}

	c.JSON(http.StatusOK, communication)
}

// MarkRead marca un comunicado como leído
// PUT /api/v1/communications/:id/read
func (h *CommunicationHandler) MarkRead(c *gin.Context) {
	communicationID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), communicationID, userID.(string)); err != nil {
		InternalServerError(c, "error marking as read")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UnreadCount retorna la cantidad de comunicados sin leer
// GET /api/v1/communications/unread-count
func (h *CommunicationHandler) UnreadCount(c *gin.Context) {
	condominioID, exists := c.Get("condominio_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}
	userRole, _ := c.Get("role")

	count, err := h.service.UnreadCount(c.Request.Context(), condominioID.(string), userID.(string), userRole.(string))
	if err != nil {
		InternalServerError(c, "error getting unread count")
		return
	}

	c.JSON(http.StatusOK, dto.UnreadCountResponse{Count: count})
}

// ListComments obtiene los comentarios de un comunicado
// GET /api/v1/communications/:id/comments
func (h *CommunicationHandler) ListComments(c *gin.Context) {
	communicationID := c.Param("id")

	comments, err := h.service.ListComments(c.Request.Context(), communicationID)
	if err != nil {
		InternalServerError(c, "error fetching comments")
		return
	}

	responses := make([]dto.CommentResponse, 0, len(comments))
	for _, com := range comments {
		responses = append(responses, dto.CommentResponse{
			ID:            com.ID,
			IDComunicado:  com.IDComunicado,
			IDUsuario:     com.IDUsuario,
			Contenido:     com.Contenido,
			AutorNombre:   com.AutorNombre,
			AutorApellido: com.AutorApellido,
			FechaCreacion: com.FechaCreacion,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// CreateComment crea un comentario en un comunicado
// POST /api/v1/communications/:id/comments
func (h *CommunicationHandler) CreateComment(c *gin.Context) {
	communicationID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request")
		return
	}

	comment, err := h.service.CreateComment(c.Request.Context(), communicationID, userID.(string), req.Contenido)
	if err != nil {
		InternalServerError(c, "error creating comment")
		return
	}

	c.JSON(http.StatusCreated, dto.CommentResponse{
		ID:            comment.ID,
		IDComunicado:  comment.IDComunicado,
		IDUsuario:     comment.IDUsuario,
		Contenido:     comment.Contenido,
		AutorNombre:   comment.AutorNombre,
		AutorApellido: comment.AutorApellido,
		FechaCreacion: comment.FechaCreacion,
	})
}
