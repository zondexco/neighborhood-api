package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"
)

// InvoiceService interfaz para servicio de facturas
type InvoiceService interface {
	// Create crea una nueva factura
	Create(ctx context.Context, req *dto.CreateInvoiceRequest, condominioID, usuarioID string) (*dto.InvoiceResponse, error)

	// GetByID obtiene una factura por ID
	GetByID(ctx context.Context, invoiceID, condominioID string) (*dto.InvoiceResponse, error)

	// List obtiene facturas de un condominio
	List(ctx context.Context, condominioID, usuarioID string, page, pageSize int) ([]*dto.InvoiceResponse, int, error)

	// ListByApartment obtiene facturas de un apartamento
	ListByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*dto.InvoiceResponse, int, error)

	// Update actualiza una factura
	Update(ctx context.Context, invoiceID string, req *dto.UpdateInvoiceRequest, condominioID string) (*dto.InvoiceResponse, error)

	// Delete elimina una factura
	Delete(ctx context.Context, invoiceID, condominioID string) error
}

// InvoiceServiceImpl implementación de InvoiceService
type InvoiceServiceImpl struct {
	invoiceRepo repositories.InvoiceRepository
	log         logger.Logger
}

// NewInvoiceService crea una nueva instancia de InvoiceService
func NewInvoiceService(invoiceRepo repositories.InvoiceRepository) InvoiceService {
	return &InvoiceServiceImpl{
		invoiceRepo: invoiceRepo,
		log:         logger.Get(),
	}
}

// Create crea una nueva factura
func (s *InvoiceServiceImpl) Create(ctx context.Context, req *dto.CreateInvoiceRequest, condominioID, usuarioID string) (*dto.InvoiceResponse, error) {
	// Validaciones
	if req.ApartmentID == "" {
		return nil, errors.ValidationErrorf("apartment_id is required")
	}
	if req.InvoiceNumber == "" {
		return nil, errors.ValidationErrorf("invoice_number is required")
	}
	if req.Amount <= 0 {
		return nil, errors.ValidationErrorf("amount must be greater than 0")
	}

	apartmentID := req.ApartmentID
	invoice := &models.Invoice{
		ID:            uuid.New().String(),
		CondominioID:  condominioID,
		UsuarioID:     usuarioID,
		ApartmentID:   &apartmentID,
		InvoiceNumber: req.InvoiceNumber,
		BaseAmount:    req.Amount,
		Total:         req.Amount,
		Status:        "pendiente",
		DueDate:       &req.DueDate,
		IssuedDate:    req.IssuedDate,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		s.log.WithError(err).Error("Error creating invoice")
		return nil, err
	}

	return toInvoiceResponse(invoice), nil
}

// GetByID obtiene una factura por ID
func (s *InvoiceServiceImpl) GetByID(ctx context.Context, invoiceID, condominioID string) (*dto.InvoiceResponse, error) {
	if invoiceID == "" {
		return nil, errors.ValidationErrorf("invoice_id is required")
	}

	invoice, err := s.invoiceRepo.FindByID(ctx, invoiceID, condominioID)
	if err != nil {
		return nil, err
	}

	return toInvoiceResponse(invoice), nil
}

// List obtiene facturas de un condominio
func (s *InvoiceServiceImpl) List(ctx context.Context, condominioID, usuarioID string, page, pageSize int) ([]*dto.InvoiceResponse, int, error) {
	invoices, total, err := s.invoiceRepo.GetByUsuario(ctx, usuarioID, condominioID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = toInvoiceResponse(invoice)
	}

	return responses, total, nil
}

// ListByApartment obtiene facturas de un apartamento
func (s *InvoiceServiceImpl) ListByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*dto.InvoiceResponse, int, error) {
	if apartmentID == "" {
		return nil, 0, errors.ValidationErrorf("apartment_id is required")
	}

	invoices, total, err := s.invoiceRepo.GetByApartment(ctx, apartmentID, condominioID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = toInvoiceResponse(invoice)
	}

	return responses, total, nil
}

// Update actualiza una factura
func (s *InvoiceServiceImpl) Update(ctx context.Context, invoiceID string, req *dto.UpdateInvoiceRequest, condominioID string) (*dto.InvoiceResponse, error) {
	if invoiceID == "" {
		return nil, errors.ValidationErrorf("invoice_id is required")
	}

	// Obtener factura actual
	invoice, err := s.invoiceRepo.FindByID(ctx, invoiceID, condominioID)
	if err != nil {
		return nil, err
	}

	// Actualizar campos
	if req.Amount != nil && *req.Amount > 0 {
		invoice.BaseAmount = *req.Amount
		invoice.Total = *req.Amount
	}
	if req.Status != nil {
		invoice.Status = *req.Status
	}
	if req.DueDate != nil {
		invoice.DueDate = req.DueDate
	}
	if req.PaymentMethod != nil {
		invoice.PaymentMethod = req.PaymentMethod
	}
	if req.PaidDate != nil {
		invoice.PaidDate = req.PaidDate
	}
	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		s.log.WithError(err).Error("Error updating invoice")
		return nil, err
	}

	return toInvoiceResponse(invoice), nil
}

// Delete elimina una factura
func (s *InvoiceServiceImpl) Delete(ctx context.Context, invoiceID, condominioID string) error {
	if invoiceID == "" {
		return errors.ValidationErrorf("invoice_id is required")
	}

	return s.invoiceRepo.Delete(ctx, invoiceID, condominioID)
}

// Helpers

func toInvoiceResponse(invoice *models.Invoice) *dto.InvoiceResponse {
	var apartmentID string
	if invoice.ApartmentID != nil {
		apartmentID = *invoice.ApartmentID
	}

	var dueDate time.Time
	if invoice.DueDate != nil {
		dueDate = *invoice.DueDate
	}

	return &dto.InvoiceResponse{
		ID:            invoice.ID,
		CondominioID:  invoice.CondominioID,
		ApartmentID:   apartmentID,
		InvoiceNumber: invoice.InvoiceNumber,
		Description:   "",
		Amount:        invoice.BaseAmount,
		Status:        invoice.Status,
		DueDate:       dueDate,
		IssuedDate:    invoice.IssuedDate,
		PaidDate:      invoice.PaidDate,
		InvoiceType:   "",
		PaymentMethod: invoice.PaymentMethod,
		Notes:         invoice.Notes,
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
	}
}
