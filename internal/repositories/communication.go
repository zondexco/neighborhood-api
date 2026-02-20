package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

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

// allColumns son las columnas base de comunicado
const commCols = `id_comunicado, titulo, contenido, fecha_publicacion, id_usuario_emisor, id_condominio,
	programado_para, roles_destino, icono, permite_comentarios, publicado, fecha_publicacion`

// scanCommunication escanea una fila en un Communication
func scanCommunication(row interface{ Scan(dest ...any) error }, c *models.Communication) error {
	return row.Scan(
		&c.ID,
		&c.Titulo,
		&c.Contenido,
		&c.Fecha,
		&c.Autor,
		&c.CondominioID,
		&c.ProgramadoPara,
		&c.RolesDestino,
		&c.Icono,
		&c.PermiteComentarios,
		&c.Publicado,
		&c.CreatedAt,
	)
}

// Create crea un nuevo comunicado
func (r *CommunicationRepositoryImpl) Create(ctx context.Context, communication *models.Communication) error {
	query := `
		INSERT INTO comunicado (
			id_comunicado, titulo, contenido, fecha_publicacion,
			id_usuario_emisor, id_condominio,
			programado_para, roles_destino, icono, permite_comentarios, publicado
		) VALUES (
			gen_random_uuid(), $1, $2, NOW(),
			$3, $4,
			$5, $6, $7, $8, $9
		)
		RETURNING id_comunicado, fecha_publicacion
	`

	var rolesDestino *pq.StringArray
	if len(communication.RolesDestino) > 0 {
		arr := pq.StringArray(communication.RolesDestino)
		rolesDestino = &arr
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		communication.Titulo,
		communication.Contenido,
		communication.Autor,
		communication.CondominioID,
		communication.ProgramadoPara,
		rolesDestino,
		communication.Icono,
		communication.PermiteComentarios,
		communication.Publicado,
	).Scan(&communication.ID, &communication.CreatedAt)

	return err
}

// FindByID obtiene un comunicado por ID (con validación multi-tenant)
func (r *CommunicationRepositoryImpl) FindByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error) {
	query := `
		SELECT ` + commCols + `
		FROM comunicado
		WHERE id_comunicado = $1 AND id_condominio = $2
	`

	var communication models.Communication
	err := scanCommunication(r.db.QueryRowContext(ctx, query, communicationID, condominioID), &communication)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("communication not found")
		}
		return nil, err
	}

	return &communication, nil
}

// GetByCondominio obtiene TODOS los comunicados de un condominio (para admin, sin filtro de visibilidad)
func (r *CommunicationRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error) {
	countQuery := `SELECT COUNT(*) FROM comunicado WHERE id_condominio = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT ` + commCols + `
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
		var c models.Communication
		if err := scanCommunication(rows, &c); err != nil {
			return nil, 0, err
		}
		communications = append(communications, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return communications, total, nil
}

// ListVisible obtiene comunicados visibles para un usuario según su rol y fecha actual.
// Filtra: publicado=true, (programado_para IS NULL OR programado_para <= NOW()),
// (roles_destino IS NULL OR userRole = ANY(roles_destino))
// Incluye flag de lectura y conteo de comentarios.
func (r *CommunicationRepositoryImpl) ListVisible(ctx context.Context, condominioID, userID, userRole string, page, pageSize int) ([]*models.Communication, int, error) {
	whereClause := `
		WHERE c.id_condominio = $1
		  AND c.publicado = true
		  AND (c.programado_para IS NULL OR c.programado_para <= NOW())
		  AND (c.roles_destino IS NULL OR $2 = ANY(c.roles_destino))
	`

	countQuery := `SELECT COUNT(*) FROM comunicado c ` + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID, userRole).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT c.id_comunicado, c.titulo, c.contenido, c.fecha_publicacion,
			c.id_usuario_emisor, c.id_condominio,
			c.programado_para, c.roles_destino, c.icono, c.permite_comentarios, c.publicado,
			c.fecha_publicacion,
			CASE WHEN cl.id IS NOT NULL THEN true ELSE false END AS leido,
			COALESCE(cc.cnt, 0) AS num_comentarios
		FROM comunicado c
		LEFT JOIN comunicado_lectura cl ON cl.id_comunicado = c.id_comunicado AND cl.id_usuario = $3
		LEFT JOIN (
			SELECT id_comunicado, COUNT(*) AS cnt FROM comunicado_comentario GROUP BY id_comunicado
		) cc ON cc.id_comunicado = c.id_comunicado
	` + whereClause + `
		ORDER BY c.fecha_publicacion DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.QueryContext(ctx, query, condominioID, userRole, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var communications []*models.Communication
	for rows.Next() {
		var c models.Communication
		err := rows.Scan(
			&c.ID, &c.Titulo, &c.Contenido, &c.Fecha,
			&c.Autor, &c.CondominioID,
			&c.ProgramadoPara, &c.RolesDestino, &c.Icono, &c.PermiteComentarios, &c.Publicado,
			&c.CreatedAt,
			&c.Leido,
			&c.NumComentarios,
		)
		if err != nil {
			return nil, 0, err
		}
		communications = append(communications, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return communications, total, nil
}

// Update actualiza un comunicado
func (r *CommunicationRepositoryImpl) Update(ctx context.Context, communication *models.Communication) error {
	var rolesDestino *pq.StringArray
	if len(communication.RolesDestino) > 0 {
		arr := pq.StringArray(communication.RolesDestino)
		rolesDestino = &arr
	}

	query := `
		UPDATE comunicado SET
			titulo = $1,
			contenido = $2,
			programado_para = $3,
			roles_destino = $4,
			icono = $5,
			permite_comentarios = $6,
			publicado = $7
		WHERE id_comunicado = $8 AND id_condominio = $9
		RETURNING fecha_publicacion
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		communication.Titulo,
		communication.Contenido,
		communication.ProgramadoPara,
		rolesDestino,
		communication.Icono,
		communication.PermiteComentarios,
		communication.Publicado,
		communication.ID,
		communication.CondominioID,
	).Scan(&communication.UpdatedAt)

	return err
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

