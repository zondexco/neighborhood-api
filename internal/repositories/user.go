package repositories

import (
	"context"
	"database/sql"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"

	"github.com/georgysavva/scany/v2/sqlscan"
)

// UserRepositoryImpl implementación de UserRepository
type UserRepositoryImpl struct {
	db  *database.DB
	log logger.Logger
}

// NewUserRepository crea una nueva instancia de UserRepository
func NewUserRepository(db *database.DB) UserRepository {
	return &UserRepositoryImpl{
		db:  db,
		log: logger.Get(),
	}
}

// FindByEmail busca un usuario por email
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string, condominioID string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id_usuario, email, nombres, apellidos, telefono, celular, id_condominio, id_apartamento, password, rol, estado, fecha_creacion, ultimo_acceso, tipo_documento, numero_documento
		FROM usuario
		WHERE email = $1 AND id_condominio = $2
	`

	err := r.db.QueryRowContext(ctx, query, email, condominioID).Scan(
		&user.ID, &user.Email, &user.Nombre, &user.Apellido, &user.Telefono,
		&user.Celular, &user.CondominioID, &user.ApartmentID, &user.PIN, &user.Rol, &user.Estado, &user.CreatedAt, &user.UpdatedAt,
		&user.TipoDoc, &user.NumeroDoc,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.WithField("email", email).Debug("User not found")
			return nil, errors.NotFoundErrorf("user not found")
		}
		r.log.WithError(err).WithField("email", email).Error("Error querying user by email")
		return nil, errors.DatabaseErrorf("error querying user").WithError(err)
	}

	return user, nil
}

// FindByEmailOnly busca un usuario por email sin requerir condominio_id (para login)
func (r *UserRepositoryImpl) FindByEmailOnly(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id_usuario, email, nombres, apellidos, telefono, celular, id_condominio, id_apartamento, password, rol, estado, fecha_creacion, ultimo_acceso, tipo_documento, numero_documento
		FROM usuario
		WHERE email = $1
	`

	if err := sqlscan.Get(ctx, r.db, &user, query, email); err != nil {
		if sqlscan.NotFound(err) {
			r.log.WithField("email", email).Debug("User not found")
			return nil, errors.NotFoundErrorf("user not found")
		}
		r.log.WithError(err).WithField("email", email).Error("Error querying user by email")
		return nil, errors.DatabaseErrorf("error querying user").WithError(err)
	}

	return &user, nil
}

// FindByID busca un usuario por ID
func (r *UserRepositoryImpl) FindByID(ctx context.Context, userID string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id_usuario, email, nombres, apellidos, telefono, celular, id_condominio, id_apartamento, password, rol, estado, fecha_creacion, ultimo_acceso, tipo_documento, numero_documento
		FROM usuario
		WHERE id_usuario = $1
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Nombre, &user.Apellido, &user.Telefono,
		&user.Celular, &user.CondominioID, &user.ApartmentID, &user.PIN, &user.Rol, &user.Estado, &user.CreatedAt, &user.UpdatedAt,
		&user.TipoDoc, &user.NumeroDoc,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.WithField("user_id", userID).Debug("User not found")
			return nil, errors.NotFoundErrorf("user not found")
		}
		r.log.WithError(err).WithField("user_id", userID).Error("Error querying user by ID")
		return nil, errors.DatabaseErrorf("error querying user").WithError(err)
	}

	return user, nil
}

// Create crea un nuevo usuario
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO usuario (
			email, apellidos, nombres, telefono, celular, id_condominio, id_apartamento,
			password, rol, estado, tipo_documento, numero_documento
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id_usuario, fecha_creacion, ultimo_acceso
	`

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Apellido,
		user.Nombre,
		user.Telefono,
		user.Celular,
		user.CondominioID,
		user.ApartmentID,
		user.PIN,
		user.Rol,
		user.Estado,
		user.TipoDoc,
		user.NumeroDoc,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.log.WithError(err).WithField("email", user.Email).Error("Error creating user")
		return errors.DatabaseErrorf("error creating user").WithError(err)
	}

	r.log.WithField("user_id", user.ID).WithField("email", user.Email).Info("User created successfully")
	return nil
}

// Update actualiza un usuario
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE usuario
		SET email = $1, apellidos = $2, nombres = $3, telefono = $4, celular = $5,
			id_apartamento = $6, rol = $7, estado = $8, tipo_documento = $9, numero_documento = $10
		WHERE id_usuario = $11
		RETURNING fecha_creacion, ultimo_acceso
	`

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Apellido,
		user.Nombre,
		user.Telefono,
		user.Celular,
		user.ApartmentID,
		user.Rol,
		user.Estado,
		user.TipoDoc,
		user.NumeroDoc,
		user.ID,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundErrorf("user not found")
		}
		r.log.WithError(err).WithField("user_id", user.ID).Error("Error updating user")
		return errors.DatabaseErrorf("error updating user").WithError(err)
	}

	r.log.WithField("user_id", user.ID).Info("User updated successfully")
	return nil
}

