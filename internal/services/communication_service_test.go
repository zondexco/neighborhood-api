package services

import (
"context"
"errors"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"

"neighborhood-api/internal/models"
)

// MockCommunicationRepository es un mock del repositorio
type MockCommunicationRepository struct {
mock.Mock
}

func (m *MockCommunicationRepository) Create(ctx context.Context, communication *models.Communication) error {
args := m.Called(ctx, communication)
return args.Error(0)
}

func (m *MockCommunicationRepository) FindByID(ctx context.Context, communicationID, condominioID string) (*models.Communication, error) {
args := m.Called(ctx, communicationID, condominioID)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*models.Communication), args.Error(1)
}

func (m *MockCommunicationRepository) GetByCondominio(ctx context.Context, condominioID string, page, pageSize int) ([]*models.Communication, int, error) {
args := m.Called(ctx, condominioID, page, pageSize)
if args.Get(0) == nil {
return nil, args.Int(1), args.Error(2)
}
return args.Get(0).([]*models.Communication), args.Int(1), args.Error(2)
}

func (m *MockCommunicationRepository) Update(ctx context.Context, communication *models.Communication) error {
args := m.Called(ctx, communication)
return args.Error(0)
}

func (m *MockCommunicationRepository) Delete(ctx context.Context, communicationID, condominioID string) error {
args := m.Called(ctx, communicationID, condominioID)
return args.Error(0)
}

func (m *MockCommunicationRepository) ListVisible(ctx context.Context, condominioID, userID, userRole string, page, pageSize int) ([]*models.Communication, int, error) {
args := m.Called(ctx, condominioID, userID, userRole, page, pageSize)
if args.Get(0) == nil {
return nil, args.Int(1), args.Error(2)
}
return args.Get(0).([]*models.Communication), args.Int(1), args.Error(2)
}

func (m *MockCommunicationRepository) MarkRead(ctx context.Context, communicationID, userID string) error {
args := m.Called(ctx, communicationID, userID)
return args.Error(0)
}

func (m *MockCommunicationRepository) UnreadCount(ctx context.Context, condominioID, userID, userRole string) (int, error) {
args := m.Called(ctx, condominioID, userID, userRole)
return args.Int(0), args.Error(1)
}

func (m *MockCommunicationRepository) ListComments(ctx context.Context, communicationID string) ([]*models.ComunicadoComentario, error) {
args := m.Called(ctx, communicationID)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).([]*models.ComunicadoComentario), args.Error(1)
}

func (m *MockCommunicationRepository) CreateComment(ctx context.Context, comment *models.ComunicadoComentario) error {
args := m.Called(ctx, comment)
return args.Error(0)
}

// Tests
func TestCommunicationService_Create_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

request := CreateCommunicationRequest{
Titulo:    "Mantenimiento",
Contenido: "Se realizará mantenimiento el próximo lunes",
}

mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(comm *models.Communication) bool {
return comm.Titulo == "Mantenimiento" &&
comm.Contenido == "Se realizará mantenimiento el próximo lunes" &&
comm.CondominioID == "cond-001" &&
comm.Autor == "user-001"
})).Run(func(args mock.Arguments) {
comm := args.Get(1).(*models.Communication)
comm.ID = "comm-001"
comm.CreatedAt = time.Now()
comm.UpdatedAt = time.Now()
}).Return(nil)

result, err := service.Create(context.Background(), request, "cond-001", "user-001")

assert.NoError(t, err)
assert.NotNil(t, result)
assert.Equal(t, "comm-001", result.ID)
assert.Equal(t, "Mantenimiento", result.Titulo)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_Create_MissingTitulo(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

request := CreateCommunicationRequest{
Titulo:    "",
Contenido: "Contenido válido",
}

result, err := service.Create(context.Background(), request, "cond-001", "user-001")

assert.Error(t, err)
assert.Nil(t, result)
assert.Equal(t, "titulo is required", err.Error())
mockRepo.AssertNotCalled(t, "Create")
}

func TestCommunicationService_Create_MissingContenido(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

request := CreateCommunicationRequest{
Titulo:    "Título válido",
Contenido: "",
}

result, err := service.Create(context.Background(), request, "cond-001", "user-001")

assert.Error(t, err)
assert.Nil(t, result)
assert.Equal(t, "contenido is required", err.Error())
mockRepo.AssertNotCalled(t, "Create")
}

func TestCommunicationService_GetByID_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

comm := &models.Communication{
ID:           "comm-001",
Titulo:       "Aviso importante",
Contenido:    "Por favor leer",
CondominioID: "cond-001",
CreatedAt:    time.Now(),
UpdatedAt:    time.Now(),
}

mockRepo.On("FindByID", mock.Anything, "comm-001", "cond-001").Return(comm, nil)

