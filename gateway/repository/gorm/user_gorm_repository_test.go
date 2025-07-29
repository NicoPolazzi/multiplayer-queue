package gorm

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestSuite struct {
	suite.Suite
	Repository repository.UserRepository
}

func (s *TestSuite) TestCreateWillSaveUserToDatabase() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&models.User{})
	s.Repository = NewGormUserRepository(db)
	var retrivedUser models.User
	err := s.Repository.Create(&models.User{Username: "test", Password: "123"})
	db.First(&retrivedUser)

	assert.Equal(s.T(), "test", retrivedUser.Username)
	assert.Equal(s.T(), "123", retrivedUser.Password)
	assert.Nil(s.T(), err)
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
