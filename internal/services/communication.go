package services

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
)

// CommunicationService define las operaciones del servicio de comunicados
type CommunicationService interface {
	Create(ctx context.Context, request interface{}, condominioID, autor string) (*models.Communication, error)
	GetByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error)
	List(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error)
	ListVisible(ctx context.Context, condominioID, userID, userRole string, page, pageSize int) ([]*models.Communication, int, error)
	Update(ctx context.Context, request interface{}, communicationID, condominioID string) (*models.Communication, error)
	Delete(ctx context.Context, communicationID, condominioID string) error
	MarkRead(ctx context.Context, communicationID, userID string) error
	UnreadCount(ctx context.Context, condominioID, userID, userRole string) (int, error)
	ListComments(ctx context.Context, communicationID string) ([]*models.ComunicadoComentario, error)
	CreateComment(ctx context.Context, communicationID, userID, contenido string) (*models.ComunicadoComentario, error)
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
	var titulo, contenido, icono string
	var fecha time.Time
	var programadoPara *time.Time
	var rolesDestino pq.StringArray
	var permiteComentarios, publicado bool

	switch req := request.(type) {
	case CreateCommunicationRequest:
		titulo = req.Titulo
		contenido = req.Contenido
		icono = req.Icono
		permiteComentarios = req.PermiteComentarios
		publicado = true
		if req.Publicado != nil {
			publicado = *req.Publicado
		}
		if req.Fecha != nil {
			fecha = *req.Fecha
		} else {
			fecha = time.Now()
		}
		programadoPara = req.ProgramadoPara
		if len(req.RolesDestino) > 0 {
			rolesDestino = pq.StringArray(req.RolesDestino)
		}
	default:
		return nil, errors.New("invalid request type")
	}

	if titulo == "" {
		return nil, errors.New("titulo is required")
	}
	if contenido == "" {
		return nil, errors.New("contenido is required")
	}
	if icono == "" {
		icono = "megaphone"
	}

	communication := &models.Communication{
		Titulo:             titulo,
		Contenido:          contenido,
		Fecha:              fecha,
		Autor:              autor,
		CondominioID:       condominioID,
		ProgramadoPara:     programadoPara,
		RolesDestino:       rolesDestino,
		Icono:              icono,
		PermiteComentarios: permiteComentarios,
		Publicado:          publicado,
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

// List obtiene comunicados de un condominio (admin, sin filtro)
func (s *CommunicationServiceImpl) List(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.GetByCondominio(ctx, condominioID, page, pageSize)
}

// ListVisible obtiene comunicados visibles para un usuario
func (s *CommunicationServiceImpl) ListVisible(ctx context.Context, condominioID, userID, userRole string, page, pageSize int) ([]*models.Communication, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.ListVisible(ctx, condominioID, userID, userRole, page, pageSize)
}

// Update valida y actualiza un comunicado
func (s *CommunicationServiceImpl) Update(ctx context.Context, request interface{}, communicationID, condominioID string) (*models.Communication, error) {
	communication, err := s.repo.FindByID(ctx, communicationID, condominioID)
	if err != nil {
		return nil, err
	}

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
		if req.ProgramadoPara != nil {
			communication.ProgramadoPara = req.ProgramadoPara
		}
		if req.RolesDestino != nil {
			communication.RolesDestino = pq.StringArray(req.RolesDestino)
		}
		if req.Icono != nil {
			communication.Icono = *req.Icono
		}
		if req.PermiteComentarios != nil {
			communication.PermiteComentarios = *req.PermiteComentarios
		}
		if req.Publicado != nil {
			communication.Publicado = *req.Publicado
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

// MarkRead marca un comunicado como leído
func (s *CommunicationServiceImpl) MarkRead(ctx context.Context, communicationID, userID string) error {
	return s.repo.MarkRead(ctx, communicationID, userID)
}

// UnreadCount retorna la cantidad de comunicados sin leer
func (s *CommunicationServiceImpl) UnreadCount(ctx context.Context, condominioID, userID, userRole string) (int, error) {
	return s.repo.UnreadCount(ctx, condominioID, userID, userRole)
}

// ListComments obtiene los comentarios de un comunicado
func (s *CommunicationServiceImpl) ListComments(ctx context.Context, communicationID string) ([]*models.ComunicadoComentario, error) {
	return s.repo.ListComments(ctx, communicationID)
}

// CreateComment crea un comentario en un comunicado
func (s *CommunicationServiceImpl) CreateComment(ctx context.Context, communicationID, userID, contenido string) (*models.ComunicadoComentario, error) {
	if contenido == "" {
		return nil, errors.New("contenido is required")
	}
	comment := &models.ComunicadoComentario{
		IDComunicado: communicationID,
		IDUsuario:    userID,
		Contenido:    contenido,
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// Internal DTO types for service layer
type CreateCommunicationRequest struct {
	Titulo             string     `json:"titulo"`
	Contenido          string     `json:"contenido"`
	Fecha              *time.Time `json:"fecha,omitempty"`
	ProgramadoPara     *time.Time `json:"programado_para,omitempty"`
	RolesDestino       []string   `json:"roles_destino,omitempty"`
	Icono              string     `json:"icono"`
	PermiteComentarios bool       `json:"permite_comentarios"`
	Publicado          *bool      `json:"publicado,omitempty"`
}

type UpdateCommunicationRequest struct {
	Titulo             *string    `json:"titulo,omitempty"`
	Contenido          *string    `json:"contenido,omitempty"`
	Fecha              *time.Time `json:"fecha,omitempty"`
	ProgramadoPara     *time.Time `json:"programado_para,omitempty"`
	RolesDestino       []string   `json:"roles_destino,omitempty"`
	Icono              *string    `json:"icono,omitempty"`
	PermiteComentarios *bool      `json:"permite_comentarios,omitempty"`
	Publicado          *bool      `json:"publicado,omitempty"`
}
