package services

import (
	"context"
	"errors"
	"time"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/dto"
)

// PackageService define operaciones para paquetería
type PackageService interface {
	Create(ctx context.Context, req dto.CreatePackageRequest, condominioID string) (*models.Package, error)
	List(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Package, int, error)
	ListByApartment(ctx context.Context, condominioID, apartmentID string, page, pageSize int) ([]*models.Package, int, error)
	MarkDelivered(ctx context.Context, packageID, condominioID string) (*models.Package, error)
	Notify(ctx context.Context, packageID, condominioID string) error
}

type PackageServiceImpl struct {
	repo repositories.PackageRepository
}

func NewPackageService(repo repositories.PackageRepository) PackageService {
	return &PackageServiceImpl{repo: repo}
}

func (s *PackageServiceImpl) Create(ctx context.Context, req dto.CreatePackageRequest, condominioID string) (*models.Package, error) {
	if req.ApartmentID == "" {
		return nil, errors.New("apartment_id is required")
	}
	if req.Resident == "" {
		return nil, errors.New("resident is required")
	}
	if req.Carrier == "" {
		return nil, errors.New("carrier is required")
	}

	pkg := &models.Package{
		ApartmentID:  req.ApartmentID,
		Resident:     req.Resident,
		Carrier:      req.Carrier,
		Notes:        req.Notes,
		CondominioID: condominioID,
		ReceivedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

func (s *PackageServiceImpl) List(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Package, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.GetByCondominio(ctx, condominioID, page, pageSize)
}

func (s *PackageServiceImpl) ListByApartment(ctx context.Context, condominioID, apartmentID string, page, pageSize int) ([]*models.Package, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.GetByApartment(ctx, condominioID, apartmentID, page, pageSize)
}

func (s *PackageServiceImpl) MarkDelivered(ctx context.Context, packageID, condominioID string) (*models.Package, error) {
	if err := s.repo.MarkDelivered(ctx, packageID, condominioID); err != nil {
		return nil, err
	}
	// Devolver el paquete actualizado
	return s.repo.FindByID(ctx, packageID, condominioID)
}

func (s *PackageServiceImpl) Notify(ctx context.Context, packageID, condominioID string) error {
	// A futuro: integrar con notificaciones push/email
	// Por ahora solo valida existencia
	_, err := s.repo.FindByID(ctx, packageID, condominioID)
	return err
}
