package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"
)

// ApartmentService define las operaciones de negocio para apartamentos
type ApartmentService interface {
	Create(ctx context.Context, req *dto.ApartmentCreateRequest, condominioID string) (*dto.ApartmentDTO, error)
	GetByID(ctx context.Context, id, condominioID string) (*dto.ApartmentDTO, error)
	List(ctx context.Context, condominioID, userID string, page, pageSize int) (*dto.ApartmentsListResponse, error)
	Update(ctx context.Context, id string, req *dto.ApartmentUpdateRequest, condominioID string) (*dto.ApartmentDTO, error)
	Delete(ctx context.Context, id, condominioID string) error
}

// ApartmentServiceImpl implementación del servicio
type ApartmentServiceImpl struct {
	repo   repositories.ApartmentRepository
	logger logger.Logger
}

// NewApartmentService crea una nueva instancia del servicio
func NewApartmentService(repo repositories.ApartmentRepository, log logger.Logger) ApartmentService {
	return &ApartmentServiceImpl{
		repo:   repo,
		logger: log,
	}
}

// Create crea un nuevo apartamento
func (s *ApartmentServiceImpl) Create(ctx context.Context, req *dto.ApartmentCreateRequest, condominioID string) (*dto.ApartmentDTO, error) {
	// Validar
	if req.Numero == "" {
		return nil, fmt.Errorf("numero is required")
	}
	if req.Area < 0 {
		return nil, fmt.Errorf("area must be > 0")
	}

	piso := fmt.Sprintf("%d", req.Piso)
	propietarioID := req.PropietarioID
	areaPrivada := req.Area
	apartment := &models.Apartment{
		ID:            uuid.New().String(),
		CondominioID:  condominioID,
		Numero:        req.Numero,
		Piso:          &piso,
		PropietarioID: &propietarioID,
		AreaPrivada:   &areaPrivada,
		Estado:        req.Estado,
	}

	if err := s.repo.Create(ctx, apartment); err != nil {
		s.logger.WithError(err).Error("error creating apartment")
		return nil, err
	}

	return s.modelToDTO(apartment, ""), nil
}

// GetByID obtiene un apartamento por ID
func (s *ApartmentServiceImpl) GetByID(ctx context.Context, id, condominioID string) (*dto.ApartmentDTO, error) {
	apartment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("apartment_id", id).Error("error getting apartment")
		return nil, err
	}

	// Verificar que pertenece al condominio
	if apartment.CondominioID != condominioID {
		s.logger.WithField("apartment_id", id).Warn("unauthorized access to apartment")
		return nil, fmt.Errorf("apartment not found")
	}

	return s.modelToDTO(apartment, ""), nil
}

// List obtiene una lista de apartamentos del condominio
func (s *ApartmentServiceImpl) List(ctx context.Context, condominioID, userID string, page, pageSize int) (*dto.ApartmentsListResponse, error) {
	// Validar paginación
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var (
		apartments []*models.Apartment
		total      int
		err        error
	)

	if userID != "" {
		apartments, total, err = s.repo.GetByOwner(ctx, condominioID, userID, page, pageSize)
	} else {
		apartments, total, err = s.repo.GetByCondominio(ctx, condominioID, page, pageSize)
	}
	if err != nil {
		s.logger.WithError(err).Error("error listing apartments")
		return nil, err
	}

	dtos := make([]dto.ApartmentDTO, len(apartments))
	for i, apt := range apartments {
		dtos[i] = *s.modelToDTO(apt, "")
	}

	return &dto.ApartmentsListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     dtos,
	}, nil
}

// Update actualiza un apartamento
func (s *ApartmentServiceImpl) Update(ctx context.Context, id string, req *dto.ApartmentUpdateRequest, condominioID string) (*dto.ApartmentDTO, error) {
	// Verificar que existe
	apartment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("apartment_id", id).Error("error getting apartment for update")
		return nil, err
	}

	// Verificar que pertenece al condominio
	if apartment.CondominioID != condominioID {
		s.logger.WithField("apartment_id", id).Warn("unauthorized update attempt")
		return nil, fmt.Errorf("apartment not found")
	}

	// Actualizar campos si están presentes
	if req.Numero != "" {
		apartment.Numero = req.Numero
	}
	if req.Estado != "" {
		apartment.Estado = req.Estado
	}
	if req.Piso >= 0 {
		piso := fmt.Sprintf("%d", req.Piso)
		apartment.Piso = &piso
	}
	if req.PropietarioID != "" {
		propietarioID := req.PropietarioID
		apartment.PropietarioID = &propietarioID
	}
	if req.Area > 0 {
		apartment.AreaPrivada = &req.Area
	}

	if err := s.repo.Update(ctx, apartment); err != nil {
		s.logger.WithError(err).WithField("apartment_id", id).Error("error updating apartment")
		return nil, err
	}

	return s.modelToDTO(apartment, ""), nil
}

// Delete elimina un apartamento
func (s *ApartmentServiceImpl) Delete(ctx context.Context, id, condominioID string) error {
	// Verificar que existe
	apartment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("apartment_id", id).Error("error getting apartment for delete")
		return err
	}

	// Verificar que pertenece al condominio
	if apartment.CondominioID != condominioID {
		s.logger.WithField("apartment_id", id).Warn("unauthorized delete attempt")
		return fmt.Errorf("apartment not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).WithField("apartment_id", id).Error("error deleting apartment")
		return err
	}

	return nil
}

// modelToDTO convierte un modelo a DTO
func (s *ApartmentServiceImpl) modelToDTO(apartment *models.Apartment, propietarioNombre string) *dto.ApartmentDTO {
	var piso int
	if apartment.Piso != nil {
		pisoInt, _ := strconv.Atoi(*apartment.Piso)
		piso = pisoInt
	}

	var torre string
	if apartment.Torre != nil {
		torre = *apartment.Torre
	}

	var bloque string
	if apartment.Bloque != nil {
		bloque = *apartment.Bloque
	}

	var propietarioID string
	if apartment.PropietarioID != nil {
		propietarioID = *apartment.PropietarioID
	}

	var area float64
	if apartment.AreaPrivada != nil {
		area = *apartment.AreaPrivada
	}

	return &dto.ApartmentDTO{
		ID:                apartment.ID,
		Numero:            apartment.Numero,
		Area:              area,
		Estado:            apartment.Estado,
		Piso:              piso,
		Torre:             torre,
		Bloque:            bloque,
		PropietarioID:     propietarioID,
		PropietarioNombre: propietarioNombre,
		CondominioID:      apartment.CondominioID,
		CreatedAt:         apartment.FechaRegistro,
		UpdatedAt:         apartment.FechaRegistro,
	}
}
