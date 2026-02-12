package repositories

import (
	"context"
	"database/sql"

	"neighborhood-api/internal/database"
	"neighborhood-api/internal/models"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"
)

// InvoiceRepositoryImpl implementación de InvoiceRepository
type InvoiceRepositoryImpl struct {
	db  *database.DB
	log logger.Logger
}

// NewInvoiceRepository crea una nueva instancia de InvoiceRepository
func NewInvoiceRepository(db *database.DB) InvoiceRepository {
	return &InvoiceRepositoryImpl{
		db:  db,
		log: logger.Get(),
	}
}

// Create crea una nueva factura
func (r *InvoiceRepositoryImpl) Create(ctx context.Context, invoice *models.Invoice) error {
	query := `
		INSERT INTO factura (id_factura, id_condominio, id_usuario, id_apartamento, id_servicio, numero_factura, monto_base, descuento, interes_mora, total, estado, fecha_vencimiento, fecha_emision, metodo_pago, notas)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING fecha_emision, fecha_pago
	`

	var service interface{} = invoice.ServiceID
	if invoice.ServiceID == "" {
		service = nil
	}

	err := r.db.QueryRowContext(ctx, query,
		invoice.ID, invoice.CondominioID, invoice.UsuarioID, invoice.ApartmentID, service,
		invoice.InvoiceNumber, invoice.BaseAmount, invoice.Discount, invoice.InterestRate, invoice.Total,
		invoice.Status, invoice.DueDate, invoice.IssuedDate, invoice.PaymentMethod, invoice.Notes,
	).Scan(&invoice.IssuedDate, &invoice.PaidDate)

	if err != nil {
		r.log.WithError(err).WithField("invoice_id", invoice.ID).Error("Error creating invoice")
		return errors.DatabaseErrorf("error creating invoice").WithError(err)
	}

	r.log.WithField("invoice_id", invoice.ID).WithField("condominio_id", invoice.CondominioID).Info("Invoice created successfully")
	return nil
}

// FindByID busca una factura por ID
func (r *InvoiceRepositoryImpl) FindByID(ctx context.Context, invoiceID, condominioID string) (*models.Invoice, error) {
	invoice := &models.Invoice{}

	query := `
		SELECT id_factura, id_condominio, id_usuario, id_apartamento, COALESCE(id_servicio::text, '') AS id_servicio, numero_factura, monto_base, descuento, interes_mora, total, estado, fecha_vencimiento, fecha_emision, fecha_pago, metodo_pago, notas
		FROM factura
		WHERE id_factura = $1 AND id_condominio = $2
	`

	err := r.db.QueryRowContext(ctx, query, invoiceID, condominioID).Scan(
		&invoice.ID, &invoice.CondominioID, &invoice.UsuarioID, &invoice.ApartmentID, &invoice.ServiceID,
		&invoice.InvoiceNumber, &invoice.BaseAmount, &invoice.Discount, &invoice.InterestRate, &invoice.Total,
		&invoice.Status, &invoice.DueDate, &invoice.IssuedDate, &invoice.PaidDate, &invoice.PaymentMethod, &invoice.Notes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.WithField("invoice_id", invoiceID).Debug("Invoice not found")
			return nil, errors.NotFoundErrorf("invoice not found")
		}
		r.log.WithError(err).WithField("invoice_id", invoiceID).Error("Error querying invoice")
		return nil, errors.DatabaseErrorf("error querying invoice").WithError(err)
	}

	return invoice, nil
}

// GetByCondominio obtiene todas las facturas de un condominio con paginación
func (r *InvoiceRepositoryImpl) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Contar total
	countQuery := `SELECT COUNT(*) FROM factura WHERE id_condominio = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, condominioID).Scan(&total)
	if err != nil {
		r.log.WithError(err).Error("Error counting invoices")
		return nil, 0, errors.DatabaseErrorf("error counting invoices").WithError(err)
	}

	// Obtener facturas
	query := `
		SELECT id_factura, id_condominio, id_apartamento, numero_factura, monto_base, total, estado, fecha_vencimiento, fecha_emision, fecha_pago, metodo_pago, notas, id_usuario, COALESCE(id_servicio::text, '') AS id_servicio
		FROM factura
		WHERE id_condominio = $1
		ORDER BY fecha_emision DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, condominioID, pageSize, offset)
	if err != nil {
		r.log.WithError(err).Error("Error querying invoices")
		return nil, 0, errors.DatabaseErrorf("error querying invoices").WithError(err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.CondominioID, &invoice.ApartmentID, &invoice.InvoiceNumber,
			&invoice.BaseAmount, &invoice.Total, &invoice.Status, &invoice.DueDate,
			&invoice.IssuedDate, &invoice.PaidDate, &invoice.PaymentMethod, &invoice.Notes,
			&invoice.UsuarioID, &invoice.ServiceID,
		)
		if err != nil {
			r.log.WithError(err).Error("Error scanning invoice")
			return nil, 0, errors.DatabaseErrorf("error scanning invoice").WithError(err)
		}
		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		r.log.WithError(err).Error("Error iterating invoices")
		return nil, 0, errors.DatabaseErrorf("error iterating invoices").WithError(err)
	}

	return invoices, total, nil
}

