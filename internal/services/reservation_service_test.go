package services

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/dto"
)

// MockReservationRepository es un mock del repositorio de reservas
type MockReservationRepository struct {
	mock.Mock
}

type MockInvoiceRepoForReservation struct {
	mock.Mock
}

func (m *MockInvoiceRepoForReservation) Create(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepoForReservation) FindByID(ctx context.Context, invoiceID, condominioID string) (*models.Invoice, error) {
	args := m.Called(ctx, invoiceID, condominioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockInvoiceRepoForReservation) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	args := m.Called(ctx, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockInvoiceRepoForReservation) GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	args := m.Called(ctx, usuarioID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockInvoiceRepoForReservation) GetByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	args := m.Called(ctx, apartmentID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockInvoiceRepoForReservation) Update(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepoForReservation) Delete(ctx context.Context, invoiceID, condominioID string) error {
	args := m.Called(ctx, invoiceID, condominioID)
	return args.Error(0)
}

func (m *MockReservationRepository) Create(ctx context.Context, reservation *models.Reservation) error {
	args := m.Called(ctx, reservation)
	return args.Error(0)
}

func (m *MockReservationRepository) FindByID(ctx context.Context, reservationID, condominioID string) (*models.Reservation, error) {
	args := m.Called(ctx, reservationID, condominioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetEspacioCostoHora(ctx context.Context, espacioID, condominioID string) (*float64, error) {
	args := m.Called(ctx, espacioID, condominioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*float64), args.Error(1)
}

func (m *MockReservationRepository) GetEspacioNombre(ctx context.Context, espacioID, condominioID string) (string, error) {
	args := m.Called(ctx, espacioID, condominioID)
	return args.String(0), args.Error(1)
}

func (m *MockReservationRepository) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	args := m.Called(ctx, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Reservation), args.Int(1), args.Error(2)
}

func (m *MockReservationRepository) GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	args := m.Called(ctx, usuarioID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Reservation), args.Int(1), args.Error(2)
}

func (m *MockReservationRepository) GetByEspacio(ctx context.Context, espacioID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	args := m.Called(ctx, espacioID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Reservation), args.Int(1), args.Error(2)
}

func (m *MockReservationRepository) GetByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	args := m.Called(ctx, apartmentID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Reservation), args.Int(1), args.Error(2)
}

func (m *MockReservationRepository) ExistsOverlap(ctx context.Context, espacioID, condominioID string, start, end time.Time) (bool, error) {
	args := m.Called(ctx, espacioID, condominioID, start, end)
	return args.Bool(0), args.Error(1)
}

func (m *MockReservationRepository) Update(ctx context.Context, reservation *models.Reservation) error {
	args := m.Called(ctx, reservation)
	return args.Error(0)
}

func (m *MockReservationRepository) Delete(ctx context.Context, reservationID, condominioID string) error {
	args := m.Called(ctx, reservationID, condominioID)
	return args.Error(0)
}

// Test Create - Éxito
func TestReservationService_Create_Success(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	condominioID := uuid.New().String()
	usuarioID := uuid.New().String()
	costo := 25.0

	req := &dto.CreateReservationRequest{
		EspacioID:         uuid.New().String(),
		FechaInicio:       time.Now().AddDate(0, 0, 1),
		FechaFin:          time.Now().AddDate(0, 0, 2),
		PersonasEsperadas: 5,
	}

	hours := int(math.Ceil(req.FechaFin.Sub(req.FechaInicio).Hours()))
	if hours < 1 {
		hours = 1
	}
	expectedTotal := costo * float64(hours)

	mockRepo.On("GetEspacioCostoHora", ctx, req.EspacioID, condominioID).Return(&costo, nil)
	mockRepo.On("ExistsOverlap", ctx, req.EspacioID, condominioID, req.FechaInicio, req.FechaFin).Return(false, nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(res *models.Reservation) bool {
		return res.EspacioID == req.EspacioID &&
			res.CondominioID == condominioID &&
			(res.PersonasEsperadas != nil && *res.PersonasEsperadas == req.PersonasEsperadas) &&
			res.Estado == "pendiente" &&
			res.UsuarioID == usuarioID &&
			res.ValorBase != nil && *res.ValorBase == costo &&
			res.CostoTotal != nil && *res.CostoTotal == expectedTotal
	})).Return(nil)

	reservation, err := service.Create(ctx, req, condominioID, usuarioID)

	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, req.EspacioID, reservation.EspacioID)
	assert.Equal(t, "pendiente", reservation.Estado)
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(res *models.Reservation) bool {
		return res.EspacioID == req.EspacioID
	}))
	mockRepo.AssertCalled(t, "GetEspacioCostoHora", ctx, req.EspacioID, condominioID)
}

// Test Create - Error validación (personas_esperadas <= 0)
func TestReservationService_Create_InvalidPersonas(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	condominioID := uuid.New().String()
	usuarioID := uuid.New().String()

	req := &dto.CreateReservationRequest{
		EspacioID:         uuid.New().String(),
		FechaInicio:       time.Now().AddDate(0, 0, 1),
		FechaFin:          time.Now().AddDate(0, 0, 2),
		PersonasEsperadas: 0,
	}

	reservation, err := service.Create(ctx, req, condominioID, usuarioID)

	assert.Error(t, err)
	assert.Nil(t, reservation)
	mockRepo.AssertNotCalled(t, "Create")
}

