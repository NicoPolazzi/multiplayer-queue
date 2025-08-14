package lobbyrepo

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usr "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	fixtureLobbyName string = "Test Lobby"
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
	if err := s.DB.Migrator().DropTable(&models.User{}); err != nil {
		s.T().Fatalf("Failed to drop User table before test run: %v", err)
	}

	if err := s.DB.AutoMigrate(&models.User{}, &models.Lobby{}); err != nil {
		s.T().Fatalf("Failed to migrate User and Lobby tables before test run: %v", err)
	}

	s.Repository = NewSQLLobbyRepository(s.DB)
}

func (s *LobbyRepositoryTestSuite) TestCreateLobbyWhenThereIsNotAlreadyTheLobby() {
	user := models.User{Username: "test", Password: "123"}
	s.DB.Create(&user)
	lobby := &models.Lobby{
		LobbyID:   uuid.New().String(),
		Name:      fixtureLobbyName,
		CreatorID: user.ID,
	}

	err := s.Repository.Create(lobby)
	assert.Nil(s.T(), err)
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
	user := models.User{Username: "test", Password: "123"}
	s.DB.Create(&user)
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
	user := models.User{Username: "test", Password: "123"}
	s.DB.Create(&user)
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

func TestLobbyRepository(t *testing.T) {
	suite.Run(t, new(LobbyRepositoryTestSuite))
}
