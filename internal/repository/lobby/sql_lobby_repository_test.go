package lobby

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	fixtureLobbyName      = "Test Lobby"
	fixtureLobbyCondition = "lobby_id = ?"
)

type LobbySQLRepositoryTestSuite struct {
	suite.Suite
	db        *gorm.DB
	lobbyRepo LobbyRepository
}

func (s *LobbySQLRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.Require().NoError(err, "Failed to connect to the database")
	s.db = db
}

func (s *LobbySQLRepositoryTestSuite) TearDownSuite() {
	db, _ := s.db.DB()
	err := db.Close()
	s.Require().NoError(err, "Failed to close the database connection")
}

func (s *LobbySQLRepositoryTestSuite) SetupTest() {
	err := s.db.Migrator().DropTable(&models.User{}, &models.Lobby{})
	s.Require().NoError(err)
	err = s.db.AutoMigrate(&models.User{}, &models.Lobby{})
	s.Require().NoError(err)

	s.lobbyRepo = NewSQLLobbyRepository(s.db)
}

func (s *LobbySQLRepositoryTestSuite) createUserInDB(username string, lobbyID *string) models.User {
	user := models.User{Username: username, Password: "password", LobbyID: lobbyID}
	err := s.db.Create(&user).Error
	s.Require().NoError(err)
	return user
}

func (s *LobbySQLRepositoryTestSuite) TestCreateSuccess() {
	creator := s.createUserInDB("creator", nil)
	lobbyToCreate := &models.Lobby{
		LobbyID: uuid.New().String(),
		Name:    fixtureLobbyName,
		Players: []models.User{creator},
	}

	err := s.lobbyRepo.Create(lobbyToCreate)
	var addedLobby models.Lobby
	s.db.First(&addedLobby)
	s.NoError(err)
	s.Equal(lobbyToCreate.Name, addedLobby.Name)
	s.Equal(lobbyToCreate.LobbyID, addedLobby.LobbyID)
	var updatedCreator models.User
	s.db.First(&updatedCreator, creator.ID)
	s.Equal(lobbyToCreate.LobbyID, *updatedCreator.LobbyID)
}

func (s *LobbySQLRepositoryTestSuite) TestFindByIdSuccess() {
	lobby := s.createLobbyInDB("FindMe", models.LobbyStatusWaiting)
	s.createUserInDB("player1", &lobby.LobbyID)
	foundLobby, err := s.lobbyRepo.FindByID(lobby.LobbyID)
	s.NoError(err)
	s.Equal(lobby.LobbyID, foundLobby.LobbyID)
	s.Len(foundLobby.Players, 1)
}

func (s *LobbySQLRepositoryTestSuite) createLobbyInDB(name string, status models.LobbyStatus) models.Lobby {
	lobby := models.Lobby{
		LobbyID: uuid.New().String(),
		Name:    name,
		Status:  status,
	}
	err := s.db.Create(&lobby).Error
	s.Require().NoError(err)
	return lobby
}

func (s *LobbySQLRepositoryTestSuite) TestFindByIdNotFound() {
	lobby, err := s.lobbyRepo.FindByID("non-existent-id")
	s.ErrorIs(err, ErrLobbyNotFound)
	s.Empty(lobby)
}

func (s *LobbySQLRepositoryTestSuite) TestListAvailableWhenThereAreWaitingLobbies() {
	s.createLobbyInDB("Lobby 1", models.LobbyStatusWaiting)
	s.createLobbyInDB("Lobby 2", models.LobbyStatusWaiting)
	s.createLobbyInDB("Lobby 3", models.LobbyStatusInProgress)
	lobbies := s.lobbyRepo.ListAvailable()
	s.Len(lobbies, 2)
}

func (s *LobbySQLRepositoryTestSuite) TestListAvailableWhenThereAreNotWaitingLobbies() {
	s.createLobbyInDB("Full Lobby", models.LobbyStatusInProgress)
	lobbies := s.lobbyRepo.ListAvailable()
	s.Empty(lobbies)
}

