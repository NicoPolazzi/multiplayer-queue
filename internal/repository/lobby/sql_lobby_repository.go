package lobbyrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"gorm.io/gorm"
)

type sqlLobbyRepository struct {
	db             *gorm.DB
	userRepository usrrepo.UserRepository
}

func NewSQLLobbyRepository(db *gorm.DB, userRepository usrrepo.UserRepository) LobbyRepository {
	return &sqlLobbyRepository{db: db, userRepository: userRepository}
}

func (r *sqlLobbyRepository) Create(lobby *models.Lobby) (*models.Lobby, error) {
	_, err := r.userRepository.FindByID(lobby.CreatorID)
	if err != nil {
		return nil, err
	}

	if err := r.db.Create(lobby).Error; err != nil {
		return nil, ErrLobbyExists
	}

	var newLobby models.Lobby
	result := r.db.First(&newLobby)
	return &newLobby, result.Error
}

func (r *sqlLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	var retrievedLobby models.Lobby
	result := r.db.Preload("Creator").Preload("Opponent").Where("lobby_id = ?", lobbyID).First(&retrievedLobby)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrLobbyNotFound
	}
	return &retrievedLobby, nil
}

func (r *sqlLobbyRepository) UpdateLobbyOpponentAndStatus(lobbyID string, opponentID uint, status models.LobbyStatus) error {
	_, err := r.userRepository.FindByID(opponentID)
	if err != nil {
		return err
	}

	result := r.db.Model(&models.Lobby{}).Where("lobby_id = ?", lobbyID).
		Updates(map[string]any{"opponent_id": &opponentID, "status": status})

	if result.RowsAffected == 0 {
		return ErrLobbyNotFound
	}

	return nil
}

// The available lobbies are the onces that are waiting for users
func (r *sqlLobbyRepository) ListAvailable() []models.Lobby {
	var lobbies []models.Lobby
	r.db.Preload("Creator").Where("status = ?", models.LobbyStatusWaiting).Find(&lobbies)
	return lobbies
}
