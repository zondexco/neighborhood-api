package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
)

// PackageRepositoryImpl implementa PackageRepository
type PackageRepositoryImpl struct {
	db *database.DB
}

func NewPackageRepository(db *database.DB) PackageRepository {
	return &PackageRepositoryImpl{db: db}
}

func (r *PackageRepositoryImpl) Create(ctx context.Context, pkg *models.Package) error {
	if pkg.ID == "" {
		pkg.ID = uuid.New().String()
	}

	query := `
        INSERT INTO paquete (
            id_paquete, id_apartamento, residente, transportadora, notas, fecha_recibido, id_condominio
        ) VALUES (
            $1, $2, $3, $4, $5, NOW(), $6
        )
        RETURNING fecha_recibido, fecha_creacion, fecha_actualizacion
    `

	return r.db.QueryRowContext(
		ctx,
		query,
		pkg.ID,
		pkg.ApartmentID,
		pkg.Resident,
		pkg.Carrier,
		pkg.Notes,
		pkg.CondominioID,
	).Scan(&pkg.ReceivedAt, &pkg.CreatedAt, &pkg.UpdatedAt)
}

func (r *PackageRepositoryImpl) FindByID(ctx context.Context, packageID, condominioID string) (*models.Package, error) {
	query := `
        SELECT p.id_paquete, p.id_apartamento, p.residente, p.transportadora, p.notas,
               p.fecha_recibido, p.fecha_entregado, p.id_condominio, p.fecha_creacion, p.fecha_actualizacion,
               a.numero, a.torre, a.bloque, a.piso
        FROM paquete p
        LEFT JOIN apartamento a ON a.id_apartamento = p.id_apartamento
        WHERE p.id_paquete = $1 AND p.id_condominio = $2
    `

	var pkg models.Package
	err := r.db.QueryRowContext(ctx, query, packageID, condominioID).Scan(
		&pkg.ID,
		&pkg.ApartmentID,
		&pkg.Resident,
		&pkg.Carrier,
		&pkg.Notes,
		&pkg.ReceivedAt,
		&pkg.DeliveredAt,
		&pkg.CondominioID,
		&pkg.CreatedAt,
		&pkg.UpdatedAt,
		&pkg.ApartmentNumber,
		&pkg.ApartmentTower,
		&pkg.ApartmentBlock,
		&pkg.ApartmentFloor,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("package not found")
		}
		return nil, err
	}

	return &pkg, nil
}

func (r *PackageRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Package, int, error) {
	countQuery := `SELECT COUNT(*) FROM paquete WHERE id_condominio = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
        SELECT p.id_paquete, p.id_apartamento, p.residente, p.transportadora, p.notas,
               p.fecha_recibido, p.fecha_entregado, p.id_condominio, p.fecha_creacion, p.fecha_actualizacion,
               a.numero, a.torre, a.bloque, a.piso
        FROM paquete p
        LEFT JOIN apartamento a ON a.id_apartamento = p.id_apartamento
        WHERE p.id_condominio = $1
        ORDER BY p.fecha_recibido DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.QueryContext(ctx, query, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		var pkg models.Package
		if err := rows.Scan(
			&pkg.ID,
			&pkg.ApartmentID,
			&pkg.Resident,
			&pkg.Carrier,
			&pkg.Notes,
			&pkg.ReceivedAt,
			&pkg.DeliveredAt,
			&pkg.CondominioID,
			&pkg.CreatedAt,
			&pkg.UpdatedAt,
			&pkg.ApartmentNumber,
			&pkg.ApartmentTower,
			&pkg.ApartmentBlock,
			&pkg.ApartmentFloor,
		); err != nil {
			return nil, 0, err
		}
		packages = append(packages, &pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return packages, total, nil
}

func (r *PackageRepositoryImpl) GetByApartment(ctx context.Context, condominioID, apartmentID string, page, pageSize int) ([]*models.Package, int, error) {
	countQuery := `SELECT COUNT(*) FROM paquete WHERE id_condominio = $1 AND id_apartamento = $2`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, condominioID, apartmentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
        SELECT p.id_paquete, p.id_apartamento, p.residente, p.transportadora, p.notas,
               p.fecha_recibido, p.fecha_entregado, p.id_condominio, p.fecha_creacion, p.fecha_actualizacion,
               a.numero, a.torre, a.bloque, a.piso
        FROM paquete p
        LEFT JOIN apartamento a ON a.id_apartamento = p.id_apartamento
        WHERE p.id_condominio = $1 AND p.id_apartamento = $2
        ORDER BY p.fecha_recibido DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.QueryContext(ctx, query, condominioID, apartmentID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		var pkg models.Package
		if err := rows.Scan(
			&pkg.ID,
			&pkg.ApartmentID,
			&pkg.Resident,
			&pkg.Carrier,
			&pkg.Notes,
			&pkg.ReceivedAt,
			&pkg.DeliveredAt,
			&pkg.CondominioID,
			&pkg.CreatedAt,
			&pkg.UpdatedAt,
			&pkg.ApartmentNumber,
			&pkg.ApartmentTower,
			&pkg.ApartmentBlock,
			&pkg.ApartmentFloor,
		); err != nil {
			return nil, 0, err
		}
		packages = append(packages, &pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return packages, total, nil
}

func (r *PackageRepositoryImpl) MarkDelivered(ctx context.Context, packageID, condominioID string) error {
	query := `
        UPDATE paquete
        SET fecha_entregado = NOW(), fecha_actualizacion = NOW()
        WHERE id_paquete = $1 AND id_condominio = $2
    `
	result, err := r.db.ExecContext(ctx, query, packageID, condominioID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}
