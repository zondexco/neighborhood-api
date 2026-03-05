package repositories

import (
	"context"
	"database/sql"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/logger"
)

// CondominioRepositoryImpl implementación del repositorio de condominios
type CondominioRepositoryImpl struct {
	db  *database.DB
	log logger.Logger
}

// NewCondominioRepository crea un nuevo repositorio de condominios
func NewCondominioRepository(db *database.DB) CondominioRepository {
	return &CondominioRepositoryImpl{
		db:  db,
		log: logger.Get(),
	}
}

// FindByID busca un condominio por ID
func (r *CondominioRepositoryImpl) FindByID(ctx context.Context, condominioID string) (*models.Condominio, error) {
	query := `
		SELECT id_condominio, nombre, direccion, ciudad, telefono, email, logo_url, nit, representante_legal, fecha_creacion, estado, permite_soporte
		FROM condominio
		WHERE id_condominio = $1
	`

	condominio := &models.Condominio{}
	err := r.db.QueryRowContext(ctx, query, condominioID).Scan(
		&condominio.ID,
		&condominio.Nombre,
		&condominio.Direccion,
		&condominio.Ciudad,
		&condominio.Telefono,
		&condominio.Email,
		&condominio.LogoUrl,
		&condominio.NIT,
		&condominio.RepresentanteLegal,
		&condominio.FechaCreacion,
		&condominio.Estado,
		&condominio.PermiteSoporte,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.WithField("condominio_id", condominioID).Info("condominio not found")
			return nil, nil
		}
		r.log.WithError(err).WithField("condominio_id", condominioID).Error("error getting condominio by id")
		return nil, err
	}

	return condominio, nil
}

// GetAll obtiene todos los condominios
func (r *CondominioRepositoryImpl) GetAll(ctx context.Context) ([]*models.Condominio, error) {
	query := `
		SELECT id_condominio, nombre, direccion, ciudad, telefono, email, logo_url, nit, representante_legal, fecha_creacion, estado, permite_soporte
		FROM condominio
		ORDER BY nombre ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.log.WithError(err).Error("error getting all condominios")
		return nil, err
	}
	defer rows.Close()

	var condominios []*models.Condominio
	for rows.Next() {
		condominio := &models.Condominio{}
		err := rows.Scan(
			&condominio.ID,
			&condominio.Nombre,
			&condominio.Direccion,
			&condominio.Ciudad,
			&condominio.Telefono,
			&condominio.Email,
			&condominio.LogoUrl,
			&condominio.NIT,
			&condominio.RepresentanteLegal,
			&condominio.FechaCreacion,
			&condominio.Estado,
			&condominio.PermiteSoporte,
		)
		if err != nil {
			r.log.WithError(err).Error("error scanning condominio")
			return nil, err
		}
		condominios = append(condominios, condominio)
	}

	if err = rows.Err(); err != nil {
		r.log.WithError(err).Error("error iterating condominios")
		return nil, err
	}

	return condominios, nil
}

// Update actualiza la información de un condominio
func (r *CondominioRepositoryImpl) Update(ctx context.Context, condominio *models.Condominio) error {
	query := `
		UPDATE condominio SET
			nombre = $1,
			direccion = $2,
			ciudad = $3,
			telefono = $4,
			email = $5,
			nit = $6,
			representante_legal = $7,
			permite_soporte = $8
		WHERE id_condominio = $9
	`

	result, err := r.db.ExecContext(ctx, query,
		condominio.Nombre,
		condominio.Direccion,
		condominio.Ciudad,
		condominio.Telefono,
		condominio.Email,
		condominio.NIT,
		condominio.RepresentanteLegal,
		condominio.PermiteSoporte,
		condominio.ID,
	)
	if err != nil {
		r.log.WithError(err).WithField("condominio_id", condominio.ID).Error("error updating condominio")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Create crea un nuevo condominio
func (r *CondominioRepositoryImpl) Create(ctx context.Context, condominio *models.Condominio) error {
	query := `
		INSERT INTO condominio (nombre, direccion, ciudad, telefono, email, nit, representante_legal, estado, permite_soporte)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id_condominio, fecha_creacion
	`

	err := r.db.QueryRowContext(ctx, query,
		condominio.Nombre,
		condominio.Direccion,
		condominio.Ciudad,
		condominio.Telefono,
		condominio.Email,
		condominio.NIT,
		condominio.RepresentanteLegal,
		condominio.Estado,
		condominio.PermiteSoporte,
	).Scan(&condominio.ID, &condominio.FechaCreacion)

	if err != nil {
		r.log.WithError(err).Error("error creating condominio")
		return err
	}

	return nil
}

// Delete elimina un condominio
func (r *CondominioRepositoryImpl) Delete(ctx context.Context, condominioID string) error {
	query := `DELETE FROM condominio WHERE id_condominio = $1`

	result, err := r.db.ExecContext(ctx, query, condominioID)
	if err != nil {
		r.log.WithError(err).WithField("condominio_id", condominioID).Error("error deleting condominio")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