// Test Create - Error validación (fecha_fin antes de fecha_inicio)
func TestReservationService_Create_InvalidDates(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	condominioID := uuid.New().String()
	usuarioID := uuid.New().String()

	req := &dto.CreateReservationRequest{
		EspacioID:         uuid.New().String(),
		FechaInicio:       time.Now().AddDate(0, 0, 2),
		FechaFin:          time.Now().AddDate(0, 0, 1),
		PersonasEsperadas: 5,
	}

	reservation, err := service.Create(ctx, req, condominioID, usuarioID)

	assert.Error(t, err)
	assert.Nil(t, reservation)
	mockRepo.AssertNotCalled(t, "Create")
}

// Test GetByID - Éxito
func TestReservationService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	reservationID := uuid.New().String()
	condominioID := uuid.New().String()

	expectedReservation := &models.Reservation{
		ID:                reservationID,
		CondominioID:      condominioID,
		UsuarioID:         uuid.New().String(),
		ApartmentID:       nil,
		EspacioID:         uuid.New().String(),
		FechaSolicitud:    time.Now(),
		FechaInicio:       time.Now().AddDate(0, 0, 1),
		FechaFin:          time.Now().AddDate(0, 0, 2),
		PersonasEsperadas: nil,
		Observaciones:     nil,
		CostoTotal:        nil,
		Pagado:            false,
		Estado:            "confirmada",
	}

	mockRepo.On("FindByID", ctx, reservationID, condominioID).Return(expectedReservation, nil)

	reservation, err := service.GetByID(ctx, reservationID, condominioID)

	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, reservationID, reservation.ID)
	mockRepo.AssertCalled(t, "FindByID", ctx, reservationID, condominioID)
}

// Test List - Éxito
func TestReservationService_List_Success(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	condominioID := uuid.New().String()

	expectedReservations := []*models.Reservation{
		{
			ID:                uuid.New().String(),
			CondominioID:      condominioID,
			UsuarioID:         uuid.New().String(),
			ApartmentID:       nil,
			EspacioID:         uuid.New().String(),
			FechaSolicitud:    time.Now(),
			FechaInicio:       time.Now().AddDate(0, 0, 1),
			FechaFin:          time.Now().AddDate(0, 0, 2),
			PersonasEsperadas: nil,
			Observaciones:     nil,
			CostoTotal:        nil,
			Pagado:            false,
			Estado:            "confirmada",
		},
	}

	mockRepo.On("GetByCondominio", ctx, condominioID, 1, 10).Return(expectedReservations, 1, nil)

	reservations, total, err := service.List(ctx, condominioID, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, reservations)
	assert.Equal(t, 1, len(reservations))
	assert.Equal(t, 1, total)
	mockRepo.AssertCalled(t, "GetByCondominio", ctx, condominioID, 1, 10)
}

// Test Update - Éxito
func TestReservationService_Update_Success(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	reservationID := uuid.New().String()
	condominioID := uuid.New().String()

	originalReservation := &models.Reservation{
		ID:                reservationID,
		CondominioID:      condominioID,
		UsuarioID:         uuid.New().String(),
		ApartmentID:       nil,
		EspacioID:         uuid.New().String(),
		FechaSolicitud:    time.Now(),
		FechaInicio:       time.Now().AddDate(0, 0, 1),
		FechaFin:          time.Now().AddDate(0, 0, 2),
		PersonasEsperadas: nil,
		Observaciones:     nil,
		CostoTotal:        nil,
		Pagado:            false,
		Estado:            "confirmada",
	}

	newPersonas := 10
	req := &dto.UpdateReservationRequest{
		PersonasEsperadas: &newPersonas,
	}

	mockRepo.On("FindByID", ctx, reservationID, condominioID).Return(originalReservation, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(res *models.Reservation) bool {
		return res.ID == reservationID && (res.PersonasEsperadas != nil && *res.PersonasEsperadas == newPersonas)
	})).Return(nil)

	reservation, err := service.Update(ctx, reservationID, req, condominioID)

	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.Equal(t, newPersonas, reservation.PersonasEsperadas)
	mockRepo.AssertCalled(t, "FindByID", ctx, reservationID, condominioID)
	mockRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(res *models.Reservation) bool {
		return res.PersonasEsperadas != nil && *res.PersonasEsperadas == newPersonas
	}))
}

// Test Delete - Éxito
func TestReservationService_Delete_Success(t *testing.T) {
	mockRepo := new(MockReservationRepository)
	mockInvoice := new(MockInvoiceRepoForReservation)
	service := NewReservationService(mockRepo, mockInvoice)

	ctx := context.Background()
	reservationID := uuid.New().String()
	condominioID := uuid.New().String()

	mockRepo.On("Delete", ctx, reservationID, condominioID).Return(nil)

	err := service.Delete(ctx, reservationID, condominioID)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Delete", ctx, reservationID, condominioID)
}