// GetByUsuario obtiene todas las facturas de un usuario dentro de un condominio
func (r *InvoiceRepositoryImpl) GetByUsuario(ctx context.Context, usuarioID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM factura WHERE id_condominio = $1 AND id_usuario = $2`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, condominioID, usuarioID).Scan(&total); err != nil {
		r.log.WithError(err).Error("Error counting user invoices")
		return nil, 0, errors.DatabaseErrorf("error counting invoices").WithError(err)
	}

	query := `
		SELECT id_factura, id_condominio, id_apartamento, numero_factura, monto_base, total, estado, fecha_vencimiento, fecha_emision, fecha_pago, metodo_pago, notas, id_usuario, COALESCE(id_servicio::text, '') AS id_servicio
		FROM factura
		WHERE id_condominio = $1 AND id_usuario = $2
		ORDER BY fecha_emision DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, condominioID, usuarioID, pageSize, offset)
	if err != nil {
		r.log.WithError(err).Error("Error querying user invoices")
		return nil, 0, errors.DatabaseErrorf("error querying invoices").WithError(err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		if err := rows.Scan(
			&invoice.ID, &invoice.CondominioID, &invoice.ApartmentID, &invoice.InvoiceNumber,
			&invoice.BaseAmount, &invoice.Total, &invoice.Status, &invoice.DueDate,
			&invoice.IssuedDate, &invoice.PaidDate, &invoice.PaymentMethod, &invoice.Notes,
			&invoice.UsuarioID, &invoice.ServiceID,
		); err != nil {
			r.log.WithError(err).Error("Error scanning user invoice")
			return nil, 0, errors.DatabaseErrorf("error scanning invoice").WithError(err)
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		r.log.WithError(err).Error("Error iterating user invoices")
		return nil, 0, errors.DatabaseErrorf("error iterating invoices").WithError(err)
	}

	return invoices, total, nil
}

// GetByApartment obtiene todas las facturas de un apartamento
func (r *InvoiceRepositoryImpl) GetByApartment(ctx context.Context, apartmentID, condominioID string, page, pageSize int) ([]*models.Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Contar total
	countQuery := `SELECT COUNT(*) FROM factura WHERE id_apartamento = $1 AND id_condominio = $2`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, apartmentID, condominioID).Scan(&total)
	if err != nil {
		r.log.WithError(err).Error("Error counting apartment invoices")
		return nil, 0, errors.DatabaseErrorf("error counting invoices").WithError(err)
	}

	// Obtener facturas
	query := `
		SELECT id_factura, id_condominio, id_usuario, id_apartamento, COALESCE(id_servicio::text, '') AS id_servicio, numero_factura, monto_base, descuento, interes_mora, total, estado, fecha_vencimiento, fecha_emision, fecha_pago, metodo_pago, notas
		FROM factura
		WHERE id_apartamento = $1 AND id_condominio = $2
		ORDER BY fecha_emision DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, apartmentID, condominioID, pageSize, offset)
	if err != nil {
		r.log.WithError(err).Error("Error querying apartment invoices")
		return nil, 0, errors.DatabaseErrorf("error querying invoices").WithError(err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.CondominioID, &invoice.UsuarioID, &invoice.ApartmentID, &invoice.ServiceID,
			&invoice.InvoiceNumber, &invoice.BaseAmount, &invoice.Discount, &invoice.InterestRate, &invoice.Total,
			&invoice.Status, &invoice.DueDate, &invoice.IssuedDate, &invoice.PaidDate, &invoice.PaymentMethod, &invoice.Notes,
		)
		if err != nil {
			r.log.WithError(err).Error("Error scanning invoice")
			return nil, 0, errors.DatabaseErrorf("error scanning invoice").WithError(err)
		}
		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		r.log.WithError(err).Error("Error iterating invoices")
		return nil, 0, errors.DatabaseErrorf("error iterating invoices").WithError(err)
	}

	return invoices, total, nil
}

// Update actualiza una factura
func (r *InvoiceRepositoryImpl) Update(ctx context.Context, invoice *models.Invoice) error {
	query := `
		UPDATE factura
		SET monto_base = $1, descuento = $2, interes_mora = $3, total = $4, estado = $5, fecha_vencimiento = $6, metodo_pago = $7, notas = $8, fecha_pago = $9
		WHERE id_factura = $10 AND id_condominio = $11
		RETURNING fecha_emision, fecha_pago
	`

	err := r.db.QueryRowContext(ctx, query,
		invoice.BaseAmount, invoice.Discount, invoice.InterestRate, invoice.Total, invoice.Status, invoice.DueDate,
		invoice.PaymentMethod, invoice.Notes, invoice.PaidDate,
		invoice.ID, invoice.CondominioID,
	).Scan(&invoice.IssuedDate, &invoice.PaidDate)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.WithField("invoice_id", invoice.ID).Warn("Invoice not found for update")
			return errors.NotFoundErrorf("invoice not found")
		}
		r.log.WithError(err).WithField("invoice_id", invoice.ID).Error("Error updating invoice")
		return errors.DatabaseErrorf("error updating invoice").WithError(err)
	}

	r.log.WithField("invoice_id", invoice.ID).Info("Invoice updated successfully")
	return nil
}

// Delete elimina una factura
func (r *InvoiceRepositoryImpl) Delete(ctx context.Context, invoiceID, condominioID string) error {
	query := `DELETE FROM factura WHERE id_factura = $1 AND id_condominio = $2`

	result, err := r.db.ExecContext(ctx, query, invoiceID, condominioID)
	if err != nil {
		r.log.WithError(err).WithField("invoice_id", invoiceID).Error("Error deleting invoice")
		return errors.DatabaseErrorf("error deleting invoice").WithError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithError(err).Error("Error getting rows affected")
		return errors.DatabaseErrorf("error deleting invoice").WithError(err)
	}

	if rowsAffected == 0 {
		r.log.WithField("invoice_id", invoiceID).Warn("Invoice not found for deletion")
		return errors.NotFoundErrorf("invoice not found")
	}

	r.log.WithField("invoice_id", invoiceID).Info("Invoice deleted successfully")
	return nil
}
