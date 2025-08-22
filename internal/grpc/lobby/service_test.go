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
)

const (
	fixtureLobbyName = "Test Lobby"
	lobbyTypeName    = "*models.Lobby"
)

type LobbyServerTestSuite struct {
	suite.Suite
	lobbyRepo *MockLobbyRepository
	usrRepo   *MockUserRepository
	server    *LobbyService
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	return nil
}
func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	return nil, nil
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
func (m *MockLobbyRepository) ListAvailable() []*models.Lobby {
	args := m.Called()
	return args.Get(0).([]*models.Lobby)
}

func (s *LobbyServerTestSuite) SetupTest() {
	s.lobbyRepo = new(MockLobbyRepository)
	s.usrRepo = new(MockUserRepository)
	s.server = NewLobbyService(s.lobbyRepo, s.usrRepo)
}

func (s *LobbyServerTestSuite) TestCreateLobbySuccess() {
	mockUser := &models.User{Username: "testuser", Password: "testpass"}
	req := &lobby.CreateLobbyRequest{Name: fixtureLobbyName, Username: "testuser"}
	s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.lobbyRepo.On("Create", mock.AnythingOfType(lobbyTypeName)).Return(nil)
	resp, err := s.server.CreateLobby(context.Background(), req)
	s.NoError(err)
	s.Equal(fixtureLobbyName, resp.Name)
	s.Equal(string(models.LobbyStatusWaiting), resp.Status)
	s.Len(resp.Players, 1)
	s.Equal(mockUser.Username, resp.Players[0].Username)
	s.usrRepo.AssertExpectations(s.T())
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestCreateLobbyWhenUserNotFound() {
	s.usrRepo.On("FindByUsername", "unknownUser").Return(nil, usrrepo.ErrUserNotFound)
	req := &lobby.CreateLobbyRequest{Name: fixtureLobbyName, Username: "unknownUser"}
	resp, err := s.server.CreateLobby(context.Background(), req)
	s.Empty(resp)
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
	s.lobbyRepo.AssertNotCalled(s.T(), "Create", mock.Anything)
}

func (s *LobbyServerTestSuite) TestCreateLobbyWhenCreateReturnsError() {
	mockUser := &models.User{Username: "testuser"}
	s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.lobbyRepo.On("Create", mock.AnythingOfType(lobbyTypeName)).Return(errors.New("database error"))
	req := &lobby.CreateLobbyRequest{Name: fixtureLobbyName, Username: "testuser"}
	resp, err := s.server.CreateLobby(context.Background(), req)
	s.Empty(resp)
	s.Equal("database error", err.Error())
	s.usrRepo.AssertExpectations(s.T())
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestJoinLobbySuccess() {
	mockPlayer := &models.User{Username: "player2"}
	mockLobby := &models.Lobby{LobbyID: "1234", Name: fixtureLobbyName, Players: []models.User{{Username: "creator"}}}
	req := &lobby.JoinLobbyRequest{LobbyId: "1234", Username: "player2"}
	s.usrRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("AddPlayer", mockLobby, mockPlayer).Return(nil)
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusInProgress).Return(nil)
	resp, err := s.server.JoinLobby(context.Background(), req)
	s.NoError(err)
	s.Equal("1234", resp.LobbyId)
	s.Equal(string(models.LobbyStatusInProgress), resp.Status)
	s.Len(resp.Players, 2)
	s.lobbyRepo.AssertExpectations(s.T())
	s.usrRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestJoinLobbyWhenUserNotFound() {
	s.usrRepo.On("FindByUsername", "unknown").Return(nil, usrrepo.ErrUserNotFound)
	req := &lobby.JoinLobbyRequest{LobbyId: "1234", Username: "unknown"}

	resp, err := s.server.JoinLobby(context.Background(), req)

	s.ErrorIs(err, usrrepo.ErrUserNotFound)
	s.Empty(resp)
	s.lobbyRepo.AssertNotCalled(s.T(), "FindByID", mock.Anything)
}

func (s *LobbyServerTestSuite) TestJoinLobbyWhenLobbyNotFound() {
	mockPlayer := &models.User{Username: "player2"}
	s.usrRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "non-existent").Return(nil, lobbyrepo.ErrLobbyNotFound)
	req := &lobby.JoinLobbyRequest{LobbyId: "non-existent", Username: "player2"}

	resp, err := s.server.JoinLobby(context.Background(), req)

	s.ErrorIs(err, lobbyrepo.ErrLobbyNotFound)
	s.Empty(resp)
	s.lobbyRepo.AssertNotCalled(s.T(), "AddPlayer", mock.Anything, mock.Anything)
}

func (s *LobbyServerTestSuite) TestJoinLobbyWhenLobbyIsFull() {
	mockPlayer := &models.User{Username: "player3"}
	mockFullLobby := &models.Lobby{Players: []models.User{{}, {}}} // Lobby with 2 players
	s.usrRepo.On("FindByUsername", "player3").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "full-lobby").Return(mockFullLobby, nil)
	req := &lobby.JoinLobbyRequest{LobbyId: "full-lobby", Username: "player3"}

	resp, err := s.server.JoinLobby(context.Background(), req)

	s.Error(err)
	s.Equal("lobby is full", err.Error())
	s.Empty(resp)
	s.lobbyRepo.AssertNotCalled(s.T(), "AddPlayer", mock.Anything, mock.Anything)
}

