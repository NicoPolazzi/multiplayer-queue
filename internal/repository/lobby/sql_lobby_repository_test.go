package lobbyrepo

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usr "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
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

type LobbyRepositoryTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	Repository LobbyRepository
}

func (s *LobbyRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.T().Fatalf("Failed to connect to database in the suite start: %v", err)
	}
	s.DB = db
}

func (s *LobbyRepositoryTestSuite) TearDownSuite() {
	db, _ := s.DB.DB()
	if err := db.Close(); err != nil {
		s.T().Fatalf("Failed to close database connection: %v", err)
	}
}

func (s *LobbyRepositoryTestSuite) SetupTest() {
	if err := s.DB.Migrator().DropTable(&models.User{}, &models.Lobby{}); err != nil {
		s.T().Fatalf("Failed to drop User table before test run: %v", err)
	}

	if err := s.DB.AutoMigrate(&models.User{}, &models.Lobby{}); err != nil {
		s.T().Fatalf("Failed to migrate User and Lobby tables before test run: %v", err)
	}

	s.Repository = NewSQLLobbyRepository(s.DB)
}

func (s *LobbyRepositoryTestSuite) TestCreateLobbyWhenThereIsNotAlreadyTheLobby() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}

	err := s.Repository.Create(lobby)
	assert.Nil(s.T(), err)
}

func (s *LobbyRepositoryTestSuite) createTestUser(username, password string) models.User {
	user := models.User{Username: username, Password: password}
	s.DB.Create(&user)
	return user
}

func (s *LobbyRepositoryTestSuite) TestCreateLobbyWhenUserIsNotAlreadyPresentShouldReturnUserNotFoundError() {
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: 1,
	}

	err := s.Repository.Create(lobby)
	s.ErrorIs(err, usr.ErrUserNotFound)
}

func (s *LobbyRepositoryTestSuite) TestCreateLobbyWhenLobbyIsAlreadySavedShouldReturnLobbyExistsError() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.DB.Create(lobby)
	err := s.Repository.Create(lobby)
	s.ErrorIs(err, ErrLobbyExists)
}

func (s *LobbyRepositoryTestSuite) TestFindByIDShouldReturnTheLobby() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.DB.Create(lobby)
	found, err := s.Repository.FindByID(lobby.LobbyID)
	s.Equal(lobby.LobbyID, found.LobbyID)
	s.Equal(fixtureLobbyName, found.Name)
	s.Equal(user.ID, found.CreatorID)
	s.NoError(err)
}

func (s *LobbyRepositoryTestSuite) TestFindByIDWhenThereIsNotAlreadyALobbyShouldReturnErrLobbyNotFound() {
	found, err := s.Repository.FindByID("not existing Lobby ID")
	s.Nil(found)
	s.ErrorIs(err, ErrLobbyNotFound)
}

func (s *LobbyRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenThereIsAlreadyLobbyAndOpponent() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	opponent := s.createTestUser("enemy", "12345")
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.DB.Create(lobby)
	err := s.Repository.UpdateLobbyOpponentAndStatus(lobby.LobbyID, opponent.ID, models.LobbyStatusInProgress)
	s.DB.Model(&models.Lobby{}).First(&lobby)
	s.NoError(err)
	s.Equal(opponent.ID, *lobby.OpponentID)
	s.Equal(models.LobbyStatusInProgress, lobby.Status)
}

func (s *LobbyRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenOpponentIsMissing() {
	user := s.createTestUser(fixtureUserUsername, fixtureUserPassword)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}
	s.DB.Create(lobby)
	notExistentOppenentId := uint(10)
	err := s.Repository.UpdateLobbyOpponentAndStatus(lobby.LobbyID, notExistentOppenentId, models.LobbyStatusInProgress)
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
}

func (s *LobbyRepositoryTestSuite) TestUpdateLobbyOpponentAndStatusWhenLobbyIsMissing() {
	opponent := s.createTestUser("enemy", "12345")
	err := s.Repository.UpdateLobbyOpponentAndStatus("not existing lobby name", opponent.ID, models.LobbyStatusInProgress)
	s.ErrorIs(err, ErrLobbyNotFound)
}

func TestLobbyRepository(t *testing.T) {
	suite.Run(t, new(LobbyRepositoryTestSuite))
}