func (s *LobbySQLRepositoryTestSuite) TestAddPlayerSuccess() {
	lobby := s.createLobbyInDB(fixtureLobbyName, models.LobbyStatusWaiting)
	player := s.createUserInDB("new_player", nil)
	err := s.lobbyRepo.AddPlayer(&lobby, &player)
	s.NoError(err)
	var updatedPlayer models.User
	s.db.First(&updatedPlayer, player.ID)
	s.Equal(lobby.LobbyID, *updatedPlayer.LobbyID)
	s.Len(lobby.Players, 1)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateStatusSuccess() {
	lobby := s.createLobbyInDB("Status Test Lobby", models.LobbyStatusWaiting)
	err := s.lobbyRepo.UpdateStatus(&lobby, models.LobbyStatusInProgress)
	s.NoError(err)
	var updatedLobby models.Lobby
	s.db.First(&updatedLobby, fixtureLobbyCondition, lobby.LobbyID)
	s.Equal(models.LobbyStatusInProgress, updatedLobby.Status)
}

func (s *LobbySQLRepositoryTestSuite) TestUpdateWinnerSuccess() {
	lobby := s.createLobbyInDB("Winner Test Lobby", models.LobbyStatusInProgress)
	winner := s.createUserInDB("the_winner", &lobby.LobbyID)
	err := s.lobbyRepo.UpdateWinner(&lobby, winner.ID)
	s.NoError(err)
	var updatedLobby models.Lobby
	s.db.First(&updatedLobby, fixtureLobbyCondition, lobby.LobbyID)
	s.Equal(winner.ID, *updatedLobby.WinnerID)
}

func (s *LobbySQLRepositoryTestSuite) TestDeleteSuccess() {
	lobby := s.createLobbyInDB(fixtureLobbyName, models.LobbyStatusInProgress)
	player1 := s.createUserInDB("player1", &lobby.LobbyID)
	player2 := s.createUserInDB("player2", &lobby.LobbyID)
	err := s.lobbyRepo.Delete(lobby.LobbyID)
	s.NoError(err)
	err = s.db.First(&lobby, fixtureLobbyCondition, lobby.LobbyID).Error
	s.ErrorIs(err, gorm.ErrRecordNotFound)
	var updatedPlayer1 models.User
	s.db.First(&updatedPlayer1, player1.ID)
	s.Empty(updatedPlayer1.LobbyID)
	var updatedPlayer2 models.User
	s.db.First(&updatedPlayer2, player2.ID)
	s.Empty(updatedPlayer2.LobbyID)
}

func (s *LobbySQLRepositoryTestSuite) TestDeleteWhenAssociationClearFails() {
	lobby := s.createLobbyInDB(fixtureLobbyName, models.LobbyStatusInProgress)
	s.createUserInDB("a_player", &lobby.LobbyID)

	err := s.db.Migrator().DropTable(&models.User{})
	s.Require().NoError(err, "Dropping user table for test setup should not fail")

	deleteErr := s.lobbyRepo.Delete(lobby.LobbyID)
	s.ErrorIs(deleteErr, ErrLobbyCleanupFailed)

	// Verify the lobby was NOT deleted, as the process failed before the final deletion step.
	var foundLobby models.Lobby
	findErr := s.db.First(&foundLobby, fixtureLobbyCondition, lobby.LobbyID).Error
	s.NoError(findErr, "Lobby should still exist because the transaction should have failed")
}

func (s *LobbySQLRepositoryTestSuite) TestDeleteNotFound() {
	err := s.lobbyRepo.Delete("non-existent-id")
	s.ErrorIs(err, ErrLobbyNotFound)
}

func TestLobbyRepository(t *testing.T) {
	suite.Run(t, new(LobbySQLRepositoryTestSuite))
}
