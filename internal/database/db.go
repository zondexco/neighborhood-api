package database

import (
	"database/sql"
	"fmt"
	"time"

	"neighborhood-api/internal/config"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB estructura para manejo de base de datos
type DB struct {
	*sql.DB
	log logger.Logger
}

// New crea una nueva conexión a la base de datos
func New(cfg config.DatabaseConfig) (*DB, error) {
	log := logger.Get()

	// Construir connection string (pgx soporta estos flags).
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d prefer_simple_protocol=true binary_parameters=no statement_cache_capacity=0 default_query_exec_mode=simple_protocol",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
		cfg.ConnTimeout,
	)

	log.WithField("host", cfg.Host).
		WithField("database", cfg.Name).
		WithField("prefer_simple_protocol", true).
		WithField("binary_parameters", "no").
		WithField("statement_cache_capacity", 0).
		WithField("default_query_exec_mode", "simple_protocol").
		Info("Connecting to database")

	// Abrir conexión
	// Usamos pgx stdlib para poder forzar simple protocol y evitar prepared statements en PgBouncer.
	db, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		log.WithError(err).Error("Failed to open database connection")
		return nil, errors.DatabaseErrorf("failed to open database connection: %v", err).WithError(err)
	}

	// Configurar connection pool
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MinConnections)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Verificar conexión
	err = db.Ping()
	if err != nil {
		log.WithError(err).Error("Failed to ping database")
		return nil, errors.DatabaseErrorf("failed to ping database: %v", err).WithError(err)
	}

	log.Info("Database connection established successfully")

	return &DB{
		DB:  db,
		log: log,
	}, nil
}

// Health verifica el estado de la base de datos
func (db *DB) Health() error {
	ctx, cancel := ContextWithTimeout(5 * time.Second)
	defer cancel()

	err := db.PingContext(ctx)
	if err != nil {
		db.log.WithError(err).Error("Database health check failed")
		return errors.DatabaseErrorf("database health check failed").WithError(err)
	}

	return nil
}

// BeginTx inicia una transacción
func (db *DB) BeginTx() (*sql.Tx, error) {
	ctx, cancel := ContextWithTimeout(30 * time.Second)
	defer cancel()

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		db.log.WithError(err).Error("Failed to begin transaction")
		return nil, errors.DatabaseErrorf("failed to begin transaction").WithError(err)
	}

	return tx, nil
}

// Close cierra la conexión a la base de datos
func (db *DB) Close() error {
	db.log.Info("Closing database connection")
	return db.DB.Close()
}
