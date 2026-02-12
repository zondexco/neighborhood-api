package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
)

// CommunicationRepositoryImpl implementa CommunicationRepository
type CommunicationRepositoryImpl struct {
	db *database.DB
}

// NewCommunicationRepository crea una nueva instancia del repositorio
func NewCommunicationRepository(db *database.DB) CommunicationRepository {
	return &CommunicationRepositoryImpl{db: db}
}

// Create crea un nuevo comunicado
func (r *CommunicationRepositoryImpl) Create(ctx context.Context, communication *models.Communication) error {
	if communication.ID == "" {
		communication.ID = uuid.New().String()
	}

	query := `
		INSERT INTO comunicado (
			id_comunicado, titulo, contenido, fecha_publicacion, id_usuario_emisor, id_condominio
		) VALUES (
			gen_random_uuid(), $1, $2, NOW(), $3, $4
		)
		RETURNING fecha_publicacion, fecha_publicacion
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		communication.Titulo,
		communication.Contenido,
		communication.Autor,
		communication.CondominioID,
	).Scan(&communication.CreatedAt, &communication.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// FindByID obtiene un comunicado por ID (con validación multi-tenant)
func (r *CommunicationRepositoryImpl) FindByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error) {
	query := `
		SELECT id_comunicado, titulo, contenido, fecha_publicacion, id_usuario_emisor, id_condominio, fecha_publicacion, fecha_publicacion
		FROM comunicado
		WHERE id_comunicado = $1 AND id_condominio = $2
	`

	var communication models.Communication
	err := r.db.QueryRowContext(ctx, query, communicationID, condominioID).Scan(
		&communication.ID,
		&communication.Titulo,
		&communication.Contenido,
		&communication.Fecha,
		&communication.Autor,
		&communication.CondominioID,
		&communication.CreatedAt,
		&communication.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("communication not found")
		}
		return nil, err
	}

	return &communication, nil
}

// GetByCondominio obtiene comunicados de un condominio con paginación
func (r *CommunicationRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error) {
	// Obtener total
	countQuery := `SELECT COUNT(*) FROM comunicado WHERE id_condominio = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Obtener registros
	offset := (page - 1) * pageSize
	query := `
		SELECT id_comunicado, titulo, contenido, fecha_publicacion, id_usuario_emisor, id_condominio, fecha_publicacion, fecha_publicacion
		FROM comunicado
		WHERE id_condominio = $1
		ORDER BY fecha_publicacion DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var communications []*models.Communication
	for rows.Next() {
		var communication models.Communication
		err := rows.Scan(
			&communication.ID,
			&communication.Titulo,
			&communication.Contenido,
			&communication.Fecha,
			&communication.Autor,
			&communication.CondominioID,
			&communication.CreatedAt,
			&communication.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		communications = append(communications, &communication)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return communications, total, nil
}

// Update actualiza un comunicado
func (r *CommunicationRepositoryImpl) Update(ctx context.Context, communication *models.Communication) error {
	query := `
		UPDATE comunicado SET
			titulo = $1,
			contenido = $2,
			fecha_publicacion = $3,
			fecha_publicacion = NOW()
		WHERE id_comunicado = $4 AND id_condominio = $5
		RETURNING fecha_publicacion
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		communication.Titulo,
		communication.Contenido,
		communication.Fecha,
		communication.ID,
		communication.CondominioID,
	).Scan(&communication.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// Delete elimina un comunicado (multi-tenant)
func (r *CommunicationRepositoryImpl) Delete(ctx context.Context, communicationID, condominioID string) error {
	query := `DELETE FROM comunicado WHERE id_comunicado = $1 AND id_condominio = $2`

	result, err := r.db.ExecContext(ctx, query, communicationID, condominioID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("communication not found")
	}

	return nil
}
