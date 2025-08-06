package models

import (
	"gorm.io/gorm"
)

type LobbyStatus string

const (
	LobbyStatusWaiting    LobbyStatus = "WAITING"     // Waiting for opponent
	LobbyStatusInProgress LobbyStatus = "IN_PROGRESS" // Game is in progress
)

type Lobby struct {
	gorm.Model
	LobbyID    string `gorm:"uniqueIndex;not null"`
	Name       string `gorm:"not null"`
	CreatorID  uint   `gorm:"not null"`
	Creator    User   `gorm:"foreignKey:CreatorID"`
	OpponentID *uint
	Opponent   *User       `gorm:"foreignKey:OpponentID"`
	Status     LobbyStatus `gorm:"type:string;not null;default:'WAITING'"`
}
