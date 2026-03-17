package services

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/dto"
	pkgerrors "neighborhood-api/pkg/errors"
)

// ReservationService interfaz para operaciones con reservas
type ReservationService interface {
	Create(ctx context.Context, req *dto.CreateReservationRequest, condominioID, usuarioID string) (*dto.ReservationResponse, error)
	GetByID(ctx context.Context, reservationID, condominioID string) (*dto.ReservationResponse, error)
	List(ctx context.Context, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error)
	ListByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error)
	ListByEspacio(ctx context.Context, espacioID, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error)
	ListByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error)
	Update(ctx context.Context, reservationID string, req *dto.UpdateReservationRequest, condominioID string) (*dto.ReservationResponse, error)
	Delete(ctx context.Context, reservationID, condominioID string) error
	GetEspacioNombre(ctx context.Context, espacioID, condominioID string) (string, error)
}

// ReservationServiceImpl implementa ReservationService
type ReservationServiceImpl struct {
	reservationRepo repositories.ReservationRepository
	invoiceRepo     repositories.InvoiceRepository
}

// NewReservationService crea una nueva instancia del servicio
func NewReservationService(
	reservationRepo repositories.ReservationRepository,
	invoiceRepo repositories.InvoiceRepository,
) ReservationService {
	return &ReservationServiceImpl{
		reservationRepo: reservationRepo,
		invoiceRepo:     invoiceRepo,
	}
}

// Create crea una nueva reserva
func (s *ReservationServiceImpl) Create(ctx context.Context, req *dto.CreateReservationRequest, condominioID, usuarioID string) (*dto.ReservationResponse, error) {
	// Validaciones
	if req.EspacioID == "" {
		return nil, errors.New("espacio_id is required")
	}
	if req.PersonasEsperadas <= 0 {
		return nil, errors.New("personas_esperadas must be greater than 0")
	}
	if req.FechaInicio.IsZero() || req.FechaFin.IsZero() {
		return nil, errors.New("fecha_inicio and fecha_fin are required")
	}
	if req.FechaFin.Before(req.FechaInicio) {
		return nil, errors.New("fecha_fin must be after fecha_inicio")
	}

	// Validar solapamiento (excluye cancelados)
	exists, err := s.reservationRepo.ExistsOverlap(ctx, req.EspacioID, condominioID, req.FechaInicio, req.FechaFin)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, pkgerrors.ConflictErrorf("ya existe una reserva activa en ese rango horario")
	}

	// Obtener precio de referencia del espacio
	valorBase, err := s.reservationRepo.GetEspacioCostoHora(ctx, req.EspacioID, condominioID)
	if err != nil {
		return nil, err
	}

	horas := int(math.Ceil(req.FechaFin.Sub(req.FechaInicio).Hours()))
	if horas < 1 {
		horas = 1
	}

	var costoTotal *float64
	if valorBase != nil {
		total := *valorBase * float64(horas)
		costoTotal = &total
	}

	personasEsperadas := req.PersonasEsperadas
	reservation := &models.Reservation{
		ID:                uuid.New().String(),
		CondominioID:      condominioID,
		UsuarioID:         usuarioID,
		ApartmentID:       nil,
		EspacioID:         req.EspacioID,
		FechaSolicitud:    time.Now(),
		FechaInicio:       req.FechaInicio,
		FechaFin:          req.FechaFin,
		PersonasEsperadas: &personasEsperadas,
		Observaciones:     nil,
		ValorBase:         valorBase,
		CostoTotal:        costoTotal,
		Pagado:            false,
		Estado:            "pendiente",
	}

	// Crear en la BD
	err = s.reservationRepo.Create(ctx, reservation)
	if err != nil {
		return nil, err
	}

	return toReservationResponse(reservation), nil
}

