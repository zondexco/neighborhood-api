package handlers

import (
	"net/http"
	"time"

	"neighborhood-api/internal/repositories"
	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	userRepo             repositories.UserRepository
	packageService       services.PackageService
	reservationService   services.ReservationService
	communicationService services.CommunicationService
	log                  logger.Logger
}

func NewDashboardHandler(
	userRepo repositories.UserRepository,
	packageService services.PackageService,
	reservationService services.ReservationService,
	communicationService services.CommunicationService,
) *DashboardHandler {
	return &DashboardHandler{
		userRepo:             userRepo,
		packageService:       packageService,
		reservationService:   reservationService,
		communicationService: communicationService,
		log:                  logger.Get(),
	}
}

// GetResidentSummary retorna el resumen para la home del residente
func (h *DashboardHandler) GetResidentSummary(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		h.log.Warn("missing user_id in context for dashboard summary")
		Unauthorized(c, "unauthorized")
		return
	}

	condominioIDVal, exists := c.Get("condominio_id")
	if !exists {
		h.log.Warn("missing condominio_id in context for dashboard summary")
		Unauthorized(c, "unauthorized")
		return
	}

	userID := userIDVal.(string)
	condominioID := condominioIDVal.(string)
	now := time.Now()

	// Obtener usuario para extraer su apartamento
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		h.log.WithError(err).WithField("user_id", userID).Error("error fetching user for dashboard")
		InternalServerError(c, "error fetching user data")
		return
	}

	// ── Paquetes pendientes ──
	pendingPackages := []dto.DashboardPackageItem{}
	packagesPendingCount := 0

	if user.ApartmentID != nil && *user.ApartmentID != "" {
		packages, _, err := h.packageService.ListByApartment(c.Request.Context(), condominioID, *user.ApartmentID, 1, 10)
		if err != nil {
			h.log.WithError(err).Warn("error fetching packages for dashboard — continuing without packages")
		} else {
			for _, pkg := range packages {
				if pkg.DeliveredAt == nil {
					packagesPendingCount++
					label := ""
					if pkg.ApartmentNumber != nil {
						label = *pkg.ApartmentNumber
					}
					if pkg.ApartmentTower != nil && *pkg.ApartmentTower != "" {
						label = "Torre " + *pkg.ApartmentTower + " - " + label
					}
					pendingPackages = append(pendingPackages, dto.DashboardPackageItem{
						ID:         pkg.ID,
						Carrier:    pkg.Carrier,
						ReceivedAt: pkg.ReceivedAt,
						Apartment:  label,
					})
				}
			}
		}
	}

	// ── Próxima reserva ──
	var nextReservation *dto.DashboardReservation
	reservations, _, err := h.reservationService.ListByUsuario(c.Request.Context(), userID, condominioID, 1, 10)
	if err != nil {
		h.log.WithError(err).Warn("error fetching reservations for dashboard — continuing without reservations")
	} else {
		for _, res := range reservations {
			isActive := res.Estado == "activo" || res.Estado == "confirmada" || res.Estado == "pendiente"
			if isActive && res.FechaInicio.After(now) {
				nextReservation = &dto.DashboardReservation{
					ID:          res.ID,
					EspacioID:   res.EspacioID,
					FechaInicio: res.FechaInicio,
					FechaFin:    res.FechaFin,
					Estado:      res.Estado,
				}
				break
			}
		}
	}

	// ── Comunicados recientes ──
	recentNews := []dto.DashboardNewsItem{}
	communications, _, err := h.communicationService.List(c.Request.Context(), condominioID, 1, 5)
	if err != nil {
		h.log.WithError(err).Warn("error fetching communications for dashboard — continuing without news")
	} else {
		for _, comm := range communications {
			recentNews = append(recentNews, dto.DashboardNewsItem{
				ID:     comm.ID,
				Titulo: comm.Titulo,
				Fecha:  comm.Fecha,
			})
		}
	}

	c.JSON(http.StatusOK, dto.DashboardSummaryResponse{
		PackagesPending: packagesPendingCount,
		PendingPackages: pendingPackages,
		NextReservation: nextReservation,
		RecentNews:      recentNews,
	})
}