func (s *LobbyServerTestSuite) TestJoinLobbyWhenAddPlayerFails() {
	mockPlayer := &models.User{Username: "player2"}
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{{}}}
	s.usrRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("AddPlayer", mockLobby, mockPlayer).Return(errors.New("db error"))
	req := &lobby.JoinLobbyRequest{LobbyId: "1234", Username: "player2"}

	resp, err := s.server.JoinLobby(context.Background(), req)

	s.Error(err)
	s.Equal("db error", err.Error())
	s.Empty(resp)
	s.lobbyRepo.AssertNotCalled(s.T(), "UpdateStatus", mock.Anything, mock.Anything)
}

func (s *LobbyServerTestSuite) TestJoinLobbyWhenUpdateStatusFails() {
	mockPlayer := &models.User{Username: "player2"}
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{{}}}
	s.usrRepo.On("FindByUsername", "player2").Return(mockPlayer, nil)
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("AddPlayer", mockLobby, mockPlayer).Return(nil)
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusInProgress).Return(errors.New("db error"))
	req := &lobby.JoinLobbyRequest{LobbyId: "1234", Username: "player2"}

	resp, err := s.server.JoinLobby(context.Background(), req)

	s.Error(err)
	s.Equal("db error", err.Error())
	s.Empty(resp)
	s.usrRepo.AssertExpectations(s.T())
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestFinishGameSuccess() {
	mockPlayer1 := models.User{Username: "player1"}
	mockPlayer1.ID = 1
	mockPlayer2 := models.User{Username: "player2"}
	mockPlayer2.ID = 2
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{mockPlayer1, mockPlayer2}}
	req := &lobby.FinishGameRequest{LobbyId: "1234"}

	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	// Since the winner is random, mock both possibilities
	s.lobbyRepo.On("UpdateWinner", mockLobby, uint(1)).Return(nil).Maybe()
	s.lobbyRepo.On("UpdateWinner", mockLobby, uint(2)).Return(nil).Maybe()
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusFinished).Return(nil)
	resp, err := s.server.FinishGame(context.Background(), req)
	s.NoError(err)
	s.Equal(string(models.LobbyStatusFinished), resp.Status)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestFinishGameWhenLobbyNotFound() {
	s.lobbyRepo.On("FindByID", "non-existent").Return(nil, lobbyrepo.ErrLobbyNotFound)
	req := &lobby.FinishGameRequest{LobbyId: "non-existent"}

	resp, err := s.server.FinishGame(context.Background(), req)

	s.ErrorIs(err, lobbyrepo.ErrLobbyNotFound)
	s.Empty(resp)
	s.lobbyRepo.AssertNotCalled(s.T(), "UpdateWinner", mock.Anything, mock.Anything)
}

