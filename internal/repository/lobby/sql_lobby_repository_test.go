package lobbyrepo

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	fixtureLobbyName    string = "Test Lobby"
	fixtureUserUsername string = "test"
	fixtureUserPassword string = "password123"
)

type LobbySQLRepositoryTestSuite struct {
	suite.Suite
	db        *gorm.DB
	lobbyRepo LobbyRepository
	usrRepo   *UserTestRepository
}

type UserTestRepository struct {
	mock.Mock
}

func (r *UserTestRepository) Save(user *models.User) error {
	args := r.Called(user)
	return args.Error(0)
}

func (r *UserTestRepository) FindByUsername(username string) (*models.User, error) {
	args := r.Called(username)
	if user, ok := args.Get(0).(*models.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (r *UserTestRepository) FindByID(id uint) (*models.User, error) {
	args := r.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (s *LobbySQLRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.T().Fatalf("Failed to connect to database at the suite start: %v", err)
	}
	s.db = db
}

func (s *LobbySQLRepositoryTestSuite) TearDownSuite() {
	db, _ := s.db.DB()
	if err := db.Close(); err != nil {
		s.T().Fatalf("Failed to close database connection at the suite end: %v", err)
	}
}

func (s *LobbySQLRepositoryTestSuite) SetupTest() {
	if err := s.db.Migrator().DropTable(&models.User{}, &models.Lobby{}); err != nil {
		s.T().Fatalf("Failed to drop User and Lobby tables before test run: %v", err)
	}

	if err := s.db.AutoMigrate(&models.User{}, &models.Lobby{}); err != nil {
		s.T().Fatalf("Failed to migrate User and Lobby tables before test run: %v", err)
	}

	s.usrRepo = new(UserTestRepository)
	s.lobbyRepo = NewSQLLobbyRepository(s.db, s.usrRepo)
}

func (s *LobbySQLRepositoryTestSuite) TestCreateLobbyWhenThereIsNotAlreadyTheLobby() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.usrRepo.On("FindByID", lobby.CreatorID).Return(&models.User{}, nil)
	err := s.lobbyRepo.Create(lobby)
	assert.Nil(s.T(), err)
}

func (s *LobbySQLRepositoryTestSuite) createTestUser(username, password string) models.User {
	user := models.User{Username: username, Password: password}
	s.db.Create(&user)
	return user
}

func (s *LobbySQLRepositoryTestSuite) TestCreateLobbyWhenUserIsNotAlreadyPresentShouldReturnUserNotFoundError() {
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: 1,
	}

	s.usrRepo.On("FindByID", lobby.CreatorID).Return(&models.User{}, usrrepo.ErrUserNotFound)
	err := s.lobbyRepo.Create(lobby)
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
}

func (s *LobbySQLRepositoryTestSuite) TestCreateLobbyWhenLobbyIsAlreadySavedShouldReturnLobbyExistsError() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(lobby)
	s.usrRepo.On("FindByID", lobby.CreatorID).Return(&models.User{}, nil)
	err := s.lobbyRepo.Create(lobby)
	s.ErrorIs(err, ErrLobbyExists)
}

func (s *LobbySQLRepositoryTestSuite) TestFindByIDWhenLobbyIsPresent() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(lobby)
	found, err := s.lobbyRepo.FindByID(lobby.LobbyID)
	s.Equal(lobby.LobbyID, found.LobbyID)
	s.Equal(fixtureLobbyName, found.Name)
	s.Equal(user.ID, found.CreatorID)
	s.NoError(err)
}

func (s *LobbySQLRepositoryTestSuite) TestFindByIDWhenLobbyIsNotAlreadyPresentShouldReturnErrLobbyNotFound() {
	found, err := s.lobbyRepo.FindByID("not existing Lobby ID")
	s.Nil(found)
	s.ErrorIs(err, ErrLobbyNotFound)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenLobbyAndOpponentArePresent() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	opponent := s.createTestUser("enemy", "12345")
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(lobby)
	s.usrRepo.On("FindByID", opponent.ID).Return(&models.User{}, nil)
	err := s.lobbyRepo.UpdateLobbyOpponentAndStatus(lobby.LobbyID, opponent.ID, models.LobbyStatusInProgress)
	s.db.Model(&models.Lobby{}).First(&lobby)
	s.NoError(err)
	s.Equal(opponent.ID, *lobby.OpponentID)
	s.Equal(models.LobbyStatusInProgress, lobby.Status)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenOpponentIsMissing() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(lobby)
	notExistentOppenentId := uint(10)
	s.usrRepo.On("FindByID", notExistentOppenentId).Return(&models.User{}, usrrepo.ErrUserNotFound)
	err := s.lobbyRepo.UpdateLobbyOpponentAndStatus(lobby.LobbyID, notExistentOppenentId, models.LobbyStatusInProgress)
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenLobbyIsMissing() {
	opponent := s.createTestUser("enemy", "12345")
	s.usrRepo.On("FindByID", opponent.ID).Return(&models.User{}, nil)
	err := s.lobbyRepo.UpdateLobbyOpponentAndStatus("not existing lobby ID", opponent.ID, models.LobbyStatusInProgress)
	s.ErrorIs(err, ErrLobbyNotFound)
}

func (s *LobbySQLRepositoryTestSuite) TestListAvailableWhenThereAreWaitingLobbies() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	firstLobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(firstLobby)
	secondLobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      "secondLobby",
		CreatorID: user.ID,
	}
	s.db.Create(secondLobby)

	lobbies := s.lobbyRepo.ListAvailable()
	s.Len(lobbies, 2)
	s.Equal(lobbies[0].LobbyID, firstLobby.LobbyID)
	s.Equal(lobbies[1].LobbyID, secondLobby.LobbyID)
}

func (s *LobbySQLRepositoryTestSuite) TestListAvailableWhenDatabseIsEmpty() {
	lobbies := s.lobbyRepo.ListAvailable()
	s.Empty(lobbies)
}

func (s *LobbySQLRepositoryTestSuite) TestListAvailableWhenThereAreNotWaitingLobbies() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	firstLobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
		Status:    models.LobbyStatusInProgress,
	}
	s.db.Create(firstLobby)
	secondLobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      "secondLobby",
		CreatorID: user.ID,
		Status:    models.LobbyStatusInProgress,
	}
	s.db.Create(secondLobby)

	lobbies := s.lobbyRepo.ListAvailable()
	s.Empty(lobbies)
}

func TestLobbyRepository(t *testing.T) {
	suite.Run(t, new(LobbySQLRepositoryTestSuite))
}
