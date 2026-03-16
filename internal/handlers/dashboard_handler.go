package handlers

import (
	"net/http"
	"sync"
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

// GetResidentSummary retorna el resumen para la home del residente.
// Las consultas independientes se ejecutan en paralelo para reducir latencia.
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
	ctx := c.Request.Context()
	now := time.Now()

	// Obtener usuario para extraer su apartamento (necesario antes de paralelizar)
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		h.log.WithError(err).WithField("user_id", userID).Error("error fetching user for dashboard")
		InternalServerError(c, "error fetching user data")
		return
	}

	// ── Ejecutar consultas independientes en paralelo ──
	var (
		mu                   sync.Mutex
		pendingPackages      []dto.DashboardPackageItem
		packagesPendingCount int
		nextReservation      *dto.DashboardReservation
		recentNews           []dto.DashboardNewsItem
		wg                   sync.WaitGroup
	)
	pendingPackages = []dto.DashboardPackageItem{}
	recentNews = []dto.DashboardNewsItem{}

	// Goroutine 1: Paquetes pendientes
	wg.Add(1)
	go func() {
		defer wg.Done()
		if user.ApartmentID == nil || *user.ApartmentID == "" {
			return
		}
		packages, _, pkgErr := h.packageService.ListWithFilters(ctx, condominioID, dto.PackageFilter{
			ApartmentID: *user.ApartmentID,
			Status:      "pending",
			Page:        1,
			PageSize:    10,
		})
		if pkgErr != nil {
			h.log.WithError(pkgErr).Warn("error fetching packages for dashboard — continuing without packages")
			return
		}
		items := make([]dto.DashboardPackageItem, 0, len(packages))
		for _, pkg := range packages {
			label := ""
			if pkg.ApartmentNumber != nil {
				label = *pkg.ApartmentNumber
			}
			if pkg.ApartmentTower != nil && *pkg.ApartmentTower != "" {
				label = "Torre " + *pkg.ApartmentTower + " - " + label
			}
			items = append(items, dto.DashboardPackageItem{
				ID:         pkg.ID,
				Carrier:    pkg.Carrier,
				ReceivedAt: pkg.ReceivedAt,
				Apartment:  label,
			})
		}
		mu.Lock()
		pendingPackages = items
		packagesPendingCount = len(packages)
		mu.Unlock()
	}()

	// Goroutine 2: Próxima reserva
	wg.Add(1)
	go func() {
		defer wg.Done()
		reservations, _, resErr := h.reservationService.ListByUsuario(ctx, userID, condominioID, 1, 10)
		if resErr != nil {
			h.log.WithError(resErr).Warn("error fetching reservations for dashboard — continuing without reservations")
			return
		}
		for _, res := range reservations {
			isActive := res.Estado == "activo" || res.Estado == "confirmada" || res.Estado == "pendiente"
			if isActive && res.FechaInicio.After(now) {
				mu.Lock()
				nextReservation = &dto.DashboardReservation{
					ID:          res.ID,
					EspacioID:   res.EspacioID,
					FechaInicio: res.FechaInicio,
					FechaFin:    res.FechaFin,
					Estado:      res.Estado,
				}
				mu.Unlock()
				break
			}
		}
	}()

	// Goroutine 3: Comunicados recientes
	wg.Add(1)
	go func() {
		defer wg.Done()
		communications, _, commErr := h.communicationService.List(ctx, condominioID, 1, 5)
		if commErr != nil {
			h.log.WithError(commErr).Warn("error fetching communications for dashboard — continuing without news")
			return
		}
		items := make([]dto.DashboardNewsItem, 0, len(communications))
		for _, comm := range communications {
			items = append(items, dto.DashboardNewsItem{
				ID:     comm.ID,
				Titulo: comm.Titulo,
				Fecha:  comm.Fecha,
			})
		}
		mu.Lock()
		recentNews = items
		mu.Unlock()
	}()

	wg.Wait()

	c.JSON(http.StatusOK, dto.DashboardSummaryResponse{
		PackagesPending: packagesPendingCount,
		PendingPackages: pendingPackages,
		NextReservation: nextReservation,
		RecentNews:      recentNews,
	})
}
