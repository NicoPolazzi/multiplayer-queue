package sqlrepo

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/gateway/models"
	"github.com/NicoPolazzi/multiplayer-queue/gateway/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	UserFixtureUsername string = "test"
	UserFixturePassword string = "123"
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
	s.Repository = NewSQLUserRepository(s.DB)
}

func (s *TestSuite) TestSaveWhenThereIsNotUser() {
	var retrievedUser models.User
	err := s.Repository.Save(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	s.DB.First(&retrievedUser)
	assert.Equal(s.T(), UserFixtureUsername, retrievedUser.Username)
	assert.Equal(s.T(), UserFixturePassword, retrievedUser.Password)
	assert.Nil(s.T(), err)
}

func (s *TestSuite) TestSaveWhenAnExistingUserIsPresentShouldRaiseAnError() {
	existingUser := models.User{Username: UserFixtureUsername, Password: UserFixturePassword}
	err := s.DB.Create(&existingUser).Error
	err = s.Repository.Save(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	assert.ErrorIs(s.T(), err, repository.ErrUserExists)
}

func (s *TestSuite) TestFindByUsernameWhenThereIsAnUserShouldRetrieveTheUser() {
	s.DB.Create(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	retrievedUser, err := s.Repository.FindByUsername(UserFixtureUsername)
	assert.Equal(s.T(), UserFixtureUsername, retrievedUser.Username)
	assert.Equal(s.T(), UserFixturePassword, retrievedUser.Password)
	assert.Nil(s.T(), err)
}

func (s *TestSuite) TestFindByUsernameWhenThereIsNotAnUserShouldThrowError() {
	retrievedUser, err := s.Repository.FindByUsername(UserFixtureUsername)
	assert.Nil(s.T(), retrievedUser)
	assert.ErrorIs(s.T(), err, repository.ErrUserNotFound)
}

func TestUserSQLRepository(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
