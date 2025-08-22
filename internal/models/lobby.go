package models

import (
	"time"

	"gorm.io/gorm"
)

type LobbyStatus string

const (
	LobbyStatusWaiting    LobbyStatus = "WAITING"     // Waiting for opponent
	LobbyStatusInProgress LobbyStatus = "IN_PROGRESS" // Game is in progress
	LobbyStatusFinished   LobbyStatus = "FINISHED"    // Game has finished
)

type Lobby struct {
	LobbyID   string `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Players   []User `gorm:"foreignKey:LobbyID"`
	WinnerID  *uint
	Winner    *User       `gorm:"foreignKey:WinnerID"`
	Status    LobbyStatus `gorm:"type:string;not null;default:'WAITING'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
