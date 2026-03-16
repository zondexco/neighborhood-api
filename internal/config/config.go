package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"
	"neighborhood-api/pkg/types"

	"github.com/joho/godotenv"
)

// Config estructura general de configuración
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      types.JWTConfig
	CORS     CORSConfig
	Logging  LoggingConfig
	App      AppConfig
}

// ServerConfig configuración del servidor
type ServerConfig struct {
	Port         int
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig configuración de base de datos
type DatabaseConfig struct {
	Host                      string
	Port                      int
	User                      string
	Password                  string
	Name                      string
	SSLMode                   string
	MaxConnections            int
	MinConnections            int
	ConnTimeout               int
	DisablePreparedStatements bool
	SupabaseURL               string
	AnonKey                   string
	ServiceRoleKey            string
}

// CORSConfig configuración de CORS
type CORSConfig struct {
	AllowedOrigins   string
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredentials bool
	MaxAge           int
}

// LoggingConfig configuración de logging
type LoggingConfig struct {
	Level      string
	Format     string
	EnableFile bool
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	AlsoStdout bool
}

// AppConfig configuración de aplicación
type AppConfig struct {
	Name    string
	Version string
}

// Load carga la configuración desde .env
func Load() (*Config, error) {
	// Cargar .env (ignorar si no existe)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:         getIntEnv("SERVER_PORT", 8080),
			Environment:  getStringEnv("SERVER_ENV", "development"),
			ReadTimeout:  getIntEnv("SERVER_READ_TIMEOUT", 10),
			WriteTimeout: getIntEnv("SERVER_WRITE_TIMEOUT", 10),
		},
		Database: DatabaseConfig{
			Host:                      getStringEnv("DB_HOST", "localhost"),
			Port:                      getIntEnv("DB_PORT", 5432),
			User:                      getStringEnv("DB_USER", "postgres"),
			Password:                  getStringEnv("DB_PASSWORD", ""),
			Name:                      getStringEnv("DB_NAME", "postgres"),
			SSLMode:                   getStringEnv("DB_SSLMODE", "disable"),
			MaxConnections:            getIntEnv("DB_MAX_CONN", 50),
			MinConnections:            getIntEnv("DB_MIN_CONN", 10),
			ConnTimeout:               getIntEnv("DB_CONN_TIMEOUT", 30),
			DisablePreparedStatements: getBoolEnv("DB_DISABLE_PREPARED_STATEMENTS", true),
			SupabaseURL:               getStringEnv("SUPABASE_URL", ""),
			AnonKey:                   getStringEnv("SUPABASE_ANON_KEY", ""),
			ServiceRoleKey:            getStringEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		},
		JWT: types.JWTConfig{
			Secret:            getStringEnv("JWT_SECRET", ""),
			Expiration:        getIntEnv("JWT_EXPIRATION", 86400),
			RefreshSecret:     getStringEnv("JWT_REFRESH_SECRET", ""),
			RefreshExpiration: getIntEnv("JWT_REFRESH_EXPIRATION", 604800),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getStringEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8081"),
			AllowedMethods:   getStringEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders:   getStringEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
			AllowCredentials: getBoolEnv("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getIntEnv("CORS_MAX_AGE", 3600),
		},
		Logging: LoggingConfig{
			Level:      getStringEnv("LOG_LEVEL", "info"),
			Format:     getStringEnv("LOG_FORMAT", "json"),
			EnableFile: getBoolEnv("LOG_ENABLE_FILE", true),
			FilePath:   getStringEnv("LOG_FILE_PATH", "logs/neighborhood-api.log"),
			MaxSizeMB:  getIntEnv("LOG_MAX_SIZE_MB", 100),
			MaxBackups: getIntEnv("LOG_MAX_BACKUPS", 10),
			MaxAgeDays: getIntEnv("LOG_MAX_AGE_DAYS", 14),
			Compress:   getBoolEnv("LOG_COMPRESS", true),
			AlsoStdout: getBoolEnv("LOG_ALSO_STDOUT", true),
		},
		App: AppConfig{
			Name:    getStringEnv("APP_NAME", "neighborhood-api"),
			Version: getStringEnv("APP_VERSION", "0.1.0"),
		},
	}

	// Validar configuración crítica
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate valida que la configuración sea válida
func (c *Config) Validate() error {
	log := logger.Get()

	// JWT Secret
	if len(c.JWT.Secret) < 32 {
		log.Error("JWT_SECRET must be at least 32 characters")
		return errors.InternalErrorf("JWT_SECRET configuration invalid")
	}

	// Database
	if c.Database.Host == "" {
		log.Error("DB_HOST is required")
		return errors.InternalErrorf("Database configuration invalid")
	}

	// Supabase (si se quiere usar Supabase en lugar de PostgreSQL directo)
	// if c.Database.SupabaseURL == "" {
	// 	log.Warn("SUPABASE_URL not configured, using direct PostgreSQL")
	// }

	log.Info(fmt.Sprintf("Configuration validated for environment: %s", c.Server.Environment))
	return nil
}

// Helper functions

func getStringEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := strings.ToLower(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// String implementa Stringer para debugging
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Server: %s:%d, DB: %s:%d, Env: %s, LogLevel: %s}",
		"0.0.0.0", c.Server.Port,
		c.Database.Host, c.Database.Port,
		c.Server.Environment,
		c.Logging.Level,
	)
}
