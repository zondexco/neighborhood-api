package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/logger"
)

// ApartmentRepositoryImpl implementación concreta
type ApartmentRepositoryImpl struct {
	db     *database.DB
	logger logger.Logger
}

// NewApartmentRepository crea una nueva instancia del repositorio
func NewApartmentRepository(db *database.DB, log logger.Logger) ApartmentRepository {
	return &ApartmentRepositoryImpl{
		db:     db,
		logger: log,
	}
}

// Create crea un nuevo apartamento
func (r *ApartmentRepositoryImpl) Create(ctx context.Context, apartment *models.Apartment) error {
	query := `
		INSERT INTO apartamento (id_apartamento, id_condominio, numero, bloque, torre, piso, id_propietario, area_privada, area_comun, estado)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING fecha_registro
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		apartment.ID,
		apartment.CondominioID,
		apartment.Numero,
		apartment.Bloque,
		apartment.Torre,
		apartment.Piso,
		apartment.PropietarioID,
		apartment.AreaPrivada,
		apartment.AreaComun,
		apartment.Estado,
	).Scan(&apartment.FechaRegistro)

	if err != nil {
		r.logger.WithError(err).WithField("apartment_id", apartment.ID).Error("error creating apartment")
		return fmt.Errorf("error creating apartment: %w", err)
	}

	r.logger.WithField("apartment_id", apartment.ID).Info("apartment created successfully")
	return nil
}

// FindByID busca un apartamento por ID
func (r *ApartmentRepositoryImpl) FindByID(ctx context.Context, apartmentID string) (*models.Apartment, error) {
	query := `
		SELECT id_apartamento, id_condominio, numero, bloque, torre, piso, id_propietario, area_privada, area_comun, fecha_registro, estado
		FROM apartamento
		WHERE id_apartamento = $1
	`

	apartment := &models.Apartment{}
	err := r.db.QueryRowContext(ctx, query, apartmentID).Scan(
		&apartment.ID,
		&apartment.CondominioID,
		&apartment.Numero,
		&apartment.Bloque,
		&apartment.Torre,
		&apartment.Piso,
		&apartment.PropietarioID,
		&apartment.AreaPrivada,
		&apartment.AreaComun,
		&apartment.FechaRegistro,
		&apartment.Estado,
	)

	if err == sql.ErrNoRows {
		r.logger.WithField("apartment_id", apartmentID).Info("apartment not found")
		return nil, errors.New("apartment not found")
	}

	if err != nil {
		r.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error getting apartment")
		return nil, fmt.Errorf("error getting apartment: %w", err)
	}

	return apartment, nil
}

// GetByCondominio obtiene una lista de apartamentos con paginación
func (r *ApartmentRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Apartment, int, error) {
	// Contar total
	countQuery := `SELECT COUNT(*) FROM apartamento WHERE id_condominio = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total)
	if err != nil {
		r.logger.WithError(err).Error("error counting apartments")
		return nil, 0, fmt.Errorf("error counting apartments: %w", err)
	}

	// Calcular offset
	offset := (page - 1) * pageSize

	// Listar con paginación
	listQuery := `
		SELECT id_apartamento, id_condominio, numero, bloque, torre, piso, id_propietario, area_privada, area_comun, fecha_registro, estado
		FROM apartamento
		WHERE id_condominio = $1
		ORDER BY piso, numero ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, listQuery, condominioID, pageSize, offset)
	if err != nil {
		r.logger.WithError(err).Error("error listing apartments")
		return nil, 0, fmt.Errorf("error listing apartments: %w", err)
	}
	defer rows.Close()

	apartments := make([]*models.Apartment, 0, pageSize)
	for rows.Next() {
		apartment := &models.Apartment{}
		err := rows.Scan(
			&apartment.ID,
			&apartment.CondominioID,
			&apartment.Numero,
			&apartment.Bloque,
			&apartment.Torre,
			&apartment.Piso,
			&apartment.PropietarioID,
			&apartment.AreaPrivada,
			&apartment.AreaComun,
			&apartment.FechaRegistro,
			&apartment.Estado,
		)
		if err != nil {
			r.logger.WithError(err).Error("error scanning apartment")
			return nil, 0, fmt.Errorf("error scanning apartment: %w", err)
		}
		apartments = append(apartments, apartment)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("error iterating apartments")
		return nil, 0, fmt.Errorf("error iterating apartments: %w", err)
	}

	return apartments, total, nil
}

// GetByOwner obtiene apartamentos por propietario dentro de un condominio
func (r *ApartmentRepositoryImpl) GetByOwner(ctx context.Context, condominioID, ownerID string, page, pageSize int) ([]*models.Apartment, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM apartamento a
		WHERE a.id_condominio = $1
		  AND (
				a.id_propietario = $2 OR
				a.id_apartamento IN (
					SELECT id_apartamento FROM usuario WHERE id_usuario = $2 AND id_apartamento IS NOT NULL
				)
			)
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID, ownerID).Scan(&total)
	if err != nil {
		r.logger.WithError(err).Error("error counting apartments by owner")
		return nil, 0, fmt.Errorf("error counting apartments: %w", err)
	}

	offset := (page - 1) * pageSize
	listQuery := `
		SELECT id_apartamento, id_condominio, numero, bloque, torre, piso, id_propietario, area_privada, area_comun, fecha_registro, estado
		FROM apartamento a
		WHERE a.id_condominio = $1
		  AND (
				a.id_propietario = $2 OR
				a.id_apartamento IN (
					SELECT id_apartamento FROM usuario WHERE id_usuario = $2 AND id_apartamento IS NOT NULL
				)
			)
		ORDER BY piso, numero ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, listQuery, condominioID, ownerID, pageSize, offset)
	if err != nil {
		r.logger.WithError(err).Error("error listing apartments by owner")
		return nil, 0, fmt.Errorf("error listing apartments: %w", err)
	}
	defer rows.Close()

	apartments := make([]*models.Apartment, 0, pageSize)
	for rows.Next() {
		apartment := &models.Apartment{}
		err := rows.Scan(
			&apartment.ID,
			&apartment.CondominioID,
			&apartment.Numero,
			&apartment.Bloque,
			&apartment.Torre,
			&apartment.Piso,
			&apartment.PropietarioID,
			&apartment.AreaPrivada,
			&apartment.AreaComun,
			&apartment.FechaRegistro,
			&apartment.Estado,
		)
		if err != nil {
			r.logger.WithError(err).Error("error scanning apartment by owner")
			return nil, 0, fmt.Errorf("error scanning apartment: %w", err)
		}
		apartments = append(apartments, apartment)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("error iterating apartments by owner")
		return nil, 0, fmt.Errorf("error iterating apartments: %w", err)
	}

	return apartments, total, nil
}

