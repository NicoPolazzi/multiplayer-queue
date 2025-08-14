package lobbyrepo

import (
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

// FindByID implements LobbyRepository.
func (r *sqlLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	var retrievedLobby models.Lobby
	r.db.Where(&models.Lobby{LobbyID: lobbyID}).First(&retrievedLobby)
	return &retrievedLobby, nil
}

// Join implements LobbyRepository.
func (r *sqlLobbyRepository) Join(lobbyID string, opponentID uint) error {
	panic("unimplemented")
}

// ListAvailable implements LobbyRepository.
func (s *sqlLobbyRepository) ListAvailable() ([]models.Lobby, error) {
	panic("unimplemented")
}
