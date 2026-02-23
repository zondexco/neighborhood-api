package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"neighborhood-api/internal/services"
)

// NotificationHandler maneja los endpoints de notificaciones in-app
type NotificationHandler struct {
	service services.NotificationService
}

func NewNotificationHandler(service services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// GET /api/v1/notifications
func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	condominioID := c.GetString("condominio_id")
	if userID == "" || condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	page, pageSize := 1, 30
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	notifications, total, err := h.service.List(c.Request.Context(), userID, condominioID, page, pageSize)
	if err != nil {
		InternalServerError(c, "error fetching notifications")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       notifications,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	condominioID := c.GetString("condominio_id")
	if userID == "" || condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	count, err := h.service.UnreadCount(c.Request.Context(), userID, condominioID)
	if err != nil {
		InternalServerError(c, "error counting notifications")
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// PUT /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	notifID := c.Param("id")
	if userID == "" || notifID == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), notifID, userID); err != nil {
		if err == sql.ErrNoRows {
			NotFound(c, "notification not found")
			return
		}
		InternalServerError(c, "error marking notification as read")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// PUT /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetString("user_id")
	condominioID := c.GetString("condominio_id")
	if userID == "" || condominioID == "" {
		Unauthorized(c, "unauthorized")
		return
	}

	if err := h.service.MarkAllRead(c.Request.Context(), userID, condominioID); err != nil {
		InternalServerError(c, "error marking all as read")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DELETE /api/v1/notifications/:id
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	notifID := c.Param("id")
	if userID == "" || notifID == "" {
		BadRequest(c, "invalid request")
		return
	}

	if err := h.service.Delete(c.Request.Context(), notifID, userID); err != nil {
		if err == sql.ErrNoRows {
			NotFound(c, "notification not found")
			return
		}
		InternalServerError(c, "error deleting notification")
		return
	}

	c.Status(http.StatusNoContent)
}
