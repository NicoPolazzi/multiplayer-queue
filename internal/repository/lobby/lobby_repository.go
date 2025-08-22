package lobbyrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
)

var (
	ErrLobbyExists   = errors.New("lobby already exists in the database")
	ErrLobbyNotFound = errors.New("lobby not found in the database")
)

type LobbyRepository interface {
	Create(lobby *models.Lobby) error
	FindByID(lobbyID string) (*models.Lobby, error)
	AddPlayer(lobby *models.Lobby, player *models.User) error
	UpdateStatus(lobby *models.Lobby, status models.LobbyStatus) error
	UpdateWinner(lobby *models.Lobby, winnerID uint) error
	Delete(lobbyID string) error
	ListAvailable() []*models.Lobby
}
