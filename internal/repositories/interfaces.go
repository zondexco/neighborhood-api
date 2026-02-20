package repositories

import (
	"context"
	"time"

	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/dto"
)

// UserRepository interfaz para operaciones con usuarios
type UserRepository interface {
	// FindByEmail busca un usuario por email dentro de un condominio específico
	FindByEmail(ctx context.Context, email string, condominioID string) (*models.User, error)

	// FindByEmailOnly busca un usuario por email sin requerir condominio_id (para login)
	FindByEmailOnly(ctx context.Context, email string) (*models.User, error)

	// FindByID busca un usuario por ID
	FindByID(ctx context.Context, userID string) (*models.User, error)

	// Create crea un nuevo usuario
	Create(ctx context.Context, user *models.User) error

	// Update actualiza un usuario
	Update(ctx context.Context, user *models.User) error

	// UpdatePIN actualiza el PIN de un usuario
	UpdatePIN(ctx context.Context, userID, newPINHash string) error

	// Delete elimina un usuario
	Delete(ctx context.Context, userID string) error

	// GetByCondominio obtiene todos los usuarios de un condominio
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.User, int, error)
}

// ApartmentRepository interfaz para operaciones con apartamentos
type ApartmentRepository interface {
	// FindByID busca un apartamento por ID
	FindByID(ctx context.Context, apartmentID string) (*models.Apartment, error)

	// GetByCondominio obtiene todos los apartamentos de un condominio
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Apartment, int, error)

	// GetByPropietario obtiene apartamentos de un propietario
	GetByPropietario(ctx context.Context, propietarioID string) ([]*models.Apartment, error)

	// GetByOwner obtiene apartamentos de un propietario
	GetByOwner(ctx context.Context, condominioID, ownerID string, page, pageSize int) ([]*models.Apartment, int, error)

	// Create crea un nuevo apartamento
	Create(ctx context.Context, apartment *models.Apartment) error

	// Update actualiza un apartamento
	Update(ctx context.Context, apartment *models.Apartment) error

	// Delete elimina un apartamento
	Delete(ctx context.Context, apartmentID string) error
}

// InvoiceRepository interfaz para operaciones con facturas
type InvoiceRepository interface {
	// FindByID busca una factura por ID dentro de un condominio
	FindByID(ctx context.Context, invoiceID, condominioID string) (*models.Invoice, error)

	// GetByCondominio obtiene todas las facturas de un condominio
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Invoice, int, error)

	// GetByUsuario obtiene facturas de un usuario
	GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error)

	// GetByApartment obtiene facturas de un apartamento
	GetByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error)

	// Create crea una nueva factura
	Create(ctx context.Context, invoice *models.Invoice) error

	// Update actualiza una factura
	Update(ctx context.Context, invoice *models.Invoice) error

	// Delete elimina una factura
	Delete(ctx context.Context, invoiceID, condominioID string) error
}

// ReservationRepository interfaz para operaciones con reservas
type ReservationRepository interface {
	// FindByID busca una reserva por ID dentro de un condominio
	FindByID(ctx context.Context, reservationID, condominioID string) (*models.Reservation, error)

	// GetEspacioNombre obtiene el nombre del espacio común
	GetEspacioNombre(ctx context.Context, espacioID, condominioID string) (string, error)

	// GetEspacioCostoHora obtiene el costo_hora de un espacio común
	GetEspacioCostoHora(ctx context.Context, espacioID, condominioID string) (*float64, error)

	// GetByCondominio obtiene todas las reservas de un condominio
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Reservation, int, error)

	// GetByUsuario obtiene las reservas de un usuario
	GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error)

	// GetByEspacio obtiene reservas de un espacio común
	GetByEspacio(ctx context.Context, espacioID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error)

	// ExistsOverlap valida solapamiento de reservas activas para un espacio
	ExistsOverlap(ctx context.Context, espacioID, condominioID string, start, end time.Time) (bool, error)

	// Create crea una nueva reserva
	Create(ctx context.Context, reservation *models.Reservation) error

	// Update actualiza una reserva
	Update(ctx context.Context, reservation *models.Reservation) error

	// Delete elimina una reserva
	Delete(ctx context.Context, reservationID, condominioID string) error
}

// SpaceRepository interfaz para operaciones con espacios comunes
type SpaceRepository interface {
	FindByID(ctx context.Context, espacioID, condominioID string) (*models.Space, error)
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int, search string) ([]*models.Space, int, error)
	Create(ctx context.Context, space *models.Space) error
	Update(ctx context.Context, space *models.Space) error
	Delete(ctx context.Context, espacioID, condominioID string) error
}

// CommunicationRepository interfaz para operaciones con comunicados
type CommunicationRepository interface {
	// FindByID busca un comunicado por ID (multi-tenant)
	FindByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error)

	// GetByCondominio obtiene todos los comunicados de un condominio (admin, sin filtro)
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error)

	// ListVisible obtiene comunicados visibles para un usuario según rol y fecha
	ListVisible(ctx context.Context, condominioID, userID, userRole string, page, pageSize int) ([]*models.Communication, int, error)

	// Create crea un nuevo comunicado
	Create(ctx context.Context, communication *models.Communication) error

	// Update actualiza un comunicado
	Update(ctx context.Context, communication *models.Communication) error

	// Delete elimina un comunicado (multi-tenant)
	Delete(ctx context.Context, communicationID, condominioID string) error

	// MarkRead marca un comunicado como leído por un usuario
	MarkRead(ctx context.Context, communicationID, userID string) error

	// UnreadCount retorna la cantidad de comunicados sin leer para un usuario
	UnreadCount(ctx context.Context, condominioID, userID, userRole string) (int, error)

	// ListComments obtiene los comentarios de un comunicado
	ListComments(ctx context.Context, communicationID string) ([]*models.ComunicadoComentario, error)

	// CreateComment crea un comentario en un comunicado
	CreateComment(ctx context.Context, comment *models.ComunicadoComentario) error
}

// PackageRepository interfaz para operaciones con paquetes
type PackageRepository interface {
	FindByID(ctx context.Context, packageID, condominioID string) (*models.Package, error)
	ListWithFilters(ctx context.Context, condominioID string, f dto.PackageFilter) ([]*models.Package, int, error)
	GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Package, int, error)
	GetByApartment(ctx context.Context, condominioID, apartmentID string, page, pageSize int) ([]*models.Package, int, error)
	Create(ctx context.Context, pkg *models.Package) error
	Update(ctx context.Context, packageID, condominioID string, fields map[string]interface{}) (*models.Package, error)
	Delete(ctx context.Context, packageID, condominioID string) error
	MarkDelivered(ctx context.Context, packageID, condominioID string) error
}

// CondominioRepository interfaz para operaciones con condominios
type CondominioRepository interface {
	// FindByID busca un condominio por ID
	FindByID(ctx context.Context, condominioID string) (*models.Condominio, error)

	// GetAll obtiene todos los condominios
	GetAll(ctx context.Context) ([]*models.Condominio, error)

	// Update actualiza la información de un condominio
	Update(ctx context.Context, condominio *models.Condominio) error
}