result, err := service.GetByID(context.Background(), "comm-001", "cond-001")

assert.NoError(t, err)
assert.NotNil(t, result)
assert.Equal(t, "comm-001", result.ID)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_GetByID_NotFound(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

mockRepo.On("FindByID", mock.Anything, "comm-999", "cond-001").Return(nil, errors.New("communication not found"))

result, err := service.GetByID(context.Background(), "comm-999", "cond-001")

assert.Error(t, err)
assert.Nil(t, result)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_List_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

comms := []*models.Communication{
{
ID:     "comm-001",
Titulo: "Aviso 1",
},
{
ID:     "comm-002",
Titulo: "Aviso 2",
},
}

mockRepo.On("GetByCondominio", mock.Anything, "cond-001", 1, 10).Return(comms, 2, nil)

result, total, err := service.List(context.Background(), "cond-001", 1, 10)

assert.NoError(t, err)
assert.Len(t, result, 2)
assert.Equal(t, 2, total)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_Update_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

oldComm := &models.Communication{
ID:           "comm-001",
Titulo:       "Título anterior",
Contenido:    "Contenido anterior",
CondominioID: "cond-001",
}

newTitulo := "Título actualizado"
request := UpdateCommunicationRequest{
Titulo: &newTitulo,
}

mockRepo.On("FindByID", mock.Anything, "comm-001", "cond-001").Return(oldComm, nil)
mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(comm *models.Communication) bool {
return comm.Titulo == "Título actualizado" &&
comm.Contenido == "Contenido anterior"
})).Return(nil)

result, err := service.Update(context.Background(), request, "comm-001", "cond-001")

assert.NoError(t, err)
assert.NotNil(t, result)
assert.Equal(t, "Título actualizado", result.Titulo)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_Delete_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

mockRepo.On("Delete", mock.Anything, "comm-001", "cond-001").Return(nil)

err := service.Delete(context.Background(), "comm-001", "cond-001")

assert.NoError(t, err)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_Delete_NotFound(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

mockRepo.On("Delete", mock.Anything, "comm-999", "cond-001").Return(errors.New("communication not found"))

err := service.Delete(context.Background(), "comm-999", "cond-001")

assert.Error(t, err)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_ListVisible_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

comms := []*models.Communication{
{ID: "comm-001", Titulo: "Visible 1", Publicado: true},
{ID: "comm-002", Titulo: "Visible 2", Publicado: true},
}

mockRepo.On("ListVisible", mock.Anything, "cond-001", "user-001", "residente", 1, 10).Return(comms, 2, nil)

result, total, err := service.ListVisible(context.Background(), "cond-001", "user-001", "residente", 1, 10)

assert.NoError(t, err)
assert.Len(t, result, 2)
assert.Equal(t, 2, total)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_MarkRead_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

mockRepo.On("MarkRead", mock.Anything, "comm-001", "user-001").Return(nil)

err := service.MarkRead(context.Background(), "comm-001", "user-001")

assert.NoError(t, err)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_UnreadCount_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

mockRepo.On("UnreadCount", mock.Anything, "cond-001", "user-001", "residente").Return(5, nil)

count, err := service.UnreadCount(context.Background(), "cond-001", "user-001", "residente")

assert.NoError(t, err)
assert.Equal(t, 5, count)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_ListComments_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

comments := []*models.ComunicadoComentario{
{ID: "comment-001", Contenido: "Buen comunicado"},
{ID: "comment-002", Contenido: "Gracias por informar"},
}

mockRepo.On("ListComments", mock.Anything, "comm-001").Return(comments, nil)

result, err := service.ListComments(context.Background(), "comm-001")

assert.NoError(t, err)
assert.Len(t, result, 2)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_CreateComment_Success(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

mockRepo.On("CreateComment", mock.Anything, mock.MatchedBy(func(c *models.ComunicadoComentario) bool {
return c.IDComunicado == "comm-001" && c.IDUsuario == "user-001" && c.Contenido == "Buen aviso"
})).Return(nil)

result, err := service.CreateComment(context.Background(), "comm-001", "user-001", "Buen aviso")

assert.NoError(t, err)
assert.NotNil(t, result)
assert.Equal(t, "Buen aviso", result.Contenido)
mockRepo.AssertExpectations(t)
}

func TestCommunicationService_CreateComment_EmptyContent(t *testing.T) {
mockRepo := new(MockCommunicationRepository)
service := NewCommunicationService(mockRepo)

result, err := service.CreateComment(context.Background(), "comm-001", "user-001", "")

assert.Error(t, err)
assert.Nil(t, result)
assert.Equal(t, "contenido is required", err.Error())
mockRepo.AssertNotCalled(t, "CreateComment")
}
