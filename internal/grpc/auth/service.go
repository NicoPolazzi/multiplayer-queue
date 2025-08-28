package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/NicoPolazzi/multiplayer-queue/gen/auth"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/NicoPolazzi/multiplayer-queue/internal/token"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthService implements the gRPC auth service server for user authentication and registration.
type AuthService struct {
	auth.UnimplementedAuthServiceServer
	userRepository usrrepo.UserRepository
	jwtManager     token.TokenManager
}

func NewAuthService(repo usrrepo.UserRepository, manager token.TokenManager) auth.AuthServiceServer {
	return &AuthService{
		userRepository: repo,
		jwtManager:     manager,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, req *auth.RegisterUserRequest) (*auth.User, error) {
	username := req.GetUsername()

	if strings.TrimSpace(username) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username cannot be empty")
	}

	if _, err := s.userRepository.FindByUsername(username); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "username is already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	userModel := &models.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := s.userRepository.Create(userModel); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return toProtoUser(userModel), nil
}

// It checks for the credentials and returns the computed JWT to the caller
func (s *AuthService) LoginUser(ctx context.Context, req *auth.LoginUserRequest) (*auth.LoginUserResponse, error) {
	user, err := s.userRepository.FindByUsername(req.GetUsername())
	if err != nil {
		if errors.Is(err, usrrepo.ErrUserNotFound) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword()))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.jwtManager.Create(user.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create token: %v", err)
	}

	return &auth.LoginUserResponse{
		Token: token,
		User:  toProtoUser(user),
	}, nil
}

func toProtoUser(user *models.User) *auth.User {
	if user == nil {
		return nil
	}

	return &auth.User{
		Id:       uint32(user.ID),
		Username: user.Username,
	}
}
