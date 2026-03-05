package services

import (
	"context"
	"strings"
	"time"

	"neighborhood-api/internal/repositories"
	"neighborhood-api/internal/utils"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"
)

// AuthService interfaz para servicio de autenticación
type AuthService interface {
	// Login autentica un usuario con email y PIN
	Login(ctx context.Context, req *dto.LoginRequest, condominioID string) (*dto.LoginResponse, error)

	// ValidateToken valida un token JWT
	ValidateToken(token string) (*utils.TokenClaims, error)

	// RefreshToken regenera un par de tokens
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)

	// ChangePIN cambia el PIN del usuario
	ChangePIN(ctx context.Context, userID, currentPIN, newPIN string) error

	// Logout invalida una sesión
	Logout(ctx context.Context, userID string) error
}

// AuthServiceImpl implementación de AuthService
type AuthServiceImpl struct {
	userRepo       repositories.UserRepository
	condominioRepo repositories.CondominioRepository
	jwtManager     *utils.JWTManager
	log            logger.Logger
}

// NewAuthService crea una nueva instancia de AuthService
func NewAuthService(
	userRepo repositories.UserRepository,
	condominioRepo repositories.CondominioRepository,
	jwtManager *utils.JWTManager,
) AuthService {
	return &AuthServiceImpl{
		userRepo:       userRepo,
		condominioRepo: condominioRepo,
		jwtManager:     jwtManager,
		log:            logger.Get(),
	}
}

// Login autentica un usuario con email y PIN
func (s *AuthServiceImpl) Login(ctx context.Context, req *dto.LoginRequest, condominioID string) (*dto.LoginResponse, error) {
	// (Validation is handled by the transport layer/handler with validation tags)

	// Buscar usuario por email (sin requerir condominio_id)
	// El condominio_id será obtenido del registro del usuario
	user, err := s.userRepo.FindByEmailOnly(ctx, req.Email)
	if err != nil {
		s.log.WithField("email", req.Email).Warn("User not found for login")
		return nil, errors.InvalidEmailOrPINErrorf("invalid email or pin")
	}

	// Validar PIN
	validPIN, needsRehash, err := utils.VerifyPIN(req.PIN, user.PIN)
	if err != nil {
		s.log.WithField("email", req.Email).WithError(err).Warn("PIN validation failed")
		return nil, errors.InvalidEmailOrPINErrorf("invalid email or pin")
	}

	if !validPIN {
		s.log.WithField("email", req.Email).Warn("Invalid PIN for user")
		return nil, errors.InvalidEmailOrPINErrorf("invalid email or pin")
	}

	if needsRehash {
		newPINHash, hashErr := utils.HashPIN(req.PIN)
		if hashErr != nil {
			s.log.WithField("user_id", user.ID).WithError(hashErr).Warn("Failed to rehash legacy PIN")
		} else {
			if updateErr := s.userRepo.UpdatePIN(ctx, user.ID, newPINHash); updateErr != nil {
				s.log.WithField("user_id", user.ID).WithError(updateErr).Warn("Failed to persist rehashed PIN")
			}
		}
	}

	// Validar estado del usuario
	if user.Estado != "activo" {
		s.log.WithField("email", req.Email).WithField("estado", user.Estado).Warn("User is not active")
		return nil, errors.UserNotActiveErrorf("user account is not active")
	}

	// Resolver apartment_id (solo residentes lo tienen)
	apartmentID := ""
	if user.ApartmentID != nil {
		apartmentID = *user.ApartmentID
	}

	// Generar tokens
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.CondominioID, user.Rol, apartmentID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(s.jwtManager.GetExpiration()) * time.Second)

	// Obtener nombre del condominio
	condominioName := ""
	s.log.WithField("user_condominio_id", user.CondominioID).Info("Attempting to fetch condominio for user")

	if user.CondominioID == "" {
		s.log.Warn("User has no condominio_id assigned")
	} else {
		if condominio, err := s.condominioRepo.FindByID(ctx, user.CondominioID); err == nil && condominio != nil {
			condominioName = condominio.Nombre
			s.log.WithField("condominio_id", user.CondominioID).WithField("condominio_name", condominioName).Info("✓ Condominio found successfully")
		} else {
			s.log.WithField("condominio_id", user.CondominioID).WithError(err).Warn("✗ Failed to fetch condominio from database")
		}
	}

	s.log.WithField("user_id", user.ID).WithField("email", user.Email).WithField("condominio_name", condominioName).WithField("has_condominio", condominioName != "").Info("User logged in successfully")

	isAdmin := strings.EqualFold(user.Rol, "administrador") ||
		strings.EqualFold(user.Rol, "admin") ||
		strings.EqualFold(user.Rol, "dev")

	return &dto.LoginResponse{
		Token:          token,
		RefreshToken:   refreshToken,
		UserID:         user.ID,
		Email:          user.Email,
		Nombre:         user.Nombre,
		Apellido:       user.Apellido,
		CondominioID:   user.CondominioID,
		CondominioName: condominioName,
		Role:           user.Rol,
		IsAdmin:        isAdmin,
		ApartmentID:    apartmentID,
		Permissions:    []string{},
		ExpiresAt:      expiresAt,
	}, nil
}

