package handlers

import (
	"net/http"

	"neighborhood-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	log logger.Logger
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		log: logger.Get(),
	}
}

// GetResidentSummary retorna el resumen para la home del residente
func (h *DashboardHandler) GetResidentSummary(c *gin.Context) {
	// TODO: Connect to real services (Invoice, Reservation, Package)
	// Mock response for UI testing
	c.JSON(http.StatusOK, gin.H{
		"debt_status": "clean", // clean | debt
		"debt_amount": 0.00,
		"packages_pending": 1,
		"next_reservation": nil,
		"news": []gin.H{
			{
				"id": "1",
				"title": "Mantenimiento Piscina",
				"date": "2026-02-10",
				"type": "maintenance",
			},
		},
	})
}
