package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/logger"
)

type NotificationRepositoryImpl struct {
	db  *database.DB
	log logger.Logger
}

func NewNotificationRepository(db *database.DB) NotificationRepository {
	return &NotificationRepositoryImpl{db: db, log: logger.Get()}
}

func (r *NotificationRepositoryImpl) Create(ctx context.Context, notif *models.Notification) error {
	query := `
		INSERT INTO notificacion (id_usuario, id_condominio, tipo, titulo, mensaje, referencia_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, fecha_creacion
	`
	return r.db.QueryRowContext(ctx, query,
		notif.UserID, notif.CondominioID, notif.Tipo,
		notif.Titulo, notif.Mensaje, notif.ReferenciaID,
	).Scan(&notif.ID, &notif.FechaCreacion)
}

// CreateBatch inserta múltiples notificaciones en un solo round-trip.
func (r *NotificationRepositoryImpl) CreateBatch(ctx context.Context, notifs []*models.Notification) error {
	if len(notifs) == 0 {
		return nil
	}

	// Build multi-row INSERT: INSERT INTO notificacion (...) VALUES ($1,$2,...), ($7,$8,...)
	valueStrings := make([]string, 0, len(notifs))
	args := make([]interface{}, 0, len(notifs)*6)
	for i, n := range notifs {
		base := i * 6
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6))
		args = append(args, n.UserID, n.CondominioID, n.Tipo, n.Titulo, n.Mensaje, n.ReferenciaID)
	}

	query := `INSERT INTO notificacion (id_usuario, id_condominio, tipo, titulo, mensaje, referencia_id) VALUES ` +
		strings.Join(valueStrings, ", ")

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *NotificationRepositoryImpl) ListByUser(ctx context.Context, userID, condominioID string, page, pageSize int) ([]*models.Notification, int, error) {
	offset := (page - 1) * pageSize

	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notificacion WHERE id_usuario = $1 AND id_condominio = $2`,
		userID, condominioID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, id_usuario, id_condominio, tipo, titulo, mensaje, leido, referencia_id, fecha_creacion
		FROM notificacion
		WHERE id_usuario = $1 AND id_condominio = $2
		ORDER BY fecha_creacion DESC
		LIMIT $3 OFFSET $4
	`, userID, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]*models.Notification, 0, pageSize)
	for rows.Next() {
		n := &models.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.CondominioID, &n.Tipo, &n.Titulo, &n.Mensaje, &n.Leido, &n.ReferenciaID, &n.FechaCreacion); err != nil {
			return nil, 0, err
		}
		result = append(result, n)
	}
	return result, total, rows.Err()
}

func (r *NotificationRepositoryImpl) MarkRead(ctx context.Context, notifID, userID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notificacion SET leido = true WHERE id = $1 AND id_usuario = $2`,
		notifID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *NotificationRepositoryImpl) MarkAllRead(ctx context.Context, userID, condominioID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notificacion SET leido = true WHERE id_usuario = $1 AND id_condominio = $2 AND leido = false`,
		userID, condominioID,
	)
	return err
}

func (r *NotificationRepositoryImpl) UnreadCount(ctx context.Context, userID, condominioID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notificacion WHERE id_usuario = $1 AND id_condominio = $2 AND leido = false`,
		userID, condominioID,
	).Scan(&count)
	return count, err
}

func (r *NotificationRepositoryImpl) Delete(ctx context.Context, notifID, userID string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM notificacion WHERE id = $1 AND id_usuario = $2`,
		notifID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
