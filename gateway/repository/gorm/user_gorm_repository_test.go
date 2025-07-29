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
	DB         *gorm.DB
	Repository repository.UserRepository
}

func (s *TestSuite) SetupSuite() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.DB = db
}

func (s *TestSuite) TearDownSuite() {
	db, _ := s.DB.DB()
	db.Close()
}

func (s *TestSuite) SetupTest() {
	s.DB.Migrator().DropTable(&models.User{})
	s.DB.AutoMigrate(&models.User{})
	s.Repository = NewGormUserRepository(s.DB)
}

func (s *TestSuite) TestSave() {
	var retrieved models.User
	err := s.Repository.Save(&models.User{Username: "test", Password: "123"})
	s.DB.First(&retrieved)

	assert.Equal(s.T(), "test", retrieved.Username)
	assert.Equal(s.T(), "123", retrieved.Password)
	assert.Nil(s.T(), err)
}

func (s *TestSuite) TestSaveShouldRaiseAnErrorWhenAnExistingUserIsPresent() {
	existingUser := models.User{Username: "test", Password: "123"}
	err := s.Repository.Save(&existingUser)
	err = s.Repository.Save(&models.User{Username: "test", Password: "123"})

	assert.ErrorIs(s.T(), err, repository.ErrUserExists)
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
