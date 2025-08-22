package lobby

import (
	"errors"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"gorm.io/gorm"
)

type sqlLobbyRepository struct {
	db *gorm.DB
}

func NewSQLLobbyRepository(db *gorm.DB) LobbyRepository {
	return &sqlLobbyRepository{db: db}
}

func (r *sqlLobbyRepository) Create(lobby *models.Lobby) error {
	return r.db.Create(lobby).Error
}

func (r *sqlLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	var lobby models.Lobby
	result := r.db.Preload("Players").Preload("Winner").First(&lobby, "lobby_id = ?", lobbyID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrLobbyNotFound
	}
	return &lobby, result.Error
}

func (r *sqlLobbyRepository) ListAvailable() []*models.Lobby {
	var lobbies []*models.Lobby
	r.db.Preload("Players").
		Where("status = ?", models.LobbyStatusWaiting).
		Find(&lobbies)
	return lobbies
}

func (r *sqlLobbyRepository) AddPlayer(lobby *models.Lobby, player *models.User) error {
	return r.db.Model(lobby).Association("Players").Append(player)
}

func (r *sqlLobbyRepository) UpdateStatus(lobby *models.Lobby, status models.LobbyStatus) error {
	return r.db.Model(lobby).Update("status", status).Error
}

func (r *sqlLobbyRepository) UpdateWinner(lobby *models.Lobby, winnerID uint) error {
	return r.db.Model(lobby).Update("winner_id", winnerID).Error
}

func (r *sqlLobbyRepository) Delete(lobbyID string) error {
	var lobby models.Lobby
	if err := r.db.First(&lobby, "lobby_id = ?", lobbyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLobbyNotFound
		}
	}

	if err := r.db.Model(&lobby).Association("Players").Clear(); err != nil {
		return err
	}

	r.db.Delete(&lobby)
	return nil
}
