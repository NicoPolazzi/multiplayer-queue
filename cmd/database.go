package main

import (
	"fmt"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabaseConnection() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Lobby{}); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return db, nil
}