// GetByID obtiene una reserva por ID
func (s *ReservationServiceImpl) GetByID(ctx context.Context, reservationID, condominioID string) (*dto.ReservationResponse, error) {
	if reservationID == "" {
		return nil, errors.New("reservation_id is required")
	}

	reservation, err := s.reservationRepo.FindByID(ctx, reservationID, condominioID)
	if err != nil {
		return nil, err
	}

	return toReservationResponse(reservation), nil
}

// List obtiene reservas de un condominio
func (s *ReservationServiceImpl) List(ctx context.Context, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error) {
	reservations, total, err := s.reservationRepo.GetByCondominio(ctx, condominioID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.ReservationResponse, len(reservations))
	for i, reservation := range reservations {
		responses[i] = toReservationResponse(reservation)
	}

	return responses, total, nil
}

// ListByUsuario obtiene reservas de un usuario
func (s *ReservationServiceImpl) ListByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error) {
	if usuarioID == "" {
		return nil, 0, errors.New("usuario_id is required")
	}

	reservations, total, err := s.reservationRepo.GetByUsuario(ctx, usuarioID, condominioID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.ReservationResponse, len(reservations))
	for i, reservation := range reservations {
		responses[i] = toReservationResponse(reservation)
	}

	return responses, total, nil
}

// ListByEspacio obtiene reservas de un espacio común
func (s *ReservationServiceImpl) ListByEspacio(ctx context.Context, espacioID, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error) {
	if espacioID == "" {
		return nil, 0, errors.New("espacio_id is required")
	}

	reservations, total, err := s.reservationRepo.GetByEspacio(ctx, espacioID, condominioID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.ReservationResponse, len(reservations))
	for i, reservation := range reservations {
		responses[i] = toReservationResponse(reservation)
	}

	return responses, total, nil
}

// ListByApartment obtiene reservas de todos los miembros de un apartamento
func (s *ReservationServiceImpl) ListByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*dto.ReservationResponse, int, error) {
	if apartmentID == "" {
		return nil, 0, errors.New("apartment_id is required")
	}

	reservations, total, err := s.reservationRepo.GetByApartment(ctx, apartmentID, condominioID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.ReservationResponse, len(reservations))
	for i, reservation := range reservations {
		responses[i] = toReservationResponse(reservation)
	}

	return responses, total, nil
}

// Update actualiza una reserva
func (s *ReservationServiceImpl) Update(ctx context.Context, reservationID string, req *dto.UpdateReservationRequest, condominioID string) (*dto.ReservationResponse, error) {
	if reservationID == "" {
		return nil, errors.New("reservation_id is required")
	}

	// Obtener reserva actual
	reservation, err := s.reservationRepo.FindByID(ctx, reservationID, condominioID)
	if err != nil {
		return nil, err
	}

	// Actualizar campos
	if req.FechaInicio != nil && !req.FechaInicio.IsZero() {
		reservation.FechaInicio = *req.FechaInicio
	}
	if req.FechaFin != nil && !req.FechaFin.IsZero() {
		reservation.FechaFin = *req.FechaFin
	}
	if req.PersonasEsperadas != nil && *req.PersonasEsperadas > 0 {
		reservation.PersonasEsperadas = req.PersonasEsperadas
	}
	prevStatus := reservation.Estado
	if req.Estado != nil && *req.Estado != "" {
		reservation.Estado = *req.Estado
	}

	// Validar que fin es después de inicio
	if reservation.FechaFin.Before(reservation.FechaInicio) {
		return nil, errors.New("fecha_fin must be after fecha_inicio")
	}

	// Actualizar en BD
	err = s.reservationRepo.Update(ctx, reservation)
	if err != nil {
		return nil, err
	}

	// Si pasó a confirmada y antes no lo estaba, generar factura
	if prevStatus != "confirmada" && reservation.Estado == "confirmada" {
		if err := s.createInvoiceForReservation(ctx, reservation); err != nil {
			// No bloquea la actualización, solo informa
			return toReservationResponse(reservation), nil
		}
	}

	return toReservationResponse(reservation), nil
}

