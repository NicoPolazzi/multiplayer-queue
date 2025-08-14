package lobbyrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"gorm.io/gorm"
)

type sqlLobbyRepository struct {
	db *gorm.DB
}

func NewSQLLobbyRepository(db *gorm.DB) LobbyRepository {
	return &sqlLobbyRepository{db: db}
}

func (r *sqlLobbyRepository) Create(lobby *models.Lobby) error {
	if err := r.db.First(&models.User{}, lobby.CreatorID).Error; err != nil {
		return usrrepo.ErrUserNotFound
	}

	if err := r.db.Create(lobby).Error; err != nil {
		return ErrLobbyExists
	}
	return nil
}

func (r *sqlLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	var retrievedLobby models.Lobby
	result := r.db.Where(&models.Lobby{LobbyID: lobbyID}).First(&retrievedLobby)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrLobbyNotFound
	}
	return &retrievedLobby, nil
}

func (r *sqlLobbyRepository) UpdateLobbyOpponentAndStatus(lobbyID string, opponentID uint, status models.LobbyStatus) error {
	if err := r.db.First(&models.User{}, opponentID).Error; err != nil {
		return usrrepo.ErrUserNotFound
	}

	result := r.db.Model(&models.Lobby{}).Where("lobby_id = ?", lobbyID).
		Updates(map[string]any{"opponent_id": &opponentID, "status": status})

	if result.RowsAffected == 0 {
		return ErrLobbyNotFound
	}

	return nil
}

// ListAvailable implements LobbyRepository.
func (s *sqlLobbyRepository) ListAvailable() ([]models.Lobby, error) {
	panic("unimplemented")
}
