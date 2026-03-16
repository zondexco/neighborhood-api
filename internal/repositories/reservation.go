package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
)

// ReservationRepositoryImpl implementa ReservationRepository
type ReservationRepositoryImpl struct {
	db *database.DB
}

// NewReservationRepository crea una nueva instancia del repositorio
func NewReservationRepository(db *database.DB) ReservationRepository {
	return &ReservationRepositoryImpl{db: db}
}

// GetEspacioCostoHora obtiene el costo_hora del espacio común
func (r *ReservationRepositoryImpl) GetEspacioCostoHora(ctx context.Context, espacioID, condominioID string) (*float64, error) {
	query := `SELECT costo_hora FROM espacio_comun WHERE id_espacio = $1 AND id_condominio = $2`
	var costo *float64

	if err := r.db.QueryRowContext(ctx, query, espacioID, condominioID).Scan(&costo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("espacio not found")
		}
		return nil, err
	}

	return costo, nil
}

// GetEspacioNombre obtiene el nombre del espacio común
func (r *ReservationRepositoryImpl) GetEspacioNombre(ctx context.Context, espacioID, condominioID string) (string, error) {
	query := `SELECT nombre FROM espacio_comun WHERE id_espacio = $1 AND id_condominio = $2`
	var nombre string

	if err := r.db.QueryRowContext(ctx, query, espacioID, condominioID).Scan(&nombre); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("espacio not found")
		}
		return "", err
	}

	return nombre, nil
}

// Create crea una nueva reserva
func (r *ReservationRepositoryImpl) Create(ctx context.Context, reservation *models.Reservation) error {
	if reservation.ID == "" {
		reservation.ID = uuid.New().String()
	}

	query := `
		INSERT INTO reserva (
			id_reserva, id_condominio, id_usuario, id_apartamento, id_espacio,
			fecha_inicio, fecha_fin, personas_esperadas, observaciones, valor_base, costo_total, pagado, estado
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING fecha_solicitud
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		reservation.ID,
		reservation.CondominioID,
		reservation.UsuarioID,
		reservation.ApartmentID,
		reservation.EspacioID,
		reservation.FechaInicio,
		reservation.FechaFin,
		reservation.PersonasEsperadas,
		reservation.Observaciones,
		reservation.ValorBase,
		reservation.CostoTotal,
		reservation.Pagado,
		reservation.Estado,
	).Scan(&reservation.FechaSolicitud)

	if err != nil {
		return err
	}

	return nil
}

// FindByID obtiene una reserva por ID (con validación multi-tenant)
func (r *ReservationRepositoryImpl) FindByID(ctx context.Context, reservationID, condominioID string) (*models.Reservation, error) {
	query := `
		SELECT id_reserva, id_condominio, id_usuario, id_apartamento, id_espacio,
		       fecha_solicitud, fecha_inicio, fecha_fin, personas_esperadas, observaciones,
		       valor_base, costo_total, pagado, estado
		FROM reserva
		WHERE id_reserva = $1 AND id_condominio = $2
	`

	var reservation models.Reservation
	err := r.db.QueryRowContext(ctx, query, reservationID, condominioID).Scan(
		&reservation.ID,
		&reservation.CondominioID,
		&reservation.UsuarioID,
		&reservation.ApartmentID,
		&reservation.EspacioID,
		&reservation.FechaSolicitud,
		&reservation.FechaInicio,
		&reservation.FechaFin,
		&reservation.PersonasEsperadas,
		&reservation.Observaciones,
		&reservation.ValorBase,
		&reservation.CostoTotal,
		&reservation.Pagado,
		&reservation.Estado,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reservation not found")
		}
		return nil, err
	}

	return &reservation, nil
}

// GetByCondominio obtiene reservas de un condominio con paginación
func (r *ReservationRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	// Obtener total
	countQuery := `SELECT COUNT(*) FROM reserva WHERE id_condominio = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Obtener registros
	offset := (page - 1) * pageSize
	query := `
		SELECT id_reserva, id_condominio, id_usuario, id_apartamento, id_espacio,
		       fecha_solicitud, fecha_inicio, fecha_fin, personas_esperadas, observaciones,
		       valor_base, costo_total, pagado, estado
		FROM reserva
		WHERE id_condominio = $1
		ORDER BY fecha_inicio DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reservations []*models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.CondominioID,
			&reservation.UsuarioID,
			&reservation.ApartmentID,
			&reservation.EspacioID,
			&reservation.FechaSolicitud,
			&reservation.FechaInicio,
			&reservation.FechaFin,
			&reservation.PersonasEsperadas,
			&reservation.Observaciones,
			&reservation.ValorBase,
			&reservation.CostoTotal,
			&reservation.Pagado,
			&reservation.Estado,
		)
		if err != nil {
			return nil, 0, err
		}
		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return reservations, total, nil
}

// GetByUsuario obtiene reservas de un usuario
func (r *ReservationRepositoryImpl) GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	// Obtener total
	countQuery := `SELECT COUNT(*) FROM reserva WHERE id_usuario = $1 AND id_condominio = $2`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, usuarioID, condominioID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Obtener registros
	offset := (page - 1) * pageSize
	query := `
		SELECT id_reserva, id_condominio, id_usuario, id_apartamento, id_espacio,
		       fecha_solicitud, fecha_inicio, fecha_fin, personas_esperadas, observaciones,
		       valor_base, costo_total, pagado, estado
		FROM reserva
		WHERE id_usuario = $1 AND id_condominio = $2
		ORDER BY fecha_inicio DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, usuarioID, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reservations []*models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.CondominioID,
			&reservation.UsuarioID,
			&reservation.ApartmentID,
			&reservation.EspacioID,
			&reservation.FechaSolicitud,
			&reservation.FechaInicio,
			&reservation.FechaFin,
			&reservation.PersonasEsperadas,
			&reservation.Observaciones,
			&reservation.ValorBase,
			&reservation.CostoTotal,
			&reservation.Pagado,
			&reservation.Estado,
		)
		if err != nil {
			return nil, 0, err
		}
		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return reservations, total, nil
}

