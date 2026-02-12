package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"
)

// MockApartmentRepository es un mock del repositorio
type MockApartmentRepository struct {
	mock.Mock
}

func (m *MockApartmentRepository) FindByID(ctx context.Context, apartmentID string) (*models.Apartment, error) {
	args := m.Called(ctx, apartmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Apartment), args.Error(1)
}

func (m *MockApartmentRepository) Create(ctx context.Context, apartment *models.Apartment) error {
	args := m.Called(ctx, apartment)
	return args.Error(0)
}

func (m *MockApartmentRepository) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Apartment, int, error) {
	args := m.Called(ctx, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Apartment), args.Int(1), args.Error(2)
}

func (m *MockApartmentRepository) GetByPropietario(ctx context.Context, propietarioID string) ([]*models.Apartment, error) {
	args := m.Called(ctx, propietarioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Apartment), args.Error(1)
}

func (m *MockApartmentRepository) GetByOwner(ctx context.Context, condominioID, ownerID string, page, pageSize int) ([]*models.Apartment, int, error) {
	args := m.Called(ctx, condominioID, ownerID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Apartment), args.Int(1), args.Error(2)
}

func (m *MockApartmentRepository) Update(ctx context.Context, apartment *models.Apartment) error {
	args := m.Called(ctx, apartment)
	return args.Error(0)
}

func (m *MockApartmentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test Create - Éxito
func TestApartmentService_Create_Success(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	condominioID := uuid.New().String()

	req := &dto.ApartmentCreateRequest{
		Numero:        "301",
		Area:          85.5,
		Estado:        "ocupado",
		Piso:          3,
		PropietarioID: uuid.New().String(),
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(apt *models.Apartment) bool {
		return apt.Numero == req.Numero && apt.CondominioID == condominioID
	})).Return(nil)

	result, err := service.Create(ctx, req, condominioID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Numero, result.Numero)
	assert.Equal(t, req.Area, result.Area)
	assert.Equal(t, condominioID, result.CondominioID)
	mockRepo.AssertExpectations(t)
}

// Test Create - Validación fallida (número vacío)
func TestApartmentService_Create_ValidationFailed_EmptyNumero(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	condominioID := uuid.New().String()

	req := &dto.ApartmentCreateRequest{
		Numero: "", // Inválido
		Area:   85.5,
		Estado: "ocupado",
		Piso:   3,
	}

	result, err := service.Create(ctx, req, condominioID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "numero is required")
}

// Test Create - Validación fallida (área negativa)
func TestApartmentService_Create_ValidationFailed_NegativeArea(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	condominioID := uuid.New().String()

	req := &dto.ApartmentCreateRequest{
		Numero: "301",
		Area:   -10, // Inválido
		Estado: "ocupado",
		Piso:   3,
	}

	result, err := service.Create(ctx, req, condominioID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "area must be > 0")
}

