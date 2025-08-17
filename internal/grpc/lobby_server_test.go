package grpc

import (
	"context"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockLobbyRepository is a mock for the LobbyRepository.
type MockLobbyRepository struct {
	mock.Mock
}

func (m *MockLobbyRepository) Create(lobby *models.Lobby) error {
	args := m.Called(lobby)
	return args.Error(0)
}

func (m *MockLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	args := m.Called(lobbyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lobby), args.Error(1)
}

func (m *MockLobbyRepository) ListAvailable() []models.Lobby {
	args := m.Called()
	return args.Get(0).([]models.Lobby)
}

func (m *MockLobbyRepository) UpdateLobbyOpponentAndStatus(lobbyID string, opponentID uint, status models.LobbyStatus) error {
	args := m.Called(lobbyID, opponentID, status)
	return args.Error(0)
}

func TestCreateLobby(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockLobbyRepo := new(MockLobbyRepository)

	server := NewLobbyServer(mockLobbyRepo, mockUserRepo)

	mockUser := &models.User{Username: "testuser", Password: "testpass"}
	mockUserRepo.On("FindByUsername", "testuser").Return(mockUser, nil)

	mockLobby := &models.Lobby{
		LobbyID:   "1234",
		Name:      "Test Lobby",
		CreatorID: mockUser.ID,
		Status:    models.LobbyStatusWaiting,
		Creator:   *mockUser,
	}
	mockLobbyRepo.On("Create", mock.AnythingOfType("*models.Lobby")).Return(nil)
	mockLobbyRepo.On("FindByID", mock.AnythingOfType("string")).Return(mockLobby, nil)

	req := &lobby.CreateLobbyRequest{
		Name:     "Test Lobby",
		Username: "testuser",
	}

	resp, err := server.CreateLobby(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, mockLobby.LobbyID, resp.LobbyId)
	assert.Equal(t, mockLobby.Name, resp.Name)
	assert.Equal(t, uint32(mockLobby.CreatorID), resp.CreatorId)
	assert.Equal(t, string(mockLobby.Status), resp.Status)
	assert.Equal(t, mockUser.Username, resp.CreatorUsername)

	mockUserRepo.AssertExpectations(t)
	mockLobbyRepo.AssertExpectations(t)
}
