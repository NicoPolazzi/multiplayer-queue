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
	Create(name string, creator *models.User) (*models.Lobby, error)
	FindByID(lobbyID string) (*models.Lobby, error)
	ListAvailable() []*models.Lobby
	AddPlayerAndSetStatus(lobbyID string, player *models.User, status models.LobbyStatus) error
	UpdateLobbyWinnerAndStatus(lobbyID string, winnerID uint, status models.LobbyStatus) error
}
