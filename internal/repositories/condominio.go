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
		SELECT id_condominio, nombre, direccion, ciudad, telefono, email, logo_url, nit, representante_legal, fecha_creacion, estado
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
		SELECT id_condominio, nombre, direccion, ciudad, telefono, email, logo_url, nit, representante_legal, fecha_creacion, estado
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
