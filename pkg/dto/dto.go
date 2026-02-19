package dto

import (
	"time"
)

// ========== Standard Error Response ==========

// ErrorResponse estructura estándar para respuestas de error
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ========== Auth DTOs ==========

// LoginRequest estructura para login
type LoginRequest struct {
	Email string `json:"email" binding:"required"`
	PIN   string `json:"pin" binding:"required"`
}

// LoginResponse estructura de respuesta de login
type LoginResponse struct {
	Token          string    `json:"token"`
	RefreshToken   string    `json:"refresh_token"`
	UserID         string    `json:"user_id"`
	Email          string    `json:"email"`
	Nombre         string    `json:"nombre"`
	Apellido       string    `json:"apellido"`
	CondominioID   string    `json:"condominio_id"`
	CondominioName string    `json:"condominio_nombre"`
	Role           string    `json:"role"`
	IsAdmin        bool      `json:"is_admin"`
	ApartmentID    string    `json:"apartment_id"` // "" si el usuario no tiene apartamento
	Permissions    []string  `json:"permissions"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// RefreshTokenRequest estructura para refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePINRequest estructura para cambiar PIN
type ChangePINRequest struct {
	CurrentPin string `json:"current_pin" binding:"required,len=6,numeric"`
	NewPin     string `json:"new_pin" binding:"required,len=6,numeric"`
}

// ========== User DTOs ==========

// UserDTO estructura de usuario
type UserDTO struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Nombre       string    `json:"nombre"`
	Apellido     string    `json:"apellido"`
	Telefono     string    `json:"telefono"`
	CondominioID string    `json:"condominio_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ========== Apartment DTOs ==========

// ApartmentDTO estructura de apartamento
type ApartmentDTO struct {
	ID                string    `json:"id"`
	Numero            string    `json:"numero"`
	Area              float64   `json:"area"`
	Estado            string    `json:"estado"`
	Piso              int       `json:"piso"`
	Torre             string    `json:"torre"`
	Bloque            string    `json:"bloque"`
	PropietarioID     string    `json:"propietario_id"`
	PropietarioNombre string    `json:"propietario_nombre"`
	CondominioID      string    `json:"condominio_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ApartmentCreateRequest estructura para crear apartamento
type ApartmentCreateRequest struct {
	Numero        string  `json:"numero" binding:"required"`
	Area          float64 `json:"area" binding:"required,gt=0"`
	Estado        string  `json:"estado" binding:"required"`
	Piso          int     `json:"piso" binding:"required,gte=0"`
	PropietarioID string  `json:"propietario_id"`
}

// ApartmentUpdateRequest estructura para actualizar apartamento
type ApartmentUpdateRequest struct {
	Numero        string  `json:"numero"`
	Area          float64 `json:"area"`
	Estado        string  `json:"estado"`
	Piso          int     `json:"piso"`
	PropietarioID string  `json:"propietario_id"`
}

// ApartmentsListResponse respuesta con lista de apartamentos
type ApartmentsListResponse struct {
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Data     []ApartmentDTO `json:"data"`
}

// ========== Space DTOs ==========

type SpaceDTO struct {
	ID           string     `json:"id"`
	Nombre       string     `json:"nombre"`
	Descripcion  *string    `json:"descripcion,omitempty"`
	CostoHora    *float64   `json:"costo_hora,omitempty"`
	Estado       string     `json:"estado"`
	CondominioID string     `json:"condominio_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type CreateSpaceRequest struct {
	Nombre      string   `json:"nombre" binding:"required"`
	Descripcion *string  `json:"descripcion"`
	CostoHora   *float64 `json:"costo_hora"`
	Estado      *string  `json:"estado"`
}

type UpdateSpaceRequest struct {
	Nombre      *string  `json:"nombre"`
	Descripcion *string  `json:"descripcion"`
	CostoHora   *float64 `json:"costo_hora"`
	Estado      *string  `json:"estado"`
}

type SpacesListResponse struct {
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Data     []SpaceDTO `json:"data"`
}

// ========== Reservation DTOs ==========

// ReservationDTO estructura de reserva
type ReservationDTO struct {
	ID                string    `json:"id"`
	EspacioID         string    `json:"espacio_id"`
	EspacioNombre     string    `json:"espacio_nombre"`
	UsuarioID         string    `json:"usuario_id"`
	ValorBase         *float64  `json:"valor_base,omitempty"`
	CostoTotal        *float64  `json:"costo_total,omitempty"`
	FechaInicio       time.Time `json:"fecha_inicio"`
	FechaFin          time.Time `json:"fecha_fin"`
	PersonasEsperadas int       `json:"personas_esperadas"`
	Estado            string    `json:"estado"`
	CondominioID      string    `json:"condominio_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateReservationRequest estructura para crear reserva
type CreateReservationRequest struct {
	EspacioID         string    `json:"espacio_id" binding:"required"`
	FechaInicio       time.Time `json:"fecha_inicio" binding:"required"`
	FechaFin          time.Time `json:"fecha_fin" binding:"required"`
	PersonasEsperadas int       `json:"personas_esperadas" binding:"required,gt=0"`
}

// UpdateReservationRequest estructura para actualizar reserva
type UpdateReservationRequest struct {
	FechaInicio       *time.Time `json:"fecha_inicio,omitempty"`
	FechaFin          *time.Time `json:"fecha_fin,omitempty"`
	PersonasEsperadas *int       `json:"personas_esperadas,omitempty"`
	Estado            *string    `json:"estado,omitempty"`
}

// ReservationResponse estructura de respuesta de reserva
type ReservationResponse struct {
	ID                string    `json:"id"`
	EspacioID         string    `json:"espacio_id"`
	UsuarioID         string    `json:"usuario_id"`
	ValorBase         *float64  `json:"valor_base,omitempty"`
	CostoTotal        *float64  `json:"costo_total,omitempty"`
	FechaInicio       time.Time `json:"fecha_inicio"`
	FechaFin          time.Time `json:"fecha_fin"`
	PersonasEsperadas int       `json:"personas_esperadas"`
	Estado            string    `json:"estado"`
	CondominioID      string    `json:"condominio_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ReservationsListResponse respuesta con lista de reservas
type ReservationsListResponse struct {
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Data     []*ReservationResponse `json:"data"`
}

// ========== Communication DTOs ==========

// CommunicationDTO estructura de comunicado
type CommunicationDTO struct {
	ID           string    `json:"id"`
	Titulo       string    `json:"titulo"`
	Contenido    string    `json:"contenido"`
	Fecha        time.Time `json:"fecha"`
	Autor        string    `json:"autor"`
	CondominioID string    `json:"condominio_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CommunicationsListResponse respuesta con lista de comunicados
type CommunicationsListResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Data     []CommunicationDTO `json:"data"`
}

// ========== Invoice DTOs ==========

// CreateInvoiceRequest estructura para crear una factura
type CreateInvoiceRequest struct {
	ApartmentID   string    `json:"apartment_id" binding:"required"`
	InvoiceNumber string    `json:"invoice_number" binding:"required"`
	Description   string    `json:"description" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
	DueDate       time.Time `json:"due_date" binding:"required"`
	IssuedDate    time.Time `json:"issued_date" binding:"required"`
	InvoiceType   string    `json:"invoice_type" binding:"required"`
	PaymentMethod *string   `json:"payment_method,omitempty"`
	Notes         *string   `json:"notes,omitempty"`
}

// UpdateInvoiceRequest estructura para actualizar una factura
type UpdateInvoiceRequest struct {
	Description   *string    `json:"description,omitempty"`
	Amount        *float64   `json:"amount,omitempty"`
	Status        *string    `json:"status,omitempty"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	PaidDate      *time.Time `json:"paid_date,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
}

// InvoiceResponse estructura de respuesta de factura
type InvoiceResponse struct {
	ID            string     `json:"id"`
	CondominioID  string     `json:"condominio_id"`
	ApartmentID   string     `json:"apartment_id"`
	InvoiceNumber string     `json:"invoice_number"`
	Description   string     `json:"description"`
	Amount        float64    `json:"amount"`
	Status        string     `json:"status"`
	DueDate       time.Time  `json:"due_date"`
	IssuedDate    time.Time  `json:"issued_date"`
	PaidDate      *time.Time `json:"paid_date,omitempty"`
	InvoiceType   string     `json:"invoice_type"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// InvoicesListResponse respuesta con lista de facturas
type InvoicesListResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Data     []*InvoiceResponse `json:"data"`
}

// ========== Communication DTOs ==========

// CreateCommunicationRequest estructura para crear un comunicado
type CreateCommunicationRequest struct {
	Titulo    string     `json:"titulo" binding:"required"`
	Contenido string     `json:"contenido" binding:"required"`
	Fecha     *time.Time `json:"fecha,omitempty"`
}

// UpdateCommunicationRequest estructura para actualizar un comunicado
type UpdateCommunicationRequest struct {
	Titulo    *string    `json:"titulo,omitempty"`
	Contenido *string    `json:"contenido,omitempty"`
	Fecha     *time.Time `json:"fecha,omitempty"`
}

// CommunicationResponse estructura de respuesta de comunicado
type CommunicationResponse struct {
	ID           string    `json:"id"`
	Titulo       string    `json:"titulo"`
	Contenido    string    `json:"contenido"`
	Fecha        time.Time `json:"fecha"`
	Autor        string    `json:"autor"`
	CondominioID string    `json:"condominio_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ========== Package DTOs ==========

type CreatePackageRequest struct {
	ApartmentID string  `json:"apartment_id" binding:"required"`
	Resident    string  `json:"resident" binding:"required"`
	Carrier     string  `json:"carrier" binding:"required"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdatePackageRequest struct {
	Notes       *string `json:"notes,omitempty"`
	Resident    *string `json:"resident,omitempty"`    // admin only
	Carrier     *string `json:"carrier,omitempty"`     // admin only
	ApartmentID *string `json:"apartment_id,omitempty"` // admin only
}

// PackageFilter parámetros de filtrado para listado de paquetes
type PackageFilter struct {
	Status      string    // "pending" | "delivered" | "" (todos)
	ApartmentID string    // UUID | "" (todos) — forzado al del JWT para residentes
	Carrier     string    // exacto | ""
	Search      string    // ILIKE en resident + carrier
	DateFrom    time.Time // zero = sin límite inferior
	DateTo      time.Time // zero = sin límite superior
	Page        int
	PageSize    int
}

type PackageResponse struct {
	ID             string     `json:"id"`
	ApartmentID    string     `json:"apartment_id"`
	ApartmentLabel string     `json:"apartment"`
	Resident       string     `json:"resident"`
	Carrier        string     `json:"carrier"`
	Notes          *string    `json:"notes,omitempty"`
	ReceivedAt     time.Time  `json:"received_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CondominioID   string     `json:"condominio_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ========== Generic Responses ==========

// SuccessResponse estructura genérica de respuesta exitosa
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationParams parámetros de paginación comunes
type PaginationParams struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

// ========== Dashboard DTOs ==========

// DashboardPackageItem resumen de un paquete pendiente
type DashboardPackageItem struct {
	ID          string    `json:"id"`
	Carrier     string    `json:"carrier"`
	ReceivedAt  time.Time `json:"received_at"`
	Apartment   string    `json:"apartment"`
}

// DashboardReservation próxima reserva del residente
type DashboardReservation struct {
	ID          string    `json:"id"`
	EspacioID   string    `json:"espacio_id"`
	FechaInicio time.Time `json:"fecha_inicio"`
	FechaFin    time.Time `json:"fecha_fin"`
	Estado      string    `json:"estado"`
}

// DashboardNewsItem comunicado reciente
type DashboardNewsItem struct {
	ID     string    `json:"id"`
	Titulo string    `json:"titulo"`
	Fecha  time.Time `json:"fecha"`
}

// DashboardSummaryResponse resumen del dashboard para el residente
type DashboardSummaryResponse struct {
	PackagesPending int                    `json:"packages_pending"`
	PendingPackages []DashboardPackageItem `json:"pending_packages"`
	NextReservation *DashboardReservation  `json:"next_reservation"`
	RecentNews      []DashboardNewsItem    `json:"recent_news"`
}
