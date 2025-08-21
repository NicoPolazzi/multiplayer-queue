package lobbyrepo

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlLobbyRepository struct {
	db             *gorm.DB
	userRepository usrrepo.UserRepository
}

func NewSQLLobbyRepository(db *gorm.DB, userRepository usrrepo.UserRepository) LobbyRepository {
	return &sqlLobbyRepository{db: db, userRepository: userRepository}
}

func (r *sqlLobbyRepository) Create(name string, creator *models.User) (*models.Lobby, error) {
	lobby := &models.Lobby{
		LobbyID: uuid.New().String(),
		Name:    name,
	}

	if err := r.db.Create(lobby).Error; err != nil {
		return nil, ErrLobbyExists
	}

	if err := r.db.Model(creator).Update("lobby_id", &lobby.LobbyID).Error; err != nil {
		r.db.Delete(lobby)
		return nil, err
	}

	lobby.Players = append(lobby.Players, *creator)
	return lobby, nil
}

func (r *sqlLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	var lobby models.Lobby
	result := r.db.Preload("Players").Preload("Winner").Where("lobby_id = ?", lobbyID).First(&lobby)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrLobbyNotFound
	}
	return &lobby, nil
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
func (r *sqlLobbyRepository) ListAvailable() []*models.Lobby {
	var lobbies []*models.Lobby
	r.db.Preload("Players").
		Where("status = ?", models.LobbyStatusWaiting).
		Find(&lobbies)

	return lobbies
}

func (r *sqlLobbyRepository) AddPlayerAndSetStatus(lobbyID string, player *models.User, status models.LobbyStatus) error {
	if err := r.db.Model(player).Update("lobby_id", &lobbyID).Error; err != nil {
		return err
	}

	if err := r.db.Model(&models.Lobby{}).Where("lobby_id = ?", lobbyID).Update("status", status).Error; err != nil {
		r.db.Model(player).Update("lobby_id", nil)
		return err
	}

	return nil
}

func (r *sqlLobbyRepository) UpdateLobbyWinnerAndStatus(lobbyID string, winnerID uint, status models.LobbyStatus) error {
	return r.db.Model(&models.Lobby{}).Where("lobby_id = ?", lobbyID).Updates(map[string]interface{}{
		"winner_id": winnerID,
		"status":    status,
	}).Error
}
