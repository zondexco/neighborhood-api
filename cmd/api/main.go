package main

import (
	"context"
	stderrs "errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"neighborhood-api/internal/config"
	"neighborhood-api/internal/database"
	"neighborhood-api/internal/handlers"
	"neighborhood-api/internal/middleware"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/internal/services"
	"neighborhood-api/internal/utils"
	"neighborhood-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

type responseStatusWriter struct {
	gin.ResponseWriter
	status int
	size   int
}

func (w *responseStatusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseStatusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger
	logger.Initialize(logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		EnableFile: cfg.Logging.EnableFile,
		FilePath:   cfg.Logging.FilePath,
		MaxSizeMB:  cfg.Logging.MaxSizeMB,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAgeDays: cfg.Logging.MaxAgeDays,
		Compress:   cfg.Logging.Compress,
		AlsoStdout: cfg.Logging.AlsoStdout,
	})
	log := logger.Get()

	healthVersion := Version
	if healthVersion == "" || healthVersion == "dev" {
		healthVersion = cfg.App.Version
	}

	log.WithField("app_name", cfg.App.Name).
		WithField("version", cfg.App.Version).
		WithField("build_version", Version).
		WithField("build_time", BuildTime).
		WithField("git_commit", GitCommit).
		WithField("environment", cfg.Server.Environment).
		Info("Starting application")

	// Conectar a base de datos
	db, err := database.New(cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Verificar salud de la base de datos
	if err := db.Health(); err != nil {
		log.WithError(err).Fatal("Database health check failed")
	}

	// Inicializar repositories
	userRepo := repositories.NewUserRepository(db)
	apartmentRepo := repositories.NewApartmentRepository(db, log)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	reservationRepo := repositories.NewReservationRepository(db)
	communicationRepo := repositories.NewCommunicationRepository(db)
	packageRepo := repositories.NewPackageRepository(db)
	condominioRepo := repositories.NewCondominioRepository(db)
	spaceRepo := repositories.NewSpaceRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)

	// Inicializar JWT manager
	jwtManager := utils.NewJWTManager(cfg.JWT)

	// Inicializar servicios
	authService := services.NewAuthService(userRepo, condominioRepo, jwtManager)
	apartmentService := services.NewApartmentService(apartmentRepo, log)
	invoiceService := services.NewInvoiceService(invoiceRepo)
	reservationService := services.NewReservationService(reservationRepo, invoiceRepo)
	communicationService := services.NewCommunicationService(communicationRepo)
	packageService := services.NewPackageService(packageRepo)
	spaceService := services.NewSpaceService(spaceRepo)
	notificationService := services.NewNotificationService(notificationRepo, userRepo)

	// Configurar Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Middleware global
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORSMiddleware(cfg.CORS))

	// Request logging middleware
	engine.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		start := time.Now()
		writer := &responseStatusWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = writer

		c.Next()

		latency := time.Since(start)
		requestLog := log.WithField("request_id", requestID).
			WithField("method", c.Request.Method).
			WithField("path", c.Request.URL.Path).
			WithField("status", writer.status).
			WithField("latency_ms", latency.Milliseconds()).
			WithField("response_bytes", writer.size).
			WithField("client_ip", c.ClientIP())

		if condominioID, exists := c.Get("condominio_id"); exists {
			requestLog = requestLog.WithField("condominio_id", condominioID)
		}

		if writer.status >= 500 {
			requestLog.Error("HTTP request completed with server error")
			return
		}
		if writer.status >= 400 {
			requestLog.Warn("HTTP request completed with client error")
			return
		}
		requestLog.Info("HTTP request completed")
	})

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	apartmentHandler := handlers.NewApartmentHandler(apartmentService, log)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService, log)
	reservationHandler := handlers.NewReservationHandler(reservationService, notificationService, log)
	communicationHandler := handlers.NewCommunicationHandler(communicationService)
	packageHandler := handlers.NewPackageHandler(packageService, notificationService)
	spaceHandler := handlers.NewSpaceHandler(spaceService)
	adminHandler := handlers.NewAdminHandler(userRepo, apartmentRepo, communicationRepo, condominioRepo, log)
	dashboardHandler := handlers.NewDashboardHandler(userRepo, packageService, reservationService, communicationService)
	devHandler := handlers.NewDevHandler(condominioRepo, userRepo, jwtManager)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Configurar rutas
	router := handlers.NewRouter(engine, healthVersion)
	router.SetupRoutes(authHandler, apartmentHandler, invoiceHandler, reservationHandler, communicationHandler, packageHandler, spaceHandler, adminHandler, dashboardHandler, devHandler, notificationHandler)

	log.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadTimeout) * time.Second,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Iniciar servidor en goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !stderrs.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("Server error")
		}
	}()

	// Esperar señal de terminación
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server shutdown with errors")
		return
	}

	log.Info("Server shutdown completed")
}
