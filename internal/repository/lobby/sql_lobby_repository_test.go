package lobbyrepo

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	db         *gorm.DB
	repository LobbyRepository
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

	s.repository = NewSQLLobbyRepository(s.db)
}

func (s *LobbySQLRepositoryTestSuite) TestCreateLobbyWhenThereIsNotAlreadyTheLobby() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}

	err := s.repository.Create(lobby)
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

	err := s.repository.Create(lobby)
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
	err := s.repository.Create(lobby)
	s.ErrorIs(err, ErrLobbyExists)
}

func (s *LobbySQLRepositoryTestSuite) TestFindByIDShouldReturnTheLobby() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(lobby)
	found, err := s.repository.FindByID(lobby.LobbyID)
	s.Equal(lobby.LobbyID, found.LobbyID)
	s.Equal(fixtureLobbyName, found.Name)
	s.Equal(user.ID, found.CreatorID)
	s.NoError(err)
}

func (s *LobbySQLRepositoryTestSuite) TestFindByIDWhenThereIsNotAlreadyALobbyShouldReturnErrLobbyNotFound() {
	found, err := s.repository.FindByID("not existing Lobby ID")
	s.Nil(found)
	s.ErrorIs(err, ErrLobbyNotFound)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenThereIsAlreadyLobbyAndOpponent() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	opponent := s.createTestUser("enemy", "12345")
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.db.Create(lobby)
	err := s.repository.UpdateLobbyOpponentAndStatus(lobby.LobbyID, opponent.ID, models.LobbyStatusInProgress)
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
	err := s.repository.UpdateLobbyOpponentAndStatus(lobby.LobbyID, notExistentOppenentId, models.LobbyStatusInProgress)
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenLobbyIsMissing() {
	opponent := s.createTestUser("enemy", "12345")
	err := s.repository.UpdateLobbyOpponentAndStatus("not existing lobby ID", opponent.ID, models.LobbyStatusInProgress)
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

	lobbies := s.repository.ListAvailable()
	s.Len(lobbies, 2)
	s.Equal(lobbies[0].LobbyID, firstLobby.LobbyID)
	s.Equal(lobbies[1].LobbyID, secondLobby.LobbyID)
}

func (s *LobbySQLRepositoryTestSuite) TestListAvailableWhenDatabseIsEmpty() {
	lobbies := s.repository.ListAvailable()
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

	lobbies := s.repository.ListAvailable()
	s.Empty(lobbies)
}

func TestLobbyRepository(t *testing.T) {
	suite.Run(t, new(LobbySQLRepositoryTestSuite))
}
