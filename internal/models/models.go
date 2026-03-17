package models

import (
	"time"

	"github.com/lib/pq"
)

// User representa un usuario en el sistema
type User struct {
	ID           string     `db:"id_usuario" json:"id"`
	Email        string     `db:"email" json:"email"`
	Nombre       string     `db:"nombres" json:"nombre"`
	Apellido     string     `db:"apellidos" json:"apellido"`
	Telefono     string     `db:"telefono" json:"telefono"`
	Celular      *string    `db:"celular" json:"celular"`
	TipoDoc      string     `db:"tipo_documento" json:"tipo_documento"`
	NumeroDoc    string     `db:"numero_documento" json:"numero_documento"`
	CondominioID string     `db:"id_condominio" json:"condominio_id"`
	ApartmentID  *string    `db:"id_apartamento" json:"apartamento_id,omitempty"`
	PIN          string     `db:"password" json:"-"` // Never expose in JSON
	Rol          string     `db:"rol" json:"rol"`
	Estado       string     `db:"estado" json:"estado"`
	CreatedAt    time.Time  `db:"fecha_creacion" json:"created_at"`
	UpdatedAt    *time.Time `db:"ultimo_acceso" json:"updated_at"` // Puede ser NULL
}

// Apartment representa un apartamento
type Apartment struct {
	ID            string    `db:"id_apartamento" json:"id"`
	CondominioID  string    `db:"id_condominio" json:"condominio_id"`
	Numero        string    `db:"numero" json:"numero"`
	Bloque        *string   `db:"bloque" json:"bloque"`
	Torre         *string   `db:"torre" json:"torre"`
	Piso          *string   `db:"piso" json:"piso"`
	PropietarioID *string   `db:"id_propietario" json:"propietario_id"`
	AreaPrivada   *float64  `db:"area_privada" json:"area_privada"`
	AreaComun     *float64  `db:"area_comun" json:"area_comun"`
	FechaRegistro time.Time `db:"fecha_registro" json:"fecha_registro"`
	Estado        string    `db:"estado" json:"estado"`
}

// Invoice representa una factura/recibo en el sistema
type Invoice struct {
	ID            string     `db:"id_factura" json:"id"`
	CondominioID  string     `db:"id_condominio" json:"condominio_id"`
	UsuarioID     string     `db:"id_usuario" json:"usuario_id"`
	ApartmentID   *string    `db:"id_apartamento" json:"apartment_id"`
	ServiceID     string     `db:"id_servicio" json:"service_id"`
	InvoiceNumber string     `db:"numero_factura" json:"invoice_number"`
	BaseAmount    float64    `db:"monto_base" json:"base_amount"`
	Discount      float64    `db:"descuento" json:"discount"`
	InterestRate  float64    `db:"interes_mora" json:"interest_rate"`
	Total         float64    `db:"total" json:"total"`
	Status        string     `db:"estado" json:"status"` // pendiente, pagada, cancelada
	IssuedDate    time.Time  `db:"fecha_emision" json:"issued_date"`
	DueDate       *time.Time `db:"fecha_vencimiento" json:"due_date"`
	PaidDate      *time.Time `db:"fecha_pago" json:"paid_date"`
	PaymentMethod *string    `db:"metodo_pago" json:"payment_method"` // transferencia, cheque, efectivo, etc
	Notes         *string    `db:"notas" json:"notes"`
}

// Reservation representa una reserva
type Reservation struct {
	ID                string    `db:"id_reserva" json:"id"`
	CondominioID      string    `db:"id_condominio" json:"condominio_id"`
	UsuarioID         string    `db:"id_usuario" json:"usuario_id"`
	ApartmentID       *string   `db:"id_apartamento" json:"apartment_id"`
	EspacioID         string    `db:"id_espacio" json:"espacio_id"`
	FechaSolicitud    time.Time `db:"fecha_solicitud" json:"fecha_solicitud"`
	FechaInicio       time.Time `db:"fecha_inicio" json:"fecha_inicio"`
	FechaFin          time.Time `db:"fecha_fin" json:"fecha_fin"`
	PersonasEsperadas *int      `db:"personas_esperadas" json:"personas_esperadas"`
	Observaciones     *string   `db:"observaciones" json:"observaciones"`
	ValorBase         *float64  `db:"valor_base" json:"valor_base"`
	CostoTotal        *float64  `db:"costo_total" json:"costo_total"`
	Pagado            bool      `db:"pagado" json:"pagado"`
	Estado            string    `db:"estado" json:"estado"`
	// Populated via JOIN in list queries — not stored in DB
	EspacioNombre string `db:"-" json:"-"`
	UsuarioNombre string `db:"-" json:"-"`
}

// Space representa un espacio comun
type Space struct {
	ID           string     `db:"id_espacio" json:"id"`
	CondominioID string     `db:"id_condominio" json:"condominio_id"`
	Nombre       string     `db:"nombre" json:"nombre"`
	Descripcion  *string    `db:"descripcion" json:"descripcion"`
	CostoHora    *float64   `db:"costo_hora" json:"costo_hora"`
	Estado       string     `db:"estado" json:"estado"`
	CreatedAt    time.Time  `db:"fecha_creacion" json:"created_at"`
	UpdatedAt    *time.Time `db:"fecha_actualizacion" json:"updated_at"`
}

