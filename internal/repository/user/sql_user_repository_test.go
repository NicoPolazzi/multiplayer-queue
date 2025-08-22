package user

import (
	"testing"

	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	UserFixtureUsername string = "test"
	UserFixturePassword string = "123"
)

type SQLUserRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repository UserRepository
}

func (s *SQLUserRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		s.T().Fatalf("Failed to connect to database at the suite start: %v", err)
	}
	s.db = db
}

func (s *SQLUserRepositoryTestSuite) TearDownSuite() {
	db, _ := s.db.DB()
	if err := db.Close(); err != nil {
		s.T().Fatalf("Failed to close database connection at the suite end: %v", err)
	}
}

func (s *SQLUserRepositoryTestSuite) SetupTest() {
	if err := s.db.Migrator().DropTable(&models.User{}); err != nil {
		s.T().Fatalf("Failed to drop User table before test run: %v", err)
	}

	if err := s.db.AutoMigrate(&models.User{}); err != nil {
		s.T().Fatalf("Failed to migrate User table before test run: %v", err)
	}
	s.repository = NewSQLUserRepository(s.db)
}

func (s *SQLUserRepositoryTestSuite) TestSaveWhenThereIsNotAlreadyTheUser() {
	var retrievedUser models.User
	err := s.repository.Create(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	s.db.First(&retrievedUser)
	s.Equal(UserFixtureUsername, retrievedUser.Username)
	s.Equal(UserFixturePassword, retrievedUser.Password)
	s.NoError(err)
}

func (s *SQLUserRepositoryTestSuite) TestSaveWhenUserIsPresentShouldReturnErrUserExists() {
	existingUser := models.User{Username: UserFixtureUsername, Password: UserFixturePassword}
	s.db.Create(&existingUser)
	err := s.repository.Create(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	s.ErrorIs(err, ErrUserExists)
}

func (s *SQLUserRepositoryTestSuite) TestFindByUsernameWhenUserIsPresent() {
	s.db.Create(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	retrievedUser, err := s.repository.FindByUsername(UserFixtureUsername)
	s.Equal(UserFixtureUsername, retrievedUser.Username)
	s.Equal(UserFixturePassword, retrievedUser.Password)
	s.NoError(err)
}

func (s *SQLUserRepositoryTestSuite) TestFindByUsernameWhenUserIsMissingShouldReturnErrUserNotFound() {
	retrievedUser, err := s.repository.FindByUsername(UserFixtureUsername)
	s.Empty(retrievedUser)
	s.ErrorIs(err, ErrUserNotFound)
}

func (s *SQLUserRepositoryTestSuite) TestFindByIDWhenUserIsPresent() {
	s.db.Create(&models.User{Username: UserFixtureUsername, Password: UserFixturePassword})
	retrievedUser, err := s.repository.FindByID(1)
	s.Equal(UserFixtureUsername, retrievedUser.Username)
	s.Equal(UserFixturePassword, retrievedUser.Password)
	s.NoError(err)
}

func (s *SQLUserRepositoryTestSuite) TestFindByIDWhenUserIsMissingShouldReturnErrUserNotFound() {
	retrievedUser, err := s.repository.FindByID(1)
	s.Empty(retrievedUser)
	s.ErrorIs(err, ErrUserNotFound)
}

func TestSQLUserRepository(t *testing.T) {
	suite.Run(t, new(SQLUserRepositoryTestSuite))
}
