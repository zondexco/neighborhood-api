package utils

import (
	"fmt"
	"time"

	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"
	"neighborhood-api/pkg/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims estructura de claims del JWT
type TokenClaims struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	CondominioID string `json:"condominio_id"`
	Role         string `json:"role"`
	ApartmentID  string `json:"apartment_id"` // "" si el usuario no tiene apartamento
	jwt.RegisteredClaims
}

// TokenResponse estructura de respuesta de token
type TokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Type         string    `json:"type"`
}

// Config estructura de configuración JWT (exportada para tests)
type JWTConfig = types.JWTConfig

// JWTManager maneja la creación y validación de JWT
type JWTManager struct {
	config types.JWTConfig
	log    logger.Logger
}

// NewJWTManager crea un nuevo JWTManager
func NewJWTManager(cfg types.JWTConfig) *JWTManager {
	return &JWTManager{
		config: cfg,
		log:    logger.Get(),
	}
}

// GenerateToken genera un token JWT
func (jm *JWTManager) GenerateToken(userID, email, condominioID, role, apartmentID string) (string, error) {
	expiresAt := time.Now().Add(time.Duration(jm.config.Expiration) * time.Second)

	claims := TokenClaims{
		UserID:       userID,
		Email:        email,
		CondominioID: condominioID,
		Role:         role,
		ApartmentID:  apartmentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.config.Secret))
	if err != nil {
		jm.log.WithError(err).Error("Failed to generate JWT token")
		return "", errors.InternalErrorf("failed to generate JWT token").WithError(err)
	}

	jm.log.WithField("user_id", userID).WithField("expires_at", expiresAt).Debug("JWT token generated")
	return tokenString, nil
}

// GenerateRefreshToken genera un refresh token
func (jm *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	expiresAt := time.Now().Add(time.Duration(jm.config.RefreshExpiration) * time.Second)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.config.RefreshSecret))
	if err != nil {
		jm.log.WithError(err).Error("Failed to generate refresh token")
		return "", errors.InternalErrorf("failed to generate refresh token").WithError(err)
	}

	jm.log.WithField("user_id", userID).Debug("Refresh token generated")
	return tokenString, nil
}

// ValidateToken valida y parsea un token JWT
func (jm *JWTManager) ValidateToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar el algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jm.config.Secret), nil
	})

	if err != nil {
		jm.log.WithError(err).Warn("Failed to parse JWT token")
		return nil, errors.UnauthorizedErrorf("invalid token: %v", err).WithError(err)
	}

	if !token.Valid {
		jm.log.Warn("JWT token is not valid")
		return nil, errors.UnauthorizedErrorf("token is not valid")
	}

	return claims, nil
}

// ValidateRefreshToken valida un refresh token
func (jm *JWTManager) ValidateRefreshToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jm.config.RefreshSecret), nil
	})

	if err != nil {
		jm.log.WithError(err).Warn("Failed to parse refresh token")
		return "", errors.UnauthorizedErrorf("invalid refresh token").WithError(err)
	}

	if !token.Valid {
		jm.log.Warn("Refresh token is not valid")
		return "", errors.UnauthorizedErrorf("refresh token is not valid")
	}

	return claims.Subject, nil
}

// GetExpiration retorna el tiempo de expiración del token en segundos
func (jm *JWTManager) GetExpiration() int {
	return jm.config.Expiration
}

// RefreshTokenPair genera un nuevo par de tokens a partir de un refresh token válido
func (jm *JWTManager) RefreshTokenPair(refreshToken, userID, email, condominioID, role, apartmentID string) (*TokenResponse, error) {
	// Validar refresh token
	subject, err := jm.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verificar que el userID coincida
	if subject != userID {
		jm.log.WithField("expected_user", userID).WithField("token_subject", subject).Warn("Refresh token subject mismatch")
		return nil, errors.UnauthorizedErrorf("refresh token mismatch")
	}

	// Generar nuevo token
	newToken, err := jm.GenerateToken(userID, email, condominioID, role, apartmentID)
	if err != nil {
		return nil, err
	}

	// Generar nuevo refresh token
	newRefreshToken, err := jm.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(jm.config.Expiration) * time.Second)

	return &TokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		Type:         "Bearer",
	}, nil
}