// Communication representa un comunicado
type Communication struct {
	ID                 string         `db:"id_comunicado" json:"id"`
	Titulo             string         `db:"titulo" json:"titulo"`
	Contenido          string         `db:"contenido" json:"contenido"`
	Fecha              time.Time      `db:"fecha_publicacion" json:"fecha"`
	Autor              string         `db:"id_usuario_emisor" json:"autor"`
	CondominioID       string         `db:"id_condominio" json:"condominio_id"`
	ProgramadoPara     *time.Time     `db:"programado_para" json:"programado_para,omitempty"`
	RolesDestino       pq.StringArray `db:"roles_destino" json:"roles_destino,omitempty"`
	Icono              string         `db:"icono" json:"icono"`
	PermiteComentarios bool           `db:"permite_comentarios" json:"permite_comentarios"`
	Publicado          bool           `db:"publicado" json:"publicado"`
	CreatedAt          time.Time      `db:"fecha_publicacion" json:"created_at"`
	UpdatedAt          time.Time      `db:"fecha_publicacion" json:"updated_at"`
	// Computed fields (not in DB directly)
	Leido         bool `db:"-" json:"leido"`
	NumComentarios int  `db:"-" json:"num_comentarios"`
}

// ComunicadoLectura representa el tracking de lectura de un comunicado
type ComunicadoLectura struct {
	ID            string    `db:"id" json:"id"`
	IDComunicado  string    `db:"id_comunicado" json:"id_comunicado"`
	IDUsuario     string    `db:"id_usuario" json:"id_usuario"`
	LeidoEn       time.Time `db:"leido_en" json:"leido_en"`
}

// ComunicadoComentario representa un comentario en un comunicado
type ComunicadoComentario struct {
	ID             string    `db:"id_comentario" json:"id"`
	IDComunicado   string    `db:"id_comunicado" json:"id_comunicado"`
	IDUsuario      string    `db:"id_usuario" json:"id_usuario"`
	Contenido      string    `db:"contenido" json:"contenido"`
	FechaCreacion  time.Time `db:"fecha_creacion" json:"fecha_creacion"`
	// Joined fields
	AutorNombre   string `db:"-" json:"autor_nombre,omitempty"`
	AutorApellido string `db:"-" json:"autor_apellido,omitempty"`
}

// Package representa un paquete/encomienda recibido
type Package struct {
	ID              string     `db:"id_paquete" json:"id"`
	ApartmentID     string     `db:"id_apartamento" json:"apartment_id"`
	Resident        string     `db:"residente" json:"resident"`
	Carrier         string     `db:"transportadora" json:"carrier"`
	Notes           *string    `db:"notas" json:"notes"`
	ReceivedAt      time.Time  `db:"fecha_recibido" json:"received_at"`
	DeliveredAt     *time.Time `db:"fecha_entregado" json:"delivered_at"`
	CondominioID    string     `db:"id_condominio" json:"condominio_id"`
	CreatedAt       time.Time  `db:"fecha_creacion" json:"created_at"`
	UpdatedAt       time.Time  `db:"fecha_actualizacion" json:"updated_at"`
	ApartmentNumber *string    `db:"numero" json:"apartment_number"`
	ApartmentTower  *string    `db:"torre" json:"apartment_tower"`
	ApartmentBlock  *string    `db:"bloque" json:"apartment_block"`
	ApartmentFloor  *string    `db:"piso" json:"apartment_floor"`
}

// Notification representa una notificación in-app personal para un usuario
type Notification struct {
	ID            string    `db:"id" json:"id"`
	UserID        string    `db:"id_usuario" json:"user_id"`
	CondominioID  string    `db:"id_condominio" json:"condominio_id"`
	Tipo          string    `db:"tipo" json:"tipo"` // "paquete" | "reserva"
	Titulo        string    `db:"titulo" json:"titulo"`
	Mensaje       string    `db:"mensaje" json:"mensaje"`
	Leido         bool      `db:"leido" json:"leido"`
	ReferenciaID  *string   `db:"referencia_id" json:"referencia_id"`
	FechaCreacion time.Time `db:"fecha_creacion" json:"fecha_creacion"`
}

// Condominio representa un condominio
// Condominio representa un condominio en el sistema
type Condominio struct {
	ID                 string    `db:"id_condominio" json:"id"`
	Nombre             string    `db:"nombre" json:"nombre"`
	Direccion          string    `db:"direccion" json:"direccion"`
	Ciudad             string    `db:"ciudad" json:"ciudad"`
	Telefono           *string   `db:"telefono" json:"telefono"`
	Email              *string   `db:"email" json:"email"`
	LogoUrl            *string   `db:"logo_url" json:"logo_url"`
	NIT                *string   `db:"nit" json:"nit"`
	RepresentanteLegal *string   `db:"representante_legal" json:"representante_legal"`
	FechaCreacion      time.Time `db:"fecha_creacion" json:"fecha_creacion"`
	Estado             string    `db:"estado" json:"estado"`
	PermiteSoporte     bool      `db:"permite_soporte" json:"permite_soporte"`
}
