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

func (s *sqlLobbyRepository) Create(lobby *models.Lobby) error {
	if err := s.db.First(&models.User{}, lobby.CreatorID).Error; err != nil {
		return usrrepo.ErrUserNotFound
	}

	if err := s.db.Create(lobby).Error; err != nil {
		return ErrLobbyExists
	}
	return nil
}

// FindByID implements LobbyRepository.
func (s *sqlLobbyRepository) FindByID(lobbyID string) (*models.Lobby, error) {
	panic("unimplemented")
}

// Join implements LobbyRepository.
func (s *sqlLobbyRepository) Join(lobbyID string, opponentID uint) error {
	panic("unimplemented")
}

// ListAvailable implements LobbyRepository.
func (s *sqlLobbyRepository) ListAvailable() ([]models.Lobby, error) {
	panic("unimplemented")
}
