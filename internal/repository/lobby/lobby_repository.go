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
	UpdateLobbyOpponentAndStatus(lobbyID string, opponentID uint, status models.LobbyStatus) error
	ListAvailable() []models.Lobby
}
