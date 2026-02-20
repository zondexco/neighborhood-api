package handlers

import (
	"neighborhood-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router estructura para manejar rutas
type Router struct {
	engine     *gin.Engine
	appVersion string
}

// NewRouter crea un nuevo router
func NewRouter(engine *gin.Engine, appVersion string) *Router {
	return &Router{
		engine:     engine,
		appVersion: appVersion,
	}
}

// SetupRoutes configura todas las rutas de la API
func (r *Router) SetupRoutes(
	authHandler *AuthHandler,
	apartmentHandler *ApartmentHandler,
	invoiceHandler *InvoiceHandler,
	reservationHandler *ReservationHandler,
	communicationHandler *CommunicationHandler,
	packageHandler *PackageHandler,
	spaceHandler *SpaceHandler,
	adminHandler *AdminHandler,
	dashboardHandler *DashboardHandler,
	devHandler *DevHandler,
) {
	// API v1
	v1 := r.engine.Group("/api/v1")
	v1.GET("/health", Health(r.appVersion))

	// Auth routes (sin protección)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// Rutas protegidas con JWT
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(authHandler.AuthService))
	{
		// Auth routes protegidas (requieren autenticación)
		authProtected := protected.Group("/auth")
		{
			authProtected.POST("/change-pin", authHandler.ChangePIN)
			authProtected.POST("/logout", authHandler.Logout)
		}
		// Apartment routes
		apartments := protected.Group("/apartments")
		{
			apartments.POST("", apartmentHandler.Create)
			apartments.GET("", apartmentHandler.List)
			apartments.GET("/:id", apartmentHandler.GetByID)
			apartments.PUT("/:id", apartmentHandler.Update)
			apartments.DELETE("/:id", apartmentHandler.Delete)
			// Invoices por apartamento - DENTRO del grupo
			apartments.GET("/:id/invoices", invoiceHandler.ListByApartment)
			apartments.GET("/:id/packages", packageHandler.ListByApartment)
		}

		// Invoice routes
		invoices := protected.Group("/invoices")
		{
			invoices.POST("", invoiceHandler.Create)
			invoices.GET("", invoiceHandler.List)
			invoices.GET("/:id", invoiceHandler.GetByID)
			invoices.PUT("/:id", invoiceHandler.Update)
			invoices.DELETE("/:id", invoiceHandler.Delete)
		}

		// Reservation routes
		reservations := protected.Group("/reservations")
		{
			reservations.POST("", reservationHandler.Create)
			reservations.GET("", reservationHandler.List)
			reservations.GET("/:id", reservationHandler.GetByID)
			reservations.PUT("/:id", reservationHandler.Update)
			reservations.DELETE("/:id", reservationHandler.Delete)
		}

		// Reservations by user
		users := protected.Group("/users")
		{
			users.GET("/:user_id/reservations", reservationHandler.ListByUsuario)
		}

		// Reservations by space
		spacesLegacy := protected.Group("/espacios")
		{
			spacesLegacy.GET("/:espacio_id/reservations", reservationHandler.ListByEspacio)
		}

		// Common spaces CRUD
		spaces := protected.Group("/spaces")
		{
			spaces.POST("", spaceHandler.Create)
			spaces.GET("", spaceHandler.List)
			spaces.GET("/:id", spaceHandler.GetByID)
			spaces.PUT("/:id", spaceHandler.Update)
			spaces.DELETE("/:id", spaceHandler.Delete)
		}

		// Communication routes (any authenticated user)
		communications := protected.Group("/communications")
		{
			communications.GET("", communicationHandler.ListCommunications)
			communications.GET("/unread-count", communicationHandler.UnreadCount)
			communications.GET("/:id", communicationHandler.GetCommunication)
			communications.PUT("/:id/read", communicationHandler.MarkRead)
			communications.GET("/:id/comments", communicationHandler.ListComments)
			communications.POST("/:id/comments", communicationHandler.CreateComment)
		}

		// Condominio info (any authenticated user can read)
		protected.GET("/condominio", adminHandler.GetCondominio)

		// Package routes
		packages := protected.Group("/packages")
		{
			packages.POST("", packageHandler.Create)
			packages.GET("", packageHandler.List)
			packages.GET("/:id", packageHandler.GetByID)
			packages.PUT("/:id", packageHandler.Update)
			packages.DELETE("/:id", packageHandler.Delete)
			packages.PUT("/:id/deliver", packageHandler.MarkDelivered)
			packages.POST("/:id/notify", packageHandler.Notify)
		}

		// Dashboard routes (for authenticated residents)
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/summary", dashboardHandler.GetResidentSummary)
		}

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRoles("administrador", "admin", "dev"))
		{
			admin.GET("/condominio", adminHandler.GetCondominio)
			admin.PUT("/condominio", adminHandler.UpdateCondominio)
			admin.GET("/stats", adminHandler.GetStats)
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUserByID)
			admin.POST("/users", adminHandler.CreateUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.GET("/apartments", adminHandler.ListApartments)
			admin.GET("/apartments/:id", adminHandler.GetApartmentByID)
			admin.POST("/apartments", adminHandler.CreateApartment)
			admin.PUT("/apartments/:id", adminHandler.UpdateApartment)
			admin.DELETE("/apartments/:id", adminHandler.DeleteApartment)
			admin.GET("/communications", adminHandler.ListCommunicationsAdmin)
			admin.POST("/communications", adminHandler.CreateCommunication)
			admin.PUT("/communications/:id", adminHandler.UpdateCommunication)
			admin.DELETE("/communications/:id", adminHandler.DeleteCommunication)
		}

		// Dev routes (dev role only)
		dev := protected.Group("/dev")
		dev.Use(middleware.RequireRoles("dev"))
		{
			dev.GET("/condominios", devHandler.ListCondominios)
			dev.POST("/condominios", devHandler.CreateCondominio)
			dev.PUT("/condominios/:id", devHandler.UpdateCondominio)
			dev.DELETE("/condominios/:id", devHandler.DeleteCondominio)
			dev.GET("/condominios/:id/admins", devHandler.ListAdminsByCondominio)
			dev.POST("/impersonate", devHandler.Impersonate)
		}
	}
}
