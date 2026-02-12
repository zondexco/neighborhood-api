package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
)

// SpaceRepositoryImpl implementa SpaceRepository
type SpaceRepositoryImpl struct {
	db *database.DB
}

// NewSpaceRepository crea una nueva instancia del repositorio
func NewSpaceRepository(db *database.DB) SpaceRepository {
	return &SpaceRepositoryImpl{db: db}
}

// FindByID busca un espacio por ID dentro del condominio
func (r *SpaceRepositoryImpl) FindByID(ctx context.Context, espacioID, condominioID string) (*models.Space, error) {
	query := `SELECT id_espacio, id_condominio, nombre, descripcion, costo_hora, estado,
	                 NOW() AS fecha_registro,
	                 NOW() AS fecha_actualizacion
	          FROM espacio_comun
	          WHERE id_espacio = $1 AND id_condominio = $2`

	var space models.Space
	err := r.db.QueryRowContext(ctx, query, espacioID, condominioID).Scan(
		&space.ID,
		&space.CondominioID,
		&space.Nombre,
		&space.Descripcion,
		&space.CostoHora,
		&space.Estado,
		&space.CreatedAt,
		&space.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("space not found")
		}
		return nil, err
	}

	return &space, nil
}

// GetByCondominio retorna espacios del condominio con paginacion y busqueda
func (r *SpaceRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int, search string) ([]*models.Space, int, error) {
	trimmed := strings.TrimSpace(search)

	// Total
	countWhere := "id_condominio = $1"
	countArgs := []interface{}{condominioID}
	idx := 2
	if trimmed != "" {
		countWhere += " AND LOWER(nombre) LIKE $" + strconv.Itoa(idx)
		countArgs = append(countArgs, "%"+strings.ToLower(trimmed)+"%")
		idx++
	}

	countQuery := "SELECT COUNT(*) FROM espacio_comun WHERE " + countWhere

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data
	offset := (page - 1) * pageSize
	dataWhere := "id_condominio = $1"
	dataArgs := []interface{}{condominioID}
	dataIdx := 2
	if trimmed != "" {
		dataWhere += " AND LOWER(nombre) LIKE $" + strconv.Itoa(dataIdx)
		dataArgs = append(dataArgs, "%"+strings.ToLower(trimmed)+"%")
		dataIdx++
	}

	limitPlaceholder := "$" + strconv.Itoa(dataIdx)
	offsetPlaceholder := "$" + strconv.Itoa(dataIdx+1)

	query := `SELECT id_espacio, id_condominio, nombre, descripcion, costo_hora, estado,
		      NOW() AS fecha_registro,
		      NOW() AS fecha_actualizacion
		      FROM espacio_comun
		      WHERE ` + dataWhere + ` ORDER BY nombre ASC LIMIT ` + limitPlaceholder + ` OFFSET ` + offsetPlaceholder

	dataArgs = append(dataArgs, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var spaces []*models.Space
	for rows.Next() {
		var space models.Space
		if err := rows.Scan(
			&space.ID,
			&space.CondominioID,
			&space.Nombre,
			&space.Descripcion,
			&space.CostoHora,
			&space.Estado,
			&space.CreatedAt,
			&space.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		spaces = append(spaces, &space)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return spaces, total, nil
}

// Create inserta un nuevo espacio comun
func (r *SpaceRepositoryImpl) Create(ctx context.Context, space *models.Space) error {
	if space.ID == "" {
		space.ID = uuid.New().String()
	}

	query := `INSERT INTO espacio_comun (id_espacio, id_condominio, nombre, descripcion, costo_hora, estado)
              VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		space.ID,
		space.CondominioID,
		space.Nombre,
		space.Descripcion,
		space.CostoHora,
		space.Estado,
	)

	return err
}

// Update actualiza un espacio comun
func (r *SpaceRepositoryImpl) Update(ctx context.Context, space *models.Space) error {
	query := `UPDATE espacio_comun
              SET nombre = $1, descripcion = $2, costo_hora = $3, estado = $4, fecha_actualizacion = NOW()
              WHERE id_espacio = $5 AND id_condominio = $6`

	res, err := r.db.ExecContext(
		ctx,
		query,
		space.Nombre,
		space.Descripcion,
		space.CostoHora,
		space.Estado,
		space.ID,
		space.CondominioID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("space not found")
	}

	return nil
}

// Delete elimina un espacio comun
func (r *SpaceRepositoryImpl) Delete(ctx context.Context, espacioID, condominioID string) error {
	query := `DELETE FROM espacio_comun WHERE id_espacio = $1 AND id_condominio = $2`
	res, err := r.db.ExecContext(ctx, query, espacioID, condominioID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("space not found")
	}

	return nil
}
