package lobbyrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
)

var (
	ErrLobbyExists = errors.New("lobby already exists in the database")
)

type LobbyRepository interface {
	Create(lobby *models.Lobby) error
	FindByID(lobbyID string) (*models.Lobby, error)
	Join(lobbyID string, opponentID uint) error
	ListAvailable() ([]models.Lobby, error)
}
