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
	fixtureLobbyName string = "Test Lobby"
)

type LobbyRepositoryTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	Repository LobbyRepository
}

func (s *LobbyRepositoryTestSuite) SetupSuite() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.DB = db
}

func (s *LobbyRepositoryTestSuite) TearDownSuite() {
	db, _ := s.DB.DB()
	db.Close()
}

func (s *LobbyRepositoryTestSuite) SetupTest() {
	s.DB.Migrator().DropTable(&models.User{})
	s.DB.AutoMigrate(&models.User{}, &models.Lobby{})
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
	s.ErrorIs(err, usrrepo.ErrUserNotFound)
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

func TestLobbyRepository(t *testing.T) {
	suite.Run(t, new(LobbyRepositoryTestSuite))
}
