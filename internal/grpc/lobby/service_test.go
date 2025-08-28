package lobby

import (
	"context"
	"errors"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

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

func (m *MockLobbyRepository) ListAvailable() []*models.Lobby {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*models.Lobby)
}

func (m *MockLobbyRepository) AddPlayer(lobby *models.Lobby, player *models.User) error {
	args := m.Called(lobby, player)
	return args.Error(0)
}

func (m *MockLobbyRepository) UpdateStatus(lobby *models.Lobby, status models.LobbyStatus) error {
	args := m.Called(lobby, status)
	return args.Error(0)
}

func (m *MockLobbyRepository) UpdateWinner(lobby *models.Lobby, winnerID uint) error {
	args := m.Called(lobby, winnerID)
	return args.Error(0)
}

func (m *MockLobbyRepository) Delete(lobbyID string) error {
	args := m.Called(lobbyID)
	return args.Error(0)
}

type LobbyServiceTestSuite struct {
	suite.Suite
	lobbyRepo *MockLobbyRepository
	userRepo  *MockUserRepository
	service   lobby.LobbyServiceServer
}

func (s *LobbyServiceTestSuite) SetupTest() {
	s.lobbyRepo = new(MockLobbyRepository)
	s.userRepo = new(MockUserRepository)
	s.service = NewLobbyService(s.lobbyRepo, s.userRepo)
}

// Helper to assert on gRPC errors cleanly
func (s *LobbyServiceTestSuite) assertGrpcError(err error, code codes.Code, msgContains string) {
	s.Error(err, "Expected an error")
	st, ok := status.FromError(err)
	s.True(ok, "Error should be a gRPC status error")
	s.Equal(code, st.Code())
	if msgContains != "" {
		s.Contains(st.Message(), msgContains)
	}
}

func (s *LobbyServiceTestSuite) TestJoinLobbySuccess() {
	mockPlayer := &models.User{Username: "player2"}
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{{Username: "creator"}}}
	req := &lobby.JoinLobbyRequest{LobbyId: "1234", Username: "player2"}

	s.userRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("AddPlayer", mockLobby, mockPlayer).Return(nil)
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusInProgress).Return(nil)

	resp, err := s.service.JoinLobby(context.Background(), req)

	s.NoError(err)
	s.Len(resp.Players, 2)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServiceTestSuite) TestJoinLobbyWhenLobbyIsFull() {
	mockPlayer := &models.User{Username: "player3"}
	mockFullLobby := &models.Lobby{Players: []models.User{{}, {}}} // Lobby with 2 players
	req := &lobby.JoinLobbyRequest{LobbyId: "full-lobby", Username: "player3"}

	s.userRepo.On("FindByUsername", "player3").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "full-lobby").Return(mockFullLobby, nil)

	_, err := s.service.JoinLobby(context.Background(), req)

	s.assertGrpcError(err, codes.FailedPrecondition, "lobby is full")
	s.lobbyRepo.AssertNotCalled(s.T(), "AddPlayer", mock.Anything, mock.Anything)
}

func (s *LobbyServiceTestSuite) TestJoinLobbyWhenAddPlayerFails() {
	mockPlayer := &models.User{Username: "player2"}
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{{}}}
	req := &lobby.JoinLobbyRequest{LobbyId: "1234", Username: "player2"}
	dbErr := errors.New("db error")

	s.userRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("AddPlayer", mockLobby, mockPlayer).Return(dbErr)

	_, err := s.service.JoinLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "db error")
	s.lobbyRepo.AssertNotCalled(s.T(), "UpdateStatus", mock.Anything, mock.Anything)
}

func TestLobbyService(t *testing.T) {
	suite.Run(t, new(LobbyServiceTestSuite))
}