func (s *LobbyServerTestSuite) TestFinishGameWhenUpdateWinnerFails() {
	mockPlayer := models.User{Username: "player1"}
	mockPlayer.ID = 1
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{mockPlayer}}
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("UpdateWinner", mockLobby, uint(1)).Return(errors.New("db error"))
	req := &lobby.FinishGameRequest{LobbyId: "1234"}

	resp, err := s.server.FinishGame(context.Background(), req)

	s.Error(err)
	s.Equal("db error", err.Error())
	s.Empty(resp)
	s.lobbyRepo.AssertNotCalled(s.T(), "UpdateStatus", mock.Anything, mock.Anything)
}

func (s *LobbyServerTestSuite) TestFinishGameWhenUpdateStatusFails() {
	mockPlayer := models.User{Username: "player1"}
	mockPlayer.ID = 1
	mockLobby := &models.Lobby{LobbyID: "1234", Players: []models.User{mockPlayer}}
	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	s.lobbyRepo.On("UpdateWinner", mockLobby, uint(1)).Return(nil)
	s.lobbyRepo.On("UpdateStatus", mockLobby, models.LobbyStatusFinished).Return(errors.New("db error"))
	req := &lobby.FinishGameRequest{LobbyId: "1234"}

	resp, err := s.server.FinishGame(context.Background(), req)

	s.Error(err)
	s.Equal("db error", err.Error())
	s.Empty(resp)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestGetLobbySuccess() {
	// Arrange
	player1 := models.User{Username: "creator"}
	player1.ID = 1
	player2 := models.User{Username: "opponent"}
	player2.ID = 2
	mockLobby := &models.Lobby{
		LobbyID:  "1234",
		Name:     fixtureLobbyName,
		Players:  []models.User{player1, player2},
		Status:   models.LobbyStatusInProgress,
		WinnerID: &player1.ID,
		Winner:   &player1,
	}

	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	req := &lobby.GetLobbyRequest{LobbyId: "1234"}
	resp, err := s.server.GetLobby(context.Background(), req)

	s.NoError(err)
	s.Equal(mockLobby.LobbyID, resp.LobbyId)
	s.Equal(mockLobby.Name, resp.Name)
	s.Len(resp.Players, 2)
	s.Equal(string(mockLobby.Status), resp.Status)
	s.Equal(uint32(*mockLobby.WinnerID), *resp.WinnerId)
	s.Equal(player1.Username, *resp.WinnerUsername)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestGetLobbyWhenLobbyIsNotFound() {
	s.lobbyRepo.On("FindByID", "non-existent").Return(nil, lobbyrepo.ErrLobbyNotFound)
	req := &lobby.GetLobbyRequest{LobbyId: "non-existent"}
	resp, err := s.server.GetLobby(context.Background(), req)
	s.ErrorIs(err, lobbyrepo.ErrLobbyNotFound)
	s.Empty(resp)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestListAvailableLobbies() {
	mockLobbies := []*models.Lobby{
		{LobbyID: "1", Name: "Lobby 1", Status: models.LobbyStatusWaiting},
		{LobbyID: "2", Name: "Lobby 2", Status: models.LobbyStatusWaiting},
	}
	s.lobbyRepo.On("ListAvailable").Return(mockLobbies)

	resp, err := s.server.ListAvailableLobbies(context.Background(), &lobby.ListAvailableLobbiesRequest{})

	s.NoError(err)
	s.Len(resp.Lobbies, 2)
	s.Equal(mockLobbies[0].LobbyID, resp.Lobbies[0].LobbyId)
	s.Equal(mockLobbies[1].LobbyID, resp.Lobbies[1].LobbyId)
	s.lobbyRepo.AssertExpectations(s.T())
}

func TestLobbyServer(t *testing.T) {
	suite.Run(t, new(LobbyServerTestSuite))
}
