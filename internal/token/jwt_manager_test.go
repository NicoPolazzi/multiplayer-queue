package token

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

type TokenManagerTestSuite struct {
	suite.Suite
	TokenManager
}

func (s *TokenManagerTestSuite) SetupTest() {
	s.TokenManager = NewJWTTokenManager([]byte("test-secret"))
}

func (s *TokenManagerTestSuite) TestCreateWhenThereIsASigningErrorShouldNotCreateTheToken() {
	mockErr := errors.New("mock signing error")
	originalSignedString := signedString
	defer func() { signedString = originalSignedString }()

	signedString = func(token *jwt.Token, key any) (string, error) {
		return "", mockErr
	}

	tokenString, err := s.TokenManager.Create("testuser", time.Hour)

	s.ErrorIs(err, ErrImpossibleCreation)
	s.Empty(tokenString)
}

func TestJWTTokenManager(t *testing.T) {
	suite.Run(t, new(TokenManagerTestSuite))
}
