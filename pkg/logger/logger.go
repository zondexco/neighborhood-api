package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wrapper para aplicación
type Logger struct {
	logger zerolog.Logger
}

// Config para inicializar logger
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	EnableFile bool
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	AlsoStdout bool
}

// Initialize inicializa el logger global
func Initialize(cfg Config) {
	outputs := make([]io.Writer, 0, 2)

	if cfg.AlsoStdout || !cfg.EnableFile {
		if cfg.Format == "text" {
			outputs = append(outputs, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
		} else {
			outputs = append(outputs, os.Stdout)
		}
	}

	if cfg.EnableFile {
		if cfg.FilePath == "" {
			cfg.FilePath = "logs/neighborhood-api.log"
		}
		if cfg.MaxSizeMB <= 0 {
			cfg.MaxSizeMB = 100
		}
		if cfg.MaxBackups <= 0 {
			cfg.MaxBackups = 10
		}
		if cfg.MaxAgeDays <= 0 {
			cfg.MaxAgeDays = 14
		}

		dir := filepath.Dir(cfg.FilePath)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}

		outputs = append(outputs, &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		})
	}

	if len(outputs) == 0 {
		outputs = append(outputs, os.Stdout)
	}

	output := io.Writer(zerolog.MultiLevelWriter(outputs...))

	// Configurar nivel
	level := zerolog.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	// Configurar logger global
	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Logger()
}

// Get retorna el logger global
func Get() Logger {
	return Logger{logger: log.Logger}
}

// WithField añade un field al logger
func (l Logger) WithField(key string, value interface{}) Logger {
	return Logger{logger: l.logger.With().Interface(key, value).Logger()}
}

// WithFields añade múltiples fields
func (l Logger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return Logger{logger: ctx.Logger()}
}

// WithError añade un error
func (l Logger) WithError(err error) Logger {
	return Logger{logger: l.logger.With().Err(err).Logger()}
}

// Debug logs a debug message
func (l Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Info logs an info message
func (l Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Warn logs a warning message
func (l Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Error logs an error message
func (l Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func (l Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}
