package grpc

import (
	"context"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/google/uuid"
)

type LobbyServer struct {
	// This is required to embed the unimplemented methods and satisfy the interface.
	lobby.UnimplementedLobbyServiceServer
	lobbyRepo lobbyrepo.LobbyRepository
	userRepo  usrrepo.UserRepository
}

func NewLobbyServer(lobbyRepo lobbyrepo.LobbyRepository, userRepo usrrepo.UserRepository) *LobbyServer {
	return &LobbyServer{
		lobbyRepo: lobbyRepo,
		userRepo:  userRepo,
	}
}

func (s *LobbyServer) CreateLobby(ctx context.Context, req *lobby.CreateLobbyRequest) (*lobby.Lobby, error) {
	user, err := s.userRepo.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, err
	}

	newLobby := &models.Lobby{
		LobbyID:   uuid.NewString(),
		Name:      req.GetName(),
		CreatorID: user.ID,
		Status:    models.LobbyStatusWaiting,
	}

	createdLobby, err := s.lobbyRepo.Create(newLobby)
	if err != nil {
		return nil, err
	}

	return toProtoLobby(createdLobby), nil
}

func toProtoLobby(m *models.Lobby) *lobby.Lobby {
	pLobby := &lobby.Lobby{
		LobbyId:         m.LobbyID,
		Name:            m.Name,
		CreatorId:       uint32(m.CreatorID),
		Status:          string(m.Status),
		CreatorUsername: m.Creator.Username,
	}

	if m.OpponentID != nil {
		opponentID := uint32(*m.OpponentID)
		pLobby.OpponentId = &opponentID
		pLobby.OpponentUsername = &m.Opponent.Username
	}
	return pLobby
}

func (s *LobbyServer) GetLobby(ctx context.Context, req *lobby.GetLobbyRequest) (*lobby.Lobby, error) {
	foundLobby, err := s.lobbyRepo.FindByID(req.GetLobbyId())
	if err != nil {
		return nil, err
	}
	return toProtoLobby(foundLobby), nil
}

func (s *LobbyServer) ListAvailableLobbies(ctx context.Context, req *lobby.ListAvailableLobbiesRequest) (*lobby.ListAvailableLobbiesResponse, error) {
	lobbies := s.lobbyRepo.ListAvailable()
	var protoLobbies []*lobby.Lobby
	for _, lobby := range lobbies {
		protoLobbies = append(protoLobbies, toProtoLobby(&lobby))
	}
	return &lobby.ListAvailableLobbiesResponse{Lobbies: protoLobbies}, nil
}
