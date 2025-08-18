package grpc

import (
	"context"
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

type LobbyServerTestSuite struct {
	suite.Suite
	lobbyRepo *MockLobbyRepository
	usrRepo   *MockUserRepository
	server    *LobbyServer
}

func (s *LobbyServerTestSuite) SetupTest() {
	s.lobbyRepo = new(MockLobbyRepository)
	s.usrRepo = new(MockUserRepository)
	s.server = NewLobbyServer(s.lobbyRepo, s.usrRepo)
}

func (s *LobbyServerTestSuite) TestCreateLobbySuccess() {
	mockUser := &models.User{Username: "testuser", Password: "testpass"}
	mockLobby := &models.Lobby{
		LobbyID:   "1234",
		Name:      "Test Lobby",
		CreatorID: mockUser.ID,
		Status:    models.LobbyStatusWaiting,
		Creator:   *mockUser,
	}

	mock.InOrder(
		s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil),
		s.lobbyRepo.On("Create", mock.AnythingOfType("*models.Lobby")).Return(nil),
		s.lobbyRepo.On("FindByID", mock.AnythingOfType("string")).Return(mockLobby, nil),
	)

	req := &lobby.CreateLobbyRequest{
		Name:     "Test Lobby",
		Username: "testuser",
	}

	resp, err := s.server.CreateLobby(context.Background(), req)
	s.NoError(err)
	s.Equal(mockLobby.LobbyID, resp.LobbyId)
	s.Equal(mockLobby.Name, resp.Name)
	s.Equal(uint32(mockLobby.CreatorID), resp.CreatorId)
	s.Equal(string(mockLobby.Status), resp.Status)
	s.Equal(mockUser.Username, resp.CreatorUsername)
	s.usrRepo.AssertExpectations(s.T())
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestCreateLobbyWhenCreatorIsNotFound() {
	s.usrRepo.On("FindByUsername", "unknown").Return(nil, usrrepo.ErrUserNotFound)
	req := &lobby.CreateLobbyRequest{
		Name:     "Test Lobby",
		Username: "unknown",
	}
	_, err := s.server.CreateLobby(context.Background(), req)
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
	s.usrRepo.AssertExpectations(s.T())
	s.lobbyRepo.AssertNotCalled(s.T(), "Create", mock.AnythingOfType("*models.Lobby"))
	s.lobbyRepo.AssertNotCalled(s.T(), "FindByID", mock.AnythingOfType("string"))
}

func (s *LobbyServerTestSuite) TestCreateLobbyWhenLobbyAlreadyExists() {
	mockUser := &models.User{Username: "testuser", Password: "testpass"}
	s.usrRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	s.lobbyRepo.On("Create", mock.AnythingOfType("*models.Lobby")).Return(lobbyrepo.ErrLobbyExists)

	req := &lobby.CreateLobbyRequest{
		Name:     "Test Lobby",
		Username: "testuser",
	}

	_, err := s.server.CreateLobby(context.Background(), req)
	s.ErrorIs(err, lobbyrepo.ErrLobbyExists)
	s.usrRepo.AssertExpectations(s.T())
	s.lobbyRepo.AssertCalled(s.T(), "Create", mock.AnythingOfType("*models.Lobby"))
	s.lobbyRepo.AssertNotCalled(s.T(), "FindByID", mock.AnythingOfType("string"))
}

func (s *LobbyServerTestSuite) TestGetLobbySuccess() {
	mockLobby := &models.Lobby{
		LobbyID:   "1234",
		Name:      "Test Lobby",
		CreatorID: 1,
		Status:    models.LobbyStatusWaiting,
	}

	s.lobbyRepo.On("FindByID", "1234").Return(mockLobby, nil)
	req := &lobby.GetLobbyRequest{LobbyId: "1234"}
	resp, err := s.server.GetLobby(context.Background(), req)
	s.NoError(err)
	s.Equal(mockLobby.LobbyID, resp.LobbyId)
	s.Equal(mockLobby.Name, resp.Name)
	s.Equal(uint32(mockLobby.CreatorID), resp.CreatorId)
	s.Equal(string(mockLobby.Status), resp.Status)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestGetLobbyWhenLobbyIsNotPresent() {
	s.lobbyRepo.On("FindByID", "notfound").Return(nil, lobbyrepo.ErrLobbyNotFound)
	req := &lobby.GetLobbyRequest{LobbyId: "notfound"}
	_, err := s.server.GetLobby(context.Background(), req)
	s.ErrorIs(err, lobbyrepo.ErrLobbyNotFound)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestListAvailableLobbiesWhenThereAreWaitingLobbies() {
	mockLobbies := []models.Lobby{
		{LobbyID: "1", Name: "Lobby 1", CreatorID: 1, Status: models.LobbyStatusWaiting},
		{LobbyID: "2", Name: "Lobby 2", CreatorID: 2, Status: models.LobbyStatusWaiting},
	}
	s.lobbyRepo.On("ListAvailable").Return(mockLobbies)
	resp, err := s.server.ListAvailableLobbies(context.Background(), &lobby.ListAvailableLobbiesRequest{})
	s.NoError(err)
	s.Len(resp.Lobbies, 2)
	s.Equal(mockLobbies[0].LobbyID, resp.Lobbies[0].LobbyId)
	s.Equal(mockLobbies[1].LobbyID, resp.Lobbies[1].LobbyId)
	s.lobbyRepo.AssertExpectations(s.T())
}

func (s *LobbyServerTestSuite) TestListAvailableLobbiesWhenThereAreNotWaitingLobbies() {
	s.lobbyRepo.On("ListAvailable").Return([]models.Lobby{})
	resp, err := s.server.ListAvailableLobbies(context.Background(), &lobby.ListAvailableLobbiesRequest{})
	s.NoError(err)
	s.Empty(resp.Lobbies)
	s.lobbyRepo.AssertExpectations(s.T())
}

func TestLobbyServer(t *testing.T) {
	suite.Run(t, new(LobbyServerTestSuite))
}