// Update actualiza un apartamento
func (r *ApartmentRepositoryImpl) Update(ctx context.Context, apartment *models.Apartment) error {
	query := `
		UPDATE apartamento
		SET numero = $1, bloque = $2, torre = $3, piso = $4, id_propietario = $5, area_privada = $6, area_comun = $7, estado = $8
		WHERE id_apartamento = $9
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		apartment.Numero,
		apartment.Bloque,
		apartment.Torre,
		apartment.Piso,
		apartment.PropietarioID,
		apartment.AreaPrivada,
		apartment.AreaComun,
		apartment.Estado,
		apartment.ID,
	)

	if err != nil {
		r.logger.WithError(err).WithField("apartment_id", apartment.ID).Error("error updating apartment")
		return fmt.Errorf("error updating apartment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("error getting rows affected")
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.WithField("apartment_id", apartment.ID).Info("apartment not found for update")
		return errors.New("apartment not found")
	}

	r.logger.WithField("apartment_id", apartment.ID).Info("apartment updated successfully")
	return nil
}

// Delete elimina un apartamento
func (r *ApartmentRepositoryImpl) Delete(ctx context.Context, apartmentID string) error {
	query := `DELETE FROM apartamento WHERE id_apartamento = $1`

	result, err := r.db.ExecContext(ctx, query, apartmentID)
	if err != nil {
		r.logger.WithError(err).WithField("apartment_id", apartmentID).Error("error deleting apartment")
		return fmt.Errorf("error deleting apartment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("error getting rows affected")
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.WithField("apartment_id", apartmentID).Info("apartment not found for delete")
		return errors.New("apartment not found")
	}

	r.logger.WithField("apartment_id", apartmentID).Info("apartment deleted successfully")
	return nil
}

// GetByPropietario obtiene apartamentos de un propietario
func (r *ApartmentRepositoryImpl) GetByPropietario(ctx context.Context, propietarioID string) ([]*models.Apartment, error) {
	query := `
		SELECT id_apartamento, id_condominio, numero, bloque, torre, piso, id_propietario, area_privada, area_comun, fecha_registro, estado
		FROM apartamento
		WHERE id_propietario = $1
		ORDER BY piso, numero ASC
	`

	rows, err := r.db.QueryContext(ctx, query, propietarioID)
	if err != nil {
		r.logger.WithError(err).Error("error getting apartments by propietario")
		return nil, fmt.Errorf("error getting apartments by propietario: %w", err)
	}
	defer rows.Close()

	apartments := make([]*models.Apartment, 0)
	for rows.Next() {
		apartment := &models.Apartment{}
		err := rows.Scan(
			&apartment.ID,
			&apartment.CondominioID,
			&apartment.Numero,
			&apartment.Bloque,
			&apartment.Torre,
			&apartment.Piso,
			&apartment.PropietarioID,
			&apartment.AreaPrivada,
			&apartment.AreaComun,
			&apartment.FechaRegistro,
			&apartment.Estado,
		)
		if err != nil {
			r.logger.WithError(err).Error("error scanning apartment")
			return nil, fmt.Errorf("error scanning apartment: %w", err)
		}
		apartments = append(apartments, apartment)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("error iterating apartments")
		return nil, fmt.Errorf("error iterating apartments: %w", err)
	}

	return apartments, nil
}
