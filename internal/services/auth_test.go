package services

import (
	"context"
	"fmt"
	"testing"

	"neighborhood-api/internal/models"
	"neighborhood-api/internal/utils"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/logger"
)

// MockUserRepository mock de UserRepository para tests
type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string, condominioID string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email && user.CondominioID == condominioID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) FindByEmailOnly(ctx context.Context, email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	if user, ok := m.users[userID]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdatePIN(ctx context.Context, userID, newPINHash string) error {
	if user, ok := m.users[userID]; ok {
		user.PIN = newPINHash
		m.users[userID] = user
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	delete(m.users, userID)
	return nil
}

func (m *MockUserRepository) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.User, int, error) {
	var users []*models.User
	for _, user := range m.users {
		if user.CondominioID == condominioID {
			users = append(users, user)
		}
	}
	return users, len(users), nil
}

// MockCondominioRepository mock de CondominioRepository para tests
type MockCondominioRepository struct {
	condominios map[string]*models.Condominio
}

func NewMockCondominioRepository() *MockCondominioRepository {
	return &MockCondominioRepository{
		condominios: make(map[string]*models.Condominio),
	}
}

func (m *MockCondominioRepository) FindByID(ctx context.Context, condominioID string) (*models.Condominio, error) {
	if cond, ok := m.condominios[condominioID]; ok {
		return cond, nil
	}
	return nil, fmt.Errorf("condominio not found")
}

func (m *MockCondominioRepository) GetAll(ctx context.Context) ([]*models.Condominio, error) {
	var condominios []*models.Condominio
	for _, cond := range m.condominios {
		condominios = append(condominios, cond)
	}
	return condominios, nil
}

// TestAuthServiceLogin test login exitoso
func TestAuthServiceLogin(t *testing.T) {
	// Setup
	logger.Initialize(logger.Config{Level: "debug", Format: "text"})

	mockRepo := NewMockUserRepository()
	mockCondominioRepo := NewMockCondominioRepository()
	jwtConfig := testJWTConfig()
	jwtManager := utils.NewJWTManager(jwtConfig)
	authService := NewAuthService(mockRepo, mockCondominioRepo, jwtManager)

	// Agregar usuario de test
	hashedPIN, err := utils.HashPIN("123456")
	if err != nil {
		t.Fatalf("Error hashing PIN: %v", err)
	}

	testUser := &models.User{
		ID:           "user-001",
		Email:        "test@example.com",
		PIN:          hashedPIN,
		CondominioID: "cond-001",
		Rol:          "administrador",
		Estado:       "activo",
	}
	mockRepo.users[testUser.ID] = testUser

	// Agregar condominio de test
	testCondominio := &models.Condominio{
		ID:     "cond-001",
		Nombre: "Test Condominio",
		Ciudad: "Bogotá",
	}
	mockCondominioRepo.condominios[testCondominio.ID] = testCondominio

	// Test
	req := &dto.LoginRequest{
		Email: "test@example.com",
		PIN:   "123456",
	}

	response, err := authService.Login(context.Background(), req, "")

	// Verificaciones
	if err != nil {
		t.Errorf("Login failed: %v", err)
	}

	if response == nil {
		t.Error("Response is nil")
	}

	if response.Token == "" {
		t.Error("Token is empty")
	}

	if response.UserID != testUser.ID {
		t.Errorf("Expected UserID %s, got %s", testUser.ID, response.UserID)
	}

	if response.CondominioID != testUser.CondominioID {
		t.Errorf("Expected CondominioID %s, got %s", testUser.CondominioID, response.CondominioID)
	}

	if response.CondominioName != testCondominio.Nombre {
		t.Errorf("Expected CondominioName %s, got %s", testCondominio.Nombre, response.CondominioName)
	}

	if response.Role != testUser.Rol {
		t.Errorf("Expected role %s, got %s", testUser.Rol, response.Role)
	}

	if !response.IsAdmin {
		t.Error("Expected is_admin flag to be true for administrator role")
	}
}

// TestAuthServiceLoginInvalidPIN test login con PIN inválido
func TestAuthServiceLoginInvalidPIN(t *testing.T) {
	logger.Initialize(logger.Config{Level: "debug", Format: "text"})

	mockRepo := NewMockUserRepository()
	mockCondominioRepo := NewMockCondominioRepository()
	jwtConfig := testJWTConfig()
	jwtManager := utils.NewJWTManager(jwtConfig)
	authService := NewAuthService(mockRepo, mockCondominioRepo, jwtManager)

	hashedPIN, err := utils.HashPIN("123456")
	if err != nil {
		t.Fatalf("Error hashing PIN: %v", err)
	}

	testUser := &models.User{
		ID:           "user-001",
		Email:        "test@example.com",
		PIN:          hashedPIN,
		CondominioID: "cond-001",
		Rol:          "residente",
		Estado:       "activo",
	}
	mockRepo.users[testUser.ID] = testUser

	req := &dto.LoginRequest{
		Email: "test@example.com",
		PIN:   "999999", // PIN incorrecto
	}

	_, err = authService.Login(context.Background(), req, "")

	if err == nil {
		t.Error("Expected error for invalid PIN")
	}
}

// TestAuthServiceValidateToken test validación de token
func TestAuthServiceValidateToken(t *testing.T) {
	logger.Initialize(logger.Config{Level: "debug", Format: "text"})

	jwtConfig := testJWTConfig()
	jwtManager := utils.NewJWTManager(jwtConfig)

	// Generar token válido
	token, err := jwtManager.GenerateToken("user-001", "test@example.com", "cond-001", "residente", "")
	if err != nil {
		t.Errorf("Error generating token: %v", err)
	}

	// Validar token
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Errorf("Error validating token: %v", err)
	}

	if claims.UserID != "user-001" {
		t.Errorf("Expected UserID user-001, got %s", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected Email test@example.com, got %s", claims.Email)
	}
}

// Helpers

func testJWTConfig() utils.JWTConfig {
	return utils.JWTConfig{
		Secret:            "test_secret_key_minimum_32_characters_required_here",
		Expiration:        3600,
		RefreshSecret:     "test_refresh_secret_key_32_chars_minimum",
		RefreshExpiration: 604800,
	}
}
