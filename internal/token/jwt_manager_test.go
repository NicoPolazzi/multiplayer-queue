package token

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

const fixtureSecret string = "test-secret"

type TokenManagerTestSuite struct {
	suite.Suite
	TokenManager
}

func (s *TokenManagerTestSuite) SetupTest() {
	s.TokenManager = NewJWTTokenManager([]byte(fixtureSecret))
}

func (s *TokenManagerTestSuite) TestCreateWhenThereIsASigningErrorShouldNotCreateTheToken() {
	mockErr := errors.New("mock signing error")
	originalSignedString := signedString
	defer func() { signedString = originalSignedString }()

	signedString = func(token *jwt.Token, key any) (string, error) {
		return "", mockErr
	}

	tokenString, err := s.TokenManager.Create("testuser")

	s.ErrorIs(err, ErrImpossibleCreation)
	s.Empty(tokenString)
}

func (s *TokenManagerTestSuite) TestCreateSuccess() {
	tokenString, err := s.TokenManager.Create("testuser")

	s.NoError(err)
	s.NotEmpty(tokenString)
}

func (s *TokenManagerTestSuite) TestValidateWhenTheFormatIsInvalidShouldRaiseInvalidTokenError() {
	username, err := s.TokenManager.Validate("not-a-valid-token")
	s.ErrorIs(err, ErrInvalidToken)
	s.Empty(username)
}

func (s *TokenManagerTestSuite) TestValidateWhenTheTokenIsExpiredShouldRaiseInvalidTokenError() {
	expiredToken := createExpiredToken([]byte(fixtureSecret), "testuser")
	username, err := s.TokenManager.Validate(expiredToken)
	s.ErrorIs(err, ErrInvalidToken)
	s.Empty(username)
}

func createExpiredToken(secretKey []byte, username string) string {
	claims := &claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secretKey)
	return tokenString
}

func (s *TokenManagerTestSuite) TestValidateSuccess() {
	expectedUsername := "testuser"
	tokenClaims := &claims{
		Username: expectedUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   expectedUsername,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, _ := token.SignedString([]byte(fixtureSecret))

	username, err := s.TokenManager.Validate(tokenString)

	s.NoError(err)
	s.Equal(expectedUsername, username)
}

func TestJWTTokenManager(t *testing.T) {
	suite.Run(t, new(TokenManagerTestSuite))
}
