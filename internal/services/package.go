package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/repositories"
	"neighborhood-api/pkg/dto"
)

// PackageService define operaciones para paquetería
type PackageService interface {
	Create(ctx context.Context, req dto.CreatePackageRequest, condominioID string) (*models.Package, error)
	ListWithFilters(ctx context.Context, condominioID string, f dto.PackageFilter) ([]*models.Package, int, error)
	GetByID(ctx context.Context, packageID, condominioID string) (*models.Package, error)
	Update(ctx context.Context, packageID, condominioID, role string, req dto.UpdatePackageRequest) (*models.Package, error)
	Delete(ctx context.Context, packageID, condominioID string) error
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
	return s.repo.FindByID(ctx, pkg.ID, condominioID)
}

func (s *PackageServiceImpl) ListWithFilters(ctx context.Context, condominioID string, f dto.PackageFilter) ([]*models.Package, int, error) {
	return s.repo.ListWithFilters(ctx, condominioID, f)
}

func (s *PackageServiceImpl) GetByID(ctx context.Context, packageID, condominioID string) (*models.Package, error) {
	return s.repo.FindByID(ctx, packageID, condominioID)
}

// Update validates role-based field access and applies the update.
// empleado: only notes
// admin/administrador: notes, resident, carrier, apartment_id
func (s *PackageServiceImpl) Update(ctx context.Context, packageID, condominioID, role string, req dto.UpdatePackageRequest) (*models.Package, error) {
	isAdmin := strings.EqualFold(role, "administrador") || strings.EqualFold(role, "admin")

	fields := map[string]interface{}{}

	if req.Notes != nil {
		fields["notes"] = *req.Notes
	}

	if isAdmin {
		if req.Resident != nil {
			fields["resident"] = *req.Resident
		}
		if req.Carrier != nil {
			fields["carrier"] = *req.Carrier
		}
		if req.ApartmentID != nil {
			fields["apartment_id"] = *req.ApartmentID
		}
	}

	return s.repo.Update(ctx, packageID, condominioID, fields)
}

func (s *PackageServiceImpl) Delete(ctx context.Context, packageID, condominioID string) error {
	return s.repo.Delete(ctx, packageID, condominioID)
}

func (s *PackageServiceImpl) MarkDelivered(ctx context.Context, packageID, condominioID string) (*models.Package, error) {
	if err := s.repo.MarkDelivered(ctx, packageID, condominioID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, packageID, condominioID)
}

func (s *PackageServiceImpl) Notify(ctx context.Context, packageID, condominioID string) error {
	_, err := s.repo.FindByID(ctx, packageID, condominioID)
	return err
}
