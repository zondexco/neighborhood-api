package services

import (
	"context"
	"errors"
	"time"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
)

// CommunicationService define las operaciones del servicio de comunicados
type CommunicationService interface {
	Create(ctx context.Context, request interface{}, condominioID, autor string) (*models.Communication, error)
	GetByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error)
	List(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error)
	Update(ctx context.Context, request interface{}, communicationID, condominioID string) (*models.Communication, error)
	Delete(ctx context.Context, communicationID, condominioID string) error
}

// CommunicationServiceImpl implementa CommunicationService
type CommunicationServiceImpl struct {
	repo repositories.CommunicationRepository
}

// NewCommunicationService crea una nueva instancia del servicio
func NewCommunicationService(repo repositories.CommunicationRepository) CommunicationService {
	return &CommunicationServiceImpl{repo: repo}
}

// Create valida y crea un nuevo comunicado
func (s *CommunicationServiceImpl) Create(ctx context.Context, request interface{}, condominioID, autor string) (*models.Communication, error) {
	// Buscar el tipo de request (para soporte de múltiples DTOs)
	var titulo, contenido string
	var fecha time.Time

	switch req := request.(type) {
	case CreateCommunicationRequest:
		titulo = req.Titulo
		contenido = req.Contenido
		if req.Fecha != nil {
			fecha = *req.Fecha
		} else {
			fecha = time.Now()
		}
	default:
		return nil, errors.New("invalid request type")
	}

	// Validaciones
	if titulo == "" {
		return nil, errors.New("titulo is required")
	}
	if contenido == "" {
		return nil, errors.New("contenido is required")
	}

	communication := &models.Communication{
		Titulo:       titulo,
		Contenido:    contenido,
		Fecha:        fecha,
		Autor:        autor,
		CondominioID: condominioID,
	}

	err := s.repo.Create(ctx, communication)
	if err != nil {
		return nil, err
	}

	return communication, nil
}

// GetByID obtiene un comunicado por ID
func (s *CommunicationServiceImpl) GetByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error) {
	return s.repo.FindByID(ctx, communicationID, condominioID)
}

// List obtiene comunicados de un condominio
func (s *CommunicationServiceImpl) List(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.repo.GetByCondominio(ctx, condominioID, page, pageSize)
}

// Update valida y actualiza un comunicado
func (s *CommunicationServiceImpl) Update(ctx context.Context, request interface{}, communicationID, condominioID string) (*models.Communication, error) {
	// Obtener comunicado actual
	communication, err := s.repo.FindByID(ctx, communicationID, condominioID)
	if err != nil {
		return nil, err
	}

	// Buscar el tipo de request
	switch req := request.(type) {
	case UpdateCommunicationRequest:
		if req.Titulo != nil && *req.Titulo != "" {
			communication.Titulo = *req.Titulo
		}
		if req.Contenido != nil && *req.Contenido != "" {
			communication.Contenido = *req.Contenido
		}
		if req.Fecha != nil {
			communication.Fecha = *req.Fecha
		}
	default:
		return nil, errors.New("invalid request type")
	}

	err = s.repo.Update(ctx, communication)
	if err != nil {
		return nil, err
	}

	return communication, nil
}

// Delete elimina un comunicado
func (s *CommunicationServiceImpl) Delete(ctx context.Context, communicationID, condominioID string) error {
	return s.repo.Delete(ctx, communicationID, condominioID)
}

// Tipos internos para conversión de DTOs
type CreateCommunicationRequest struct {
	Titulo    string     `json:"titulo"`
	Contenido string     `json:"contenido"`
	Fecha     *time.Time `json:"fecha,omitempty"`
}

type UpdateCommunicationRequest struct {
	Titulo    *string    `json:"titulo,omitempty"`
	Contenido *string    `json:"contenido,omitempty"`
	Fecha     *time.Time `json:"fecha,omitempty"`
}

// toCommunicationResponse convierte un modelo a respuesta
func toCommunicationResponse(c *models.Communication) interface{} {
	return map[string]interface{}{
		"id":            c.ID,
		"titulo":        c.Titulo,
		"contenido":     c.Contenido,
		"fecha":         c.Fecha,
		"autor":         c.Autor,
		"id_condominio": c.CondominioID,
		"created_at":    c.CreatedAt,
		"updated_at":    c.UpdatedAt,
	}
}
