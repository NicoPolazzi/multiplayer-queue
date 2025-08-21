package grpc

import (
	"context"
	"math/rand"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
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
	creator, err := s.userRepo.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, err
	}

	createdLobby, err := s.lobbyRepo.Create(req.GetName(), creator)
	if err != nil {
		return nil, err
	}

	return toProtoLobby(createdLobby), nil
}

func toProtoLobby(m *models.Lobby) *lobby.Lobby {
	pLobby := &lobby.Lobby{
		LobbyId: m.LobbyID,
		Name:    m.Name,
		Status:  string(m.Status),
		Players: make([]*lobby.Player, len(m.Players)),
	}

	for i, player := range m.Players {
		pLobby.Players[i] = &lobby.Player{
			Id:       uint32(player.ID),
			Username: player.Username,
		}
	}

	if m.WinnerID != nil {
		winnerID := uint32(*m.WinnerID)
		pLobby.WinnerId = &winnerID
		if m.Winner != nil {
			pLobby.WinnerUsername = &m.Winner.Username
		}
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

func (s *LobbyServer) JoinLobby(ctx context.Context, req *lobby.JoinLobbyRequest) (*lobby.Lobby, error) {
	player, err := s.userRepo.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, err
	}

	err = s.lobbyRepo.AddPlayerAndSetStatus(req.GetLobbyId(), player, models.LobbyStatusInProgress)
	if err != nil {
		return nil, err
	}

	joinedLobby, err := s.lobbyRepo.FindByID(req.GetLobbyId())
	if err != nil {
		return nil, err
	}

	return toProtoLobby(joinedLobby), nil
}

func (s *LobbyServer) FinishGame(ctx context.Context, req *lobby.FinishGameRequest) (*lobby.Lobby, error) {
	gameLobby, err := s.lobbyRepo.FindByID(req.GetLobbyId())
	if err != nil {
		return nil, err
	}

	winnerIndex := rand.Intn(len(gameLobby.Players))
	winnerID := gameLobby.Players[winnerIndex].ID

	err = s.lobbyRepo.UpdateLobbyWinnerAndStatus(req.GetLobbyId(), winnerID, models.LobbyStatusFinished)
	if err != nil {
		return nil, err
	}

	finishedLobby, err := s.lobbyRepo.FindByID(req.GetLobbyId())
	if err != nil {
		return nil, err
	}

	return toProtoLobby(finishedLobby), nil
}

func (s *LobbyServer) ListAvailableLobbies(ctx context.Context, req *lobby.ListAvailableLobbiesRequest) (*lobby.ListAvailableLobbiesResponse, error) {
	lobbies := s.lobbyRepo.ListAvailable()

	protoLobbies := make([]*lobby.Lobby, len(lobbies))
	for i, l := range lobbies {
		protoLobbies[i] = toProtoLobby(l)
	}

	return &lobby.ListAvailableLobbiesResponse{Lobbies: protoLobbies}, nil
}
