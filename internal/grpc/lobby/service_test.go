package lobby

import (
	"context"
	"errors"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fixtureLobbyName = "Test Lobby"
	fixtureLobbyID   = "lobby-123"
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

func (s *LobbyServiceTestSuite) TestCreateLobbySuccess() {
	// Arrange
	mockUser := &models.User{Username: "testuser"}
	req := &lobby.CreateLobbyRequest{Name: fixtureLobbyName, Username: "testuser"}
	s.userRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.lobbyRepo.On("Create", mock.AnythingOfType("*models.Lobby")).Return(nil)

	// Act
	resp, err := s.service.CreateLobby(context.Background(), req)

	// Assert
	s.NoError(err)
	s.Equal(fixtureLobbyName, resp.Name)
	s.Len(resp.Players, 1)
	s.Equal("testuser", resp.Players[0].Username)
	s.lobbyRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
}

func (s *LobbyServiceTestSuite) TestCreateLobbyFailsWithEmptyName() {
	req := &lobby.CreateLobbyRequest{Name: "   ", Username: "testuser"} // Whitespace name

	_, err := s.service.CreateLobby(context.Background(), req)

	s.assertGrpcError(err, codes.InvalidArgument, "lobby name cannot be empty")
	s.userRepo.AssertNotCalled(s.T(), "FindByUsername", mock.Anything)
	s.lobbyRepo.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func (s *LobbyServiceTestSuite) TestCreateLobbyFailsWhenUserNotFound() {
	req := &lobby.CreateLobbyRequest{Name: fixtureLobbyName, Username: "unknownUser"}
	s.userRepo.On("FindByUsername", "unknownUser").Return(nil, usrrepo.ErrUserNotFound)

	_, err := s.service.CreateLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Invalid creator")
	s.lobbyRepo.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func (s *LobbyServiceTestSuite) TestCreateLobbyFailsWhenRepoCreateFails() {
	mockUser := &models.User{Username: "testuser"}
	req := &lobby.CreateLobbyRequest{Name: fixtureLobbyName, Username: "testuser"}
	dbError := errors.New("database connection failed")
	s.userRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.lobbyRepo.On("Create", mock.AnythingOfType("*models.Lobby")).Return(dbError)

	_, err := s.service.CreateLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Lobby DB error")
	s.lobbyRepo.AssertExpectations(s.T())
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

func (s *LobbyServiceTestSuite) TestJoinLobbyFailsWhenUserNotFound() {
	req := &lobby.JoinLobbyRequest{LobbyId: fixtureLobbyID, Username: "unknownUser"}
	s.userRepo.On("FindByUsername", "unknownUser").Return(nil, usrrepo.ErrUserNotFound)

	_, err := s.service.JoinLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Invalid player")
	s.lobbyRepo.AssertNotCalled(s.T(), "FindByID", mock.Anything)
}

func (s *LobbyServiceTestSuite) TestJoinLobbyFailsWhenLobbyNotFound() {
	// Arrange
	mockPlayer := &models.User{Username: "player2"}
	req := &lobby.JoinLobbyRequest{LobbyId: "non-existent-lobby", Username: "player2"}

	s.userRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "non-existent-lobby").Return(nil, lobbyrepo.ErrLobbyNotFound)

	_, err := s.service.JoinLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Lobby not found")
	s.lobbyRepo.AssertNotCalled(s.T(), "AddPlayer", mock.Anything, mock.Anything)
}

func (s *LobbyServiceTestSuite) TestJoinLobbyFailsOnUpdateStatus() {
	mockPlayer := &models.User{Username: "player2"}
	mockLobby := &models.Lobby{LobbyID: fixtureLobbyID, Players: []models.User{{Username: "creator"}}}
	req := &lobby.JoinLobbyRequest{LobbyId: fixtureLobbyID, Username: "player2"}
	dbError := errors.New("status update failed")

	s.userRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", fixtureLobbyID).Return(mockLobby, nil)
	s.lobbyRepo.On("AddPlayer", mockLobby, mockPlayer).Return(nil) // This call succeeds
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusInProgress).Return(dbError)

	_, err := s.service.JoinLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Lobby DB error")
	s.lobbyRepo.AssertExpectations(s.T()) // Verify all expected calls were made
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

func (s *LobbyServiceTestSuite) TestFinishGameSuccess() {
	mockPlayer1 := models.User{Username: "player1"}
	mockPlayer1.ID = 1
	mockLobby := &models.Lobby{
		LobbyID: fixtureLobbyID,
		Players: []models.User{mockPlayer1}, // Lobby with one player
	}
	req := &lobby.FinishGameRequest{LobbyId: fixtureLobbyID}

	s.lobbyRepo.On("FindByID", fixtureLobbyID).Return(mockLobby, nil)
	s.lobbyRepo.On("UpdateWinner", mockLobby, mockPlayer1.ID).Return(nil)
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusFinished).Return(nil)

	resp, err := s.service.FinishGame(context.Background(), req)

	s.NoError(err)
	s.Equal(string(models.LobbyStatusFinished), resp.Status)
	s.NotNil(resp.WinnerId)
	s.Equal(uint32(mockPlayer1.ID), *resp.WinnerId)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServiceTestSuite) TestFinishGameFailsWhenLobbyNotFound() {
	req := &lobby.FinishGameRequest{LobbyId: "non-existent"}
	s.lobbyRepo.On("FindByID", "non-existent").Return(nil, lobbyrepo.ErrLobbyNotFound)

	_, err := s.service.FinishGame(context.Background(), req)

	s.ErrorIs(err, lobbyrepo.ErrLobbyNotFound)
	s.lobbyRepo.AssertNotCalled(s.T(), "UpdateWinner", mock.Anything, mock.Anything)
}

func (s *LobbyServiceTestSuite) TestFinishGameFailsOnUpdateWinner() {
	mockPlayer1 := models.User{Username: "player1"}
	mockPlayer1.ID = 1
	mockLobby := &models.Lobby{LobbyID: fixtureLobbyID, Players: []models.User{mockPlayer1}}
	req := &lobby.FinishGameRequest{LobbyId: fixtureLobbyID}
	dbError := errors.New("db write failed")

	s.lobbyRepo.On("FindByID", fixtureLobbyID).Return(mockLobby, nil)
	s.lobbyRepo.On("UpdateWinner", mockLobby, mockPlayer1.ID).Return(dbError)

	_, err := s.service.FinishGame(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Lobby DB error")
	s.lobbyRepo.AssertNotCalled(s.T(), "UpdateStatus", mock.Anything, mock.Anything)
}

func (s *LobbyServiceTestSuite) TestFinishGameFailsOnUpdateStatus() {
	mockPlayer1 := models.User{Username: "player1"}
	mockPlayer1.ID = 1
	mockLobby := &models.Lobby{LobbyID: fixtureLobbyID, Players: []models.User{mockPlayer1}}
	req := &lobby.FinishGameRequest{LobbyId: fixtureLobbyID}
	dbError := errors.New("db status update failed")

	s.lobbyRepo.On("FindByID", fixtureLobbyID).Return(mockLobby, nil)
	s.lobbyRepo.On("UpdateWinner", mockLobby, mockPlayer1.ID).Return(nil)
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusFinished).Return(dbError)

	_, err := s.service.FinishGame(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Lobby DB error")
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServiceTestSuite) TestGetLobbySuccess() {
	// Arrange
	mockLobby := &models.Lobby{
		LobbyID: fixtureLobbyID,
		Name:    fixtureLobbyName,
		Players: []models.User{{Username: "player1"}},
	}
	req := &lobby.GetLobbyRequest{LobbyId: fixtureLobbyID}
	s.lobbyRepo.On("FindByID", fixtureLobbyID).Return(mockLobby, nil)

	resp, err := s.service.GetLobby(context.Background(), req)

	s.NoError(err)
	s.Equal(fixtureLobbyID, resp.LobbyId)
	s.Equal(mockLobby.Name, resp.Name)
	s.Len(resp.Players, 1)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServiceTestSuite) TestGetLobbyFailsWhenNotFound() {
	req := &lobby.GetLobbyRequest{LobbyId: "non-existent"}
	s.lobbyRepo.On("FindByID", "non-existent").Return(nil, lobbyrepo.ErrLobbyNotFound)

	_, err := s.service.GetLobby(context.Background(), req)

	s.assertGrpcError(err, codes.Internal, "Invalid Lobby ID")
}

func (s *LobbyServiceTestSuite) TestListAvailableLobbiesSuccess() {
	mockLobbies := []*models.Lobby{
		{LobbyID: "lobby-1", Name: "First Lobby"},
		{LobbyID: "lobby-2", Name: "Second Lobby"},
	}
	s.lobbyRepo.On("ListAvailable").Return(mockLobbies)
	resp, err := s.service.ListAvailableLobbies(context.Background(), &lobby.ListAvailableLobbiesRequest{})

	s.NoError(err)
	s.Len(resp.Lobbies, 2)
	s.Equal("lobby-1", resp.Lobbies[0].LobbyId)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServiceTestSuite) TestListAvailableLobbiesSuccessWhenEmpty() {
	mockLobbies := []*models.Lobby{} // Return an empty slice
	s.lobbyRepo.On("ListAvailable").Return(mockLobbies)

	resp, err := s.service.ListAvailableLobbies(context.Background(), &lobby.ListAvailableLobbiesRequest{})

	s.NoError(err)
	s.Len(resp.Lobbies, 0)
	s.lobbyRepo.AssertExpectations(s.T())
}

func TestLobbyService(t *testing.T) {
	suite.Run(t, new(LobbyServiceTestSuite))
}