// GetByApartment obtiene reservas de todos los usuarios de un apartamento
func (r *ReservationRepositoryImpl) GetByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	countQuery := `SELECT COUNT(*) FROM reserva WHERE id_apartamento = $1 AND id_condominio = $2`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, apartmentID, condominioID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id_reserva, id_condominio, id_usuario, id_apartamento, id_espacio,
		       fecha_solicitud, fecha_inicio, fecha_fin, personas_esperadas, observaciones,
		       valor_base, costo_total, pagado, estado
		FROM reserva
		WHERE id_apartamento = $1 AND id_condominio = $2
		ORDER BY fecha_inicio DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, apartmentID, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reservations []*models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		if err := rows.Scan(
			&reservation.ID,
			&reservation.CondominioID,
			&reservation.UsuarioID,
			&reservation.ApartmentID,
			&reservation.EspacioID,
			&reservation.FechaSolicitud,
			&reservation.FechaInicio,
			&reservation.FechaFin,
			&reservation.PersonasEsperadas,
			&reservation.Observaciones,
			&reservation.ValorBase,
			&reservation.CostoTotal,
			&reservation.Pagado,
			&reservation.Estado,
		); err != nil {
			return nil, 0, err
		}
		reservations = append(reservations, &reservation)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return reservations, total, nil
}

// Update actualiza una reserva
func (r *ReservationRepositoryImpl) Update(ctx context.Context, reservation *models.Reservation) error {
	query := `
		UPDATE reserva SET
			id_apartamento = $1,
			fecha_inicio = $2,
			fecha_fin = $3,
			personas_esperadas = $4,
			observaciones = $5,
			valor_base = $6,
			costo_total = $7,
			pagado = $8,
			estado = $9
		WHERE id_reserva = $10 AND id_condominio = $11
		RETURNING fecha_solicitud
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		reservation.ApartmentID,
		reservation.FechaInicio,
		reservation.FechaFin,
		reservation.PersonasEsperadas,
		reservation.Observaciones,
		reservation.ValorBase,
		reservation.CostoTotal,
		reservation.Pagado,
		reservation.Estado,
		reservation.ID,
		reservation.CondominioID,
	).Scan(&reservation.FechaSolicitud)

	if err != nil {
		return err
	}

	return nil
}

// Delete elimina una reserva (multi-tenant)
func (r *ReservationRepositoryImpl) Delete(ctx context.Context, reservationID, condominioID string) error {
	query := `DELETE FROM reserva WHERE id_reserva = $1 AND id_condominio = $2`

	result, err := r.db.ExecContext(ctx, query, reservationID, condominioID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found")
	}

	return nil
}

// ExistsOverlap valida solapamiento de reservas activas para un espacio.
// Usa EXISTS para cortocircuitar en la primera fila encontrada.
func (r *ReservationRepositoryImpl) ExistsOverlap(ctx context.Context, espacioID, condominioID string, start, end time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM reserva
			WHERE id_espacio = $1
				AND id_condominio = $2
				AND estado NOT IN ('cancelado', 'rechazada')
				AND fecha_inicio < $4
				AND fecha_fin > $3
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, espacioID, condominioID, start, end).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// GetByEspacio obtiene reservas de un espacio común
func (r *ReservationRepositoryImpl) GetByEspacio(ctx context.Context, espacioID, condominioID string, page, pageSize int) ([]*models.Reservation, int, error) {
	// Obtener total
	countQuery := `SELECT COUNT(*) FROM reserva WHERE id_espacio = $1 AND id_condominio = $2`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, espacioID, condominioID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Obtener registros
	offset := (page - 1) * pageSize
	query := `
		SELECT id_reserva, id_condominio, id_usuario, id_apartamento, id_espacio,
		       fecha_solicitud, fecha_inicio, fecha_fin, personas_esperadas, observaciones,
		       valor_base, costo_total, pagado, estado
		FROM reserva
		WHERE id_espacio = $1 AND id_condominio = $2
		ORDER BY fecha_inicio DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, espacioID, condominioID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reservations []*models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.CondominioID,
			&reservation.UsuarioID,
			&reservation.ApartmentID,
			&reservation.EspacioID,
			&reservation.FechaSolicitud,
			&reservation.FechaInicio,
			&reservation.FechaFin,
			&reservation.PersonasEsperadas,
			&reservation.Observaciones,
			&reservation.ValorBase,
			&reservation.CostoTotal,
			&reservation.Pagado,
			&reservation.Estado,
		)
		if err != nil {
			return nil, 0, err
		}
		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return reservations, total, nil
}