// MarkRead marca un comunicado como leído (INSERT ... ON CONFLICT DO NOTHING)
func (r *CommunicationRepositoryImpl) MarkRead(ctx context.Context, communicationID, userID string) error {
	query := `
		INSERT INTO comunicado_lectura (id_comunicado, id_usuario)
		VALUES ($1, $2)
		ON CONFLICT (id_comunicado, id_usuario) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, communicationID, userID)
	return err
}

// UnreadCount retorna la cantidad de comunicados visibles no leídos para un usuario
func (r *CommunicationRepositoryImpl) UnreadCount(ctx context.Context, condominioID, userID, userRole string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM comunicado c
		WHERE c.id_condominio = $1
		  AND c.publicado = true
		  AND (c.programado_para IS NULL OR c.programado_para <= NOW())
		  AND (c.roles_destino IS NULL OR $2 = ANY(c.roles_destino))
		  AND NOT EXISTS (
			SELECT 1 FROM comunicado_lectura cl
			WHERE cl.id_comunicado = c.id_comunicado AND cl.id_usuario = $3
		  )
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, condominioID, userRole, userID).Scan(&count)
	return count, err
}

// ListComments obtiene los comentarios de un comunicado con nombre del autor
func (r *CommunicationRepositoryImpl) ListComments(ctx context.Context, communicationID string) ([]*models.ComunicadoComentario, error) {
	query := `
		SELECT cc.id_comentario, cc.id_comunicado, cc.id_usuario, cc.contenido, cc.fecha_creacion,
			u.nombres, u.apellidos
		FROM comunicado_comentario cc
		JOIN usuario u ON u.id_usuario = cc.id_usuario
		WHERE cc.id_comunicado = $1
		ORDER BY cc.fecha_creacion ASC
	`

	rows, err := r.db.QueryContext(ctx, query, communicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.ComunicadoComentario
	for rows.Next() {
		var c models.ComunicadoComentario
		if err := rows.Scan(
			&c.ID, &c.IDComunicado, &c.IDUsuario, &c.Contenido, &c.FechaCreacion,
			&c.AutorNombre, &c.AutorApellido,
		); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}

	return comments, rows.Err()
}

// CreateComment crea un comentario en un comunicado
func (r *CommunicationRepositoryImpl) CreateComment(ctx context.Context, comment *models.ComunicadoComentario) error {
	query := `
		INSERT INTO comunicado_comentario (id_comunicado, id_usuario, contenido)
		VALUES ($1, $2, $3)
		RETURNING id_comentario, fecha_creacion
	`
	return r.db.QueryRowContext(
		ctx,
		query,
		comment.IDComunicado,
		comment.IDUsuario,
		comment.Contenido,
	).Scan(&comment.ID, &comment.FechaCreacion)
}