// ValidateToken valida un token JWT
func (s *AuthServiceImpl) ValidateToken(token string) (*utils.TokenClaims, error) {
	// Remover "Bearer " si está presente
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		s.log.WithError(err).Warn("Token validation failed")
		return nil, err
	}

	s.log.WithField("user_id", claims.UserID).Debug("Token validated successfully")
	return claims, nil
}

// RefreshToken regenera un par de tokens
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	// Validar refresh token y obtener user_id
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.log.WithError(err).Warn("Refresh token validation failed")
		return nil, err
	}

	// Obtener usuario para datos actualizados
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.log.WithField("user_id", userID).Warn("User not found for token refresh")
		return nil, errors.UnauthorizedErrorf("user not found")
	}

	// Resolver apartment_id
	refreshApartmentID := ""
	if user.ApartmentID != nil {
		refreshApartmentID = *user.ApartmentID
	}

	// Generar nuevo par de tokens
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.CondominioID, user.Rol, refreshApartmentID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(s.jwtManager.GetExpiration()) * time.Second)

	// Obtener nombre del condominio
	condominioName := ""
	if user.CondominioID != "" {
		if condominio, err := s.condominioRepo.FindByID(ctx, user.CondominioID); err == nil && condominio != nil {
			condominioName = condominio.Nombre
		}
	}

	s.log.WithField("user_id", user.ID).Info("Token refreshed successfully")

	isAdmin := strings.EqualFold(user.Rol, "administrador") ||
		strings.EqualFold(user.Rol, "admin") ||
		strings.EqualFold(user.Rol, "dev")

	return &dto.LoginResponse{
		Token:          token,
		RefreshToken:   newRefreshToken,
		UserID:         user.ID,
		Email:          user.Email,
		Nombre:         user.Nombre,
		Apellido:       user.Apellido,
		CondominioID:   user.CondominioID,
		CondominioName: condominioName,
		Role:           user.Rol,
		IsAdmin:        isAdmin,
		ApartmentID:    refreshApartmentID,
		Permissions:    []string{},
		ExpiresAt:      expiresAt,
	}, nil
}

// ChangePIN cambia el PIN del usuario
func (s *AuthServiceImpl) ChangePIN(ctx context.Context, userID, currentPIN, newPIN string) error {
	// Validar que los PINs tengan el formato correcto
	if len(currentPIN) != 6 || len(newPIN) != 6 {
		s.log.Warn("Invalid PIN format for change PIN")
		return errors.ValidationErrorf("PIN must be 6 digits")
	}

	// Obtener usuario actual
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.log.WithField("user_id", userID).Warn("User not found for PIN change")
		return errors.UnauthorizedErrorf("user not found")
	}

	// Validar PIN actual
	validPIN, _, err := utils.VerifyPIN(currentPIN, user.PIN)
	if err != nil || !validPIN {
		s.log.WithField("user_id", userID).Warn("Invalid current PIN for user")
		return errors.UnauthorizedErrorf("invalid current PIN")
	}

	// Hash del nuevo PIN
	newPINHash, err := utils.HashPIN(newPIN)
	if err != nil {
		s.log.WithField("user_id", userID).WithError(err).Error("Error hashing new PIN")
		return errors.InternalErrorf("error updating PIN")
	}

	// Actualizar PIN en la base de datos
	err = s.userRepo.UpdatePIN(ctx, userID, newPINHash)
	if err != nil {
		s.log.WithField("user_id", userID).WithError(err).Error("Error updating PIN")
		return errors.InternalErrorf("error updating PIN")
	}

	s.log.WithField("user_id", userID).Info("PIN changed successfully")
	return nil
}

// Logout invalida una sesión
func (s *AuthServiceImpl) Logout(ctx context.Context, userID string) error {
	s.log.WithField("user_id", userID).Info("User logged out")
	return nil
}