// Test Create - Error del repositorio
func TestApartmentService_Create_RepositoryError(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	condominioID := uuid.New().String()

	req := &dto.ApartmentCreateRequest{
		Numero: "301",
		Area:   85.5,
		Estado: "ocupado",
		Piso:   3,
	}

	mockRepo.On("Create", ctx, mock.Anything).Return(errors.New("database error"))

	result, err := service.Create(ctx, req, condominioID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// Test GetByID - Éxito
func TestApartmentService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()

	bloque := "A"
	torre := "1"
	piso := "3"
	propietarioID := uuid.New().String()
	areaPrivada := 85.5
	areaComun := 15.0

	mockApartment := &models.Apartment{
		ID:            apartmentID,
		CondominioID:  condominioID,
		Numero:        "301",
		Bloque:        &bloque,
		Torre:         &torre,
		Piso:          &piso,
		PropietarioID: &propietarioID,
		AreaPrivada:   &areaPrivada,
		AreaComun:     &areaComun,
		FechaRegistro: time.Now(),
		Estado:        "activo",
	}

	mockRepo.On("FindByID", ctx, apartmentID).Return(mockApartment, nil)

	result, err := service.GetByID(ctx, apartmentID, condominioID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, apartmentID, result.ID)
	assert.Equal(t, "301", result.Numero)
	mockRepo.AssertExpectations(t)
}

// Test GetByID - Apartamento no encontrado
func TestApartmentService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()

	mockRepo.On("FindByID", ctx, apartmentID).Return(nil, errors.New("apartment not found"))

	result, err := service.GetByID(ctx, apartmentID, condominioID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// Test GetByID - Acceso no autorizado (condominio diferente)
func TestApartmentService_GetByID_UnauthorizedAccess(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()
	differentCondominioID := uuid.New().String()

	mockApartment := &models.Apartment{
		ID:           apartmentID,
		Numero:       "301",
		CondominioID: condominioID,
		Estado:       "activo",
	}

	mockRepo.On("FindByID", ctx, apartmentID).Return(mockApartment, nil)

	result, err := service.GetByID(ctx, apartmentID, differentCondominioID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "apartment not found")
}

// Test List - Éxito
func TestApartmentService_List_Success(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	condominioID := uuid.New().String()

	mockApartments := []*models.Apartment{
		{
			ID:           uuid.New().String(),
			Numero:       "301",
			CondominioID: condominioID,
			Estado:       "activo",
		},
		{
			ID:           uuid.New().String(),
			Numero:       "302",
			CondominioID: condominioID,
			Estado:       "activo",
		},
	}

	mockRepo.On("GetByCondominio", ctx, condominioID, 1, 10).Return(mockApartments, 2, nil)

	result, err := service.List(ctx, condominioID, "", 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, len(result.Data))
	mockRepo.AssertExpectations(t)
}

// Test List - Paginación por defecto
func TestApartmentService_List_DefaultPagination(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	condominioID := uuid.New().String()

	mockRepo.On("GetByCondominio", ctx, condominioID, 1, 10).Return([]*models.Apartment{}, 0, nil)

	result, err := service.List(ctx, condominioID, "", 0, 0) // Valores por defecto

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
	mockRepo.AssertExpectations(t)
}

// Test Update - Éxito
func TestApartmentService_Update_Success(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()

	estado := "ocupado"
	existingApartment := &models.Apartment{
		ID:           apartmentID,
		Numero:       "301",
		Estado:       estado,
		CondominioID: condominioID,
	}

	updateReq := &dto.ApartmentUpdateRequest{
		Area:   90.0,
		Estado: "desocupado",
	}

	mockRepo.On("FindByID", ctx, apartmentID).Return(existingApartment, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(apt *models.Apartment) bool {
		return apt.Estado == "desocupado"
	})).Return(nil)

	result, err := service.Update(ctx, apartmentID, updateReq, condominioID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 90.0, result.Area)
	assert.Equal(t, "desocupado", result.Estado)
	mockRepo.AssertExpectations(t)
}

// Test Update - Acceso no autorizado
func TestApartmentService_Update_UnauthorizedAccess(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()
	differentCondominioID := uuid.New().String()

	existingApartment := &models.Apartment{
		ID:            apartmentID,
		Numero:        "301",
		CondominioID:  condominioID, // Diferente
		FechaRegistro: time.Now(),
		Estado:        "activo",
	}

	updateReq := &dto.ApartmentUpdateRequest{
		Area: 90.0,
	}

	mockRepo.On("FindByID", ctx, apartmentID).Return(existingApartment, nil)

	result, err := service.Update(ctx, apartmentID, updateReq, differentCondominioID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// Test Delete - Éxito
func TestApartmentService_Delete_Success(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()

	mockApartment := &models.Apartment{
		ID:            apartmentID,
		Numero:        "301",
		CondominioID:  condominioID,
		FechaRegistro: time.Now(),
		Estado:        "activo",
	}

	mockRepo.On("FindByID", ctx, apartmentID).Return(mockApartment, nil)
	mockRepo.On("Delete", ctx, apartmentID).Return(nil)

	err := service.Delete(ctx, apartmentID, condominioID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test Delete - Apartamento no encontrado
func TestApartmentService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()

	mockRepo.On("FindByID", ctx, apartmentID).Return(nil, errors.New("not found"))

	err := service.Delete(ctx, apartmentID, condominioID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// Test Delete - Acceso no autorizado
func TestApartmentService_Delete_UnauthorizedAccess(t *testing.T) {
	mockRepo := new(MockApartmentRepository)
	log := logger.Get()
	service := NewApartmentService(mockRepo, log)

	ctx := context.Background()
	apartmentID := uuid.New().String()
	condominioID := uuid.New().String()
	differentCondominioID := uuid.New().String()

	mockApartment := &models.Apartment{
		ID:            apartmentID,
		CondominioID:  condominioID,
		FechaRegistro: time.Now(),
		Estado:        "activo",
	}

	mockRepo.On("FindByID", ctx, apartmentID).Return(mockApartment, nil)

	err := service.Delete(ctx, apartmentID, differentCondominioID)

	assert.Error(t, err)
}