// UpdatePIN actualiza el PIN de un usuario
func (r *UserRepositoryImpl) UpdatePIN(ctx context.Context, userID, newPINHash string) error {
	query := `
		UPDATE usuario
		SET password = $1
		WHERE id_usuario = $2
	`

	result, err := r.db.ExecContext(ctx, query, newPINHash, userID)
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("Error updating PIN")
		return errors.DatabaseErrorf("error updating PIN").WithError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("Error getting rows affected")
		return errors.DatabaseErrorf("error updating PIN").WithError(err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundErrorf("user not found")
	}

	r.log.WithField("user_id", userID).Info("PIN updated successfully")
	return nil
}

// Delete elimina un usuario
func (r *UserRepositoryImpl) Delete(ctx context.Context, userID string) error {
	query := "DELETE FROM usuario WHERE id_usuario = $1"

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("Error deleting user")
		return errors.DatabaseErrorf("error deleting user").WithError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("Error getting rows affected")
		return errors.DatabaseErrorf("error deleting user").WithError(err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundErrorf("user not found")
	}

	r.log.WithField("user_id", userID).Info("User deleted successfully")
	return nil
}

// GetByCondominio obtiene todos los usuarios de un condominio
func (r *UserRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.User, int, error) {
	offset := (page - 1) * pageSize

	// Obtener total
	var total int
	countQuery := "SELECT COUNT(*) FROM usuario WHERE id_condominio = $1"
	err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total)
	if err != nil {
		r.log.WithError(err).Error("Error counting users")
		return nil, 0, errors.DatabaseErrorf("error counting users").WithError(err)
	}

	// Obtener registros (excluye password de listados)
	query := `
		SELECT id_usuario, email, nombres, apellidos, telefono, celular, id_condominio, id_apartamento,
			'' AS password, rol, estado, fecha_creacion, ultimo_acceso, tipo_documento, numero_documento
		FROM usuario
		WHERE id_condominio = $1
		ORDER BY fecha_creacion DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, condominioID, pageSize, offset)
	if err != nil {
		r.log.WithError(err).Error("Error querying users")
		return nil, 0, errors.DatabaseErrorf("error querying users").WithError(err)
	}
	defer rows.Close()

	users := make([]*models.User, 0, pageSize)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Nombre, &user.Apellido, &user.Telefono,
			&user.Celular, &user.CondominioID, &user.ApartmentID, &user.PIN, &user.Rol, &user.Estado, &user.CreatedAt, &user.UpdatedAt,
			&user.TipoDoc, &user.NumeroDoc,
		)
		if err != nil {
			r.log.WithError(err).Error("Error scanning user")
			return nil, 0, errors.DatabaseErrorf("error scanning user").WithError(err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		r.log.WithError(err).Error("Error iterating users")
		return nil, 0, errors.DatabaseErrorf("error iterating users").WithError(err)
	}

	return users, total, nil
}

// GetByApartment obtiene todos los usuarios asignados a un apartamento
func (r *UserRepositoryImpl) GetByApartment(ctx context.Context, apartmentID string) ([]*models.User, error) {
	query := `
		SELECT id_usuario, email, nombres, apellidos, telefono, celular, id_condominio, id_apartamento,
			'' AS password, rol, estado, fecha_creacion, ultimo_acceso, tipo_documento, numero_documento
		FROM usuario
		WHERE id_apartamento = $1
	`

	rows, err := r.db.QueryContext(ctx, query, apartmentID)
	if err != nil {
		r.log.WithError(err).Error("Error querying users by apartment")
		return nil, errors.DatabaseErrorf("error querying users by apartment").WithError(err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Nombre, &user.Apellido, &user.Telefono,
			&user.Celular, &user.CondominioID, &user.ApartmentID, &user.PIN, &user.Rol, &user.Estado, &user.CreatedAt, &user.UpdatedAt,
			&user.TipoDoc, &user.NumeroDoc,
		)
		if err != nil {
			r.log.WithError(err).Error("Error scanning user")
			return nil, errors.DatabaseErrorf("error scanning user").WithError(err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseErrorf("error iterating users").WithError(err)
	}

	return users, nil
}