func (s *ReservationServiceImpl) createInvoiceForReservation(ctx context.Context, reservation *models.Reservation) error {
	if reservation == nil {
		return nil
	}

	amount := 0.0
	if reservation.CostoTotal != nil {
		amount = *reservation.CostoTotal
	} else if reservation.ValorBase != nil {
		amount = *reservation.ValorBase
	}
	if amount <= 0 {
		return nil
	}

	due := reservation.FechaInicio.Add(30 * 24 * time.Hour)
	now := time.Now()
	invoiceNumber := generateInvoiceNumber()

	spaceName, err := s.reservationRepo.GetEspacioNombre(ctx, reservation.EspacioID, reservation.CondominioID)
	if err != nil || spaceName == "" {
		spaceName = reservation.EspacioID
	}

	invoice := &models.Invoice{
		ID:            uuid.New().String(),
		CondominioID:  reservation.CondominioID,
		UsuarioID:     reservation.UsuarioID,
		ApartmentID:   reservation.ApartmentID,
		ServiceID:     "", // evitar FK inválida; no hay servicio asociado
		InvoiceNumber: invoiceNumber,
		BaseAmount:    amount,
		Discount:      0,
		InterestRate:  0,
		Total:         amount,
		Status:        "pendiente",
		IssuedDate:    now,
		DueDate:       &due,
		PaymentMethod: nil,
		Notes:         buildInvoiceNote(reservation, spaceName),
	}

	return s.invoiceRepo.Create(ctx, invoice)
}

func generateInvoiceNumber() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	res := make([]rune, 6)
	for i := range res {
		b := make([]byte, 1)
		_, _ = cryptorand.Read(b)
		res[i] = letters[int(b[0])%len(letters)]
	}
	return "RES-" + string(res)
}

func buildInvoiceNote(res *models.Reservation, spaceName string) *string {
	if res == nil {
		return nil
	}
	start := res.FechaInicio.Format("2006-01-02 15:04")
	end := res.FechaFin.Format("2006-01-02 15:04")
	label := spaceName
	if label == "" {
		label = "espacio común"
	}
	note := "Factura generada por reserva del espacio " + label + " del " + start + " al " + end
	return &note
}

// GetEspacioNombre obtiene el nombre de un espacio común
func (s *ReservationServiceImpl) GetEspacioNombre(ctx context.Context, espacioID, condominioID string) (string, error) {
	name, err := s.reservationRepo.GetEspacioNombre(ctx, espacioID, condominioID)
	if err != nil || name == "" {
		return espacioID, err
	}
	return name, nil
}

// Delete elimina una reserva
func (s *ReservationServiceImpl) Delete(ctx context.Context, reservationID, condominioID string) error {
	if reservationID == "" {
		return errors.New("reservation_id is required")
	}

	err := s.reservationRepo.Delete(ctx, reservationID, condominioID)
	if err != nil {
		return err
	}

	return nil
}

// Helper function to convert model to DTO
func toReservationResponse(reservation *models.Reservation) *dto.ReservationResponse {
	var personasEsperadas int
	if reservation.PersonasEsperadas != nil {
		personasEsperadas = *reservation.PersonasEsperadas
	}

	return &dto.ReservationResponse{
		ID:                reservation.ID,
		EspacioID:         reservation.EspacioID,
		EspacioNombre:     reservation.EspacioNombre,
		UsuarioID:         reservation.UsuarioID,
		UsuarioNombre:     reservation.UsuarioNombre,
		ValorBase:         reservation.ValorBase,
		CostoTotal:        reservation.CostoTotal,
		FechaSolicitud:    reservation.FechaSolicitud,
		FechaInicio:       reservation.FechaInicio,
		FechaFin:          reservation.FechaFin,
		PersonasEsperadas: personasEsperadas,
		Estado:            reservation.Estado,
		CondominioID:      reservation.CondominioID,
	}
}
