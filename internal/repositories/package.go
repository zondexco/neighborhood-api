package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/dto"
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

// ListWithFilters lista paquetes con filtros dinámicos server-side.
// Ordena: pendientes primero (fecha_entregado IS NULL), luego por fecha_recibido DESC.
func (r *PackageRepositoryImpl) ListWithFilters(ctx context.Context, condominioID string, f dto.PackageFilter) ([]*models.Package, int, error) {
	conditions := []string{"p.id_condominio = $1"}
	args := []interface{}{condominioID}
	idx := 2

	if f.ApartmentID != "" {
		conditions = append(conditions, fmt.Sprintf("p.id_apartamento = $%d", idx))
		args = append(args, f.ApartmentID)
		idx++
	}
	switch f.Status {
	case "pending":
		conditions = append(conditions, "p.fecha_entregado IS NULL")
	case "delivered":
		conditions = append(conditions, "p.fecha_entregado IS NOT NULL")
	}
	if f.Carrier != "" {
		conditions = append(conditions, fmt.Sprintf("p.transportadora ILIKE $%d", idx))
		args = append(args, f.Carrier)
		idx++
	}
	if f.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.residente ILIKE $%d OR p.transportadora ILIKE $%d)", idx, idx+1))
		term := "%" + f.Search + "%"
		args = append(args, term, term)
		idx += 2
	}
	if !f.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("p.fecha_recibido >= $%d", idx))
		args = append(args, f.DateFrom)
		idx++
	}
	if !f.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("p.fecha_recibido <= $%d", idx))
		args = append(args, f.DateTo)
		idx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM paquete p %s", where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize

	query := fmt.Sprintf(`
        SELECT p.id_paquete, p.id_apartamento, p.residente, p.transportadora, p.notas,
               p.fecha_recibido, p.fecha_entregado, p.id_condominio, p.fecha_creacion, p.fecha_actualizacion,
               a.numero, a.torre, a.bloque, a.piso
        FROM paquete p
        LEFT JOIN apartamento a ON a.id_apartamento = p.id_apartamento
        %s
        ORDER BY (p.fecha_entregado IS NOT NULL), p.fecha_recibido DESC
        LIMIT $%d OFFSET $%d
    `, where, idx, idx+1)

	args = append(args, f.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	pkgs, _, err := scanPackageRows(rows)
	return pkgs, total, err
}

// GetByCondominio mantiene compatibilidad con código existente.
func (r *PackageRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Package, int, error) {
	return r.ListWithFilters(ctx, condominioID, dto.PackageFilter{Page: page, PageSize: pageSize})
}

func (r *PackageRepositoryImpl) GetByApartment(ctx context.Context, condominioID, apartmentID string, page, pageSize int) ([]*models.Package, int, error) {
	return r.ListWithFilters(ctx, condominioID, dto.PackageFilter{
		ApartmentID: apartmentID,
		Page:        page,
		PageSize:    pageSize,
	})
}

// Update actualiza campos permitidos de un paquete.
func (r *PackageRepositoryImpl) Update(ctx context.Context, packageID, condominioID string, fields map[string]interface{}) (*models.Package, error) {
	if len(fields) == 0 {
		return r.FindByID(ctx, packageID, condominioID)
	}

	columnMap := map[string]string{
		"notes":        "notas",
		"resident":     "residente",
		"carrier":      "transportadora",
		"apartment_id": "id_apartamento",
	}

	setClauses := []string{}
	args := []interface{}{}
	idx := 1
	for field, value := range fields {
		col, ok := columnMap[field]
		if !ok {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, value)
		idx++
	}
	if len(setClauses) == 0 {
		return r.FindByID(ctx, packageID, condominioID)
	}

	setClauses = append(setClauses, "fecha_actualizacion = NOW()")
	args = append(args, packageID, condominioID)

	query := fmt.Sprintf(
		"UPDATE paquete SET %s WHERE id_paquete = $%d AND id_condominio = $%d",
		strings.Join(setClauses, ", "), idx, idx+1,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	rowsAff, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAff == 0 {
		return nil, errors.New("package not found")
	}

	return r.FindByID(ctx, packageID, condominioID)
}

// Delete elimina un paquete.
func (r *PackageRepositoryImpl) Delete(ctx context.Context, packageID, condominioID string) error {
	query := `DELETE FROM paquete WHERE id_paquete = $1 AND id_condominio = $2`
	result, err := r.db.ExecContext(ctx, query, packageID, condominioID)
	if err != nil {
		return err
	}
	rowsAff, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return errors.New("package not found")
	}
	return nil
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
	rowsAff, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return errors.New("package not found")
	}
	return nil
}

func scanPackageRows(rows *sql.Rows) ([]*models.Package, int, error) {
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
	return packages, len(packages), nil
}
