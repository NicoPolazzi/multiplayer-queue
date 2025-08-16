package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type MiddlewareUtilsTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	context  *gin.Context
}

func (s *MiddlewareUtilsTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.recorder = httptest.NewRecorder()
	s.context, _ = gin.CreateTestContext(s.recorder)
}

func TestMiddlewareUtils(t *testing.T) {
	suite.Run(t, new(MiddlewareUtilsTestSuite))
}
