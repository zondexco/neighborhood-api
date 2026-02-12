package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/dto"
)

// SpaceService define operaciones de negocio para espacios comunes
type SpaceService interface {
	Create(ctx context.Context, req *dto.CreateSpaceRequest, condominioID string) (*dto.SpaceDTO, error)
	GetByID(ctx context.Context, id, condominioID string) (*dto.SpaceDTO, error)
	List(ctx context.Context, condominioID string, page, pageSize int, search string) (*dto.SpacesListResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSpaceRequest, condominioID string) (*dto.SpaceDTO, error)
	Delete(ctx context.Context, id, condominioID string) error
}

// SpaceServiceImpl implementación
type SpaceServiceImpl struct {
	repo repositories.SpaceRepository
}

// NewSpaceService crea nueva instancia
func NewSpaceService(repo repositories.SpaceRepository) SpaceService {
	return &SpaceServiceImpl{repo: repo}
}

// Create crea un espacio
func (s *SpaceServiceImpl) Create(ctx context.Context, req *dto.CreateSpaceRequest, condominioID string) (*dto.SpaceDTO, error) {
	if req == nil || strings.TrimSpace(req.Nombre) == "" {
		return nil, errors.New("nombre is required")
	}

	estado := "activo"
	if req.Estado != nil && strings.TrimSpace(*req.Estado) != "" {
		estado = strings.ToLower(strings.TrimSpace(*req.Estado))
	}

	space := &models.Space{
		ID:           uuid.New().String(),
		CondominioID: condominioID,
		Nombre:       strings.TrimSpace(req.Nombre),
		Descripcion:  req.Descripcion,
		CostoHora:    req.CostoHora,
		Estado:       estado,
	}

	if err := s.repo.Create(ctx, space); err != nil {
		return nil, err
	}

	return s.toDTO(space), nil
}

// GetByID obtiene un espacio
func (s *SpaceServiceImpl) GetByID(ctx context.Context, id, condominioID string) (*dto.SpaceDTO, error) {
	space, err := s.repo.FindByID(ctx, id, condominioID)
	if err != nil {
		return nil, err
	}
	return s.toDTO(space), nil
}

// List retorna espacios
func (s *SpaceServiceImpl) List(ctx context.Context, condominioID string, page, pageSize int, search string) (*dto.SpacesListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	spaces, total, err := s.repo.GetByCondominio(ctx, condominioID, page, pageSize, search)
	if err != nil {
		return nil, err
	}

	data := make([]dto.SpaceDTO, 0, len(spaces))
	for _, sp := range spaces {
		data = append(data, *s.toDTO(sp))
	}

	return &dto.SpacesListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

// Update actualiza un espacio
func (s *SpaceServiceImpl) Update(ctx context.Context, id string, req *dto.UpdateSpaceRequest, condominioID string) (*dto.SpaceDTO, error) {
	space, err := s.repo.FindByID(ctx, id, condominioID)
	if err != nil {
		return nil, err
	}

	if req.Nombre != nil && strings.TrimSpace(*req.Nombre) != "" {
		space.Nombre = strings.TrimSpace(*req.Nombre)
	}
	if req.Descripcion != nil {
		trimmed := strings.TrimSpace(*req.Descripcion)
		if trimmed == "" {
			space.Descripcion = nil
		} else {
			space.Descripcion = &trimmed
		}
	}
	if req.CostoHora != nil {
		space.CostoHora = req.CostoHora
	}
	if req.Estado != nil && strings.TrimSpace(*req.Estado) != "" {
		space.Estado = strings.ToLower(strings.TrimSpace(*req.Estado))
	}

	if err := s.repo.Update(ctx, space); err != nil {
		return nil, err
	}

	return s.toDTO(space), nil
}

// Delete elimina un espacio
func (s *SpaceServiceImpl) Delete(ctx context.Context, id, condominioID string) error {
	// ensure exists
	if _, err := s.repo.FindByID(ctx, id, condominioID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, condominioID)
}

func (s *SpaceServiceImpl) toDTO(space *models.Space) *dto.SpaceDTO {
	if space == nil {
		return nil
	}
	return &dto.SpaceDTO{
		ID:           space.ID,
		Nombre:       space.Nombre,
		Descripcion:  space.Descripcion,
		CostoHora:    space.CostoHora,
		Estado:       space.Estado,
		CondominioID: space.CondominioID,
		CreatedAt:    space.CreatedAt,
		UpdatedAt:    space.UpdatedAt,
	}
}
