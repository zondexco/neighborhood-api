package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/dto"
)

// MockInvoiceRepository es un mock del repositorio de facturas
type MockInvoiceRepository struct {
	mock.Mock
}

func (m *MockInvoiceRepository) Create(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) FindByID(ctx context.Context, invoiceID, condominioID string) (*models.Invoice, error) {
	args := m.Called(ctx, invoiceID, condominioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	args := m.Called(ctx, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockInvoiceRepository) GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	args := m.Called(ctx, usuarioID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockInvoiceRepository) GetByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	args := m.Called(ctx, apartmentID, condominioID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockInvoiceRepository) Update(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) Delete(ctx context.Context, invoiceID, condominioID string) error {
	args := m.Called(ctx, invoiceID, condominioID)
	return args.Error(0)
}

// Test Create - Éxito
func TestInvoiceService_Create_Success(t *testing.T) {
	mockRepo := new(MockInvoiceRepository)
	service := NewInvoiceService(mockRepo)

	ctx := context.Background()
	condominioID := uuid.New().String()
	usuarioID := uuid.New().String()
	apartmentID := uuid.New().String()

	req := &dto.CreateInvoiceRequest{
		ApartmentID:   apartmentID,
		InvoiceNumber: "INV-2025-001",
		Description:   "Cuota de mantenimiento",
		Amount:        250000,
		DueDate:       time.Now().AddDate(0, 1, 0),
		IssuedDate:    time.Now(),
		InvoiceType:   "mantenimiento",
		PaymentMethod: nil,
		Notes:         nil,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(inv *models.Invoice) bool {
		return inv.InvoiceNumber == req.InvoiceNumber &&
			inv.CondominioID == condominioID &&
			inv.UsuarioID == usuarioID &&
			(inv.ApartmentID != nil && *inv.ApartmentID == apartmentID) &&
			inv.Status == "pendiente"
	})).Return(nil)

	invoice, err := service.Create(ctx, req, condominioID, usuarioID)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, req.InvoiceNumber, invoice.InvoiceNumber)
	assert.Equal(t, condominioID, invoice.CondominioID)
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(inv *models.Invoice) bool {
		return inv.InvoiceNumber == req.InvoiceNumber
	}))
}

// Test Create - Error validación (amount <= 0)
func TestInvoiceService_Create_InvalidAmount(t *testing.T) {
	mockRepo := new(MockInvoiceRepository)
	service := NewInvoiceService(mockRepo)

	ctx := context.Background()
	condominioID := uuid.New().String()

	req := &dto.CreateInvoiceRequest{
		ApartmentID:   uuid.New().String(),
		InvoiceNumber: "INV-2025-001",
		Description:   "Test",
		Amount:        0,
		DueDate:       time.Now().AddDate(0, 1, 0),
		IssuedDate:    time.Now(),
		InvoiceType:   "mantenimiento",
	}

	invoice, err := service.Create(ctx, req, condominioID, "")

	assert.Error(t, err)
	assert.Nil(t, invoice)
	mockRepo.AssertNotCalled(t, "Create")
}

// Test GetByID - Éxito
func TestInvoiceService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockInvoiceRepository)
	service := NewInvoiceService(mockRepo)

	ctx := context.Background()
	invoiceID := uuid.New().String()
	condominioID := uuid.New().String()
	apartmentID := uuid.New().String()

	expectedInvoice := &models.Invoice{
		ID:            invoiceID,
		CondominioID:  condominioID,
		UsuarioID:     uuid.New().String(),
		ApartmentID:   &apartmentID,
		ServiceID:     uuid.New().String(),
		InvoiceNumber: "INV-2025-001",
		BaseAmount:    250000,
		Discount:      0,
		InterestRate:  0,
		Total:         250000,
		Status:        "pendiente",
		IssuedDate:    time.Now(),
		DueDate:       &time.Time{},
		PaidDate:      nil,
		PaymentMethod: nil,
		Notes:         nil,
	}

	mockRepo.On("FindByID", ctx, invoiceID, condominioID).Return(expectedInvoice, nil)

	invoice, err := service.GetByID(ctx, invoiceID, condominioID)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, invoiceID, invoice.ID)
	mockRepo.AssertCalled(t, "FindByID", ctx, invoiceID, condominioID)
}

// Test List - Éxito
func TestInvoiceService_List_Success(t *testing.T) {
	mockRepo := new(MockInvoiceRepository)
	service := NewInvoiceService(mockRepo)

	ctx := context.Background()
	condominioID := uuid.New().String()
	usuarioID := uuid.New().String()
	apartmentID := uuid.New().String()

	expectedInvoices := []*models.Invoice{
		{
			ID:            uuid.New().String(),
			CondominioID:  condominioID,
			UsuarioID:     uuid.New().String(),
			ApartmentID:   &apartmentID,
			ServiceID:     uuid.New().String(),
			InvoiceNumber: "INV-2025-001",
			BaseAmount:    250000,
			Discount:      0,
			InterestRate:  0,
			Total:         250000,
			Status:        "pendiente",
			IssuedDate:    time.Now(),
			DueDate:       nil,
			PaidDate:      nil,
			PaymentMethod: nil,
			Notes:         nil,
		},
	}

	mockRepo.On("GetByUsuario", ctx, usuarioID, condominioID, 1, 10).Return(expectedInvoices, 1, nil)

	invoices, total, err := service.List(ctx, condominioID, usuarioID, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, invoices)
	assert.Equal(t, 1, len(invoices))
	assert.Equal(t, 1, total)
	mockRepo.AssertCalled(t, "GetByUsuario", ctx, usuarioID, condominioID, 1, 10)
}

// Test Delete - Éxito
func TestInvoiceService_Delete_Success(t *testing.T) {
	mockRepo := new(MockInvoiceRepository)
	service := NewInvoiceService(mockRepo)

	ctx := context.Background()
	invoiceID := uuid.New().String()
	condominioID := uuid.New().String()

	mockRepo.On("Delete", ctx, invoiceID, condominioID).Return(nil)

	err := service.Delete(ctx, invoiceID, condominioID)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Delete", ctx, invoiceID, condominioID)
}
