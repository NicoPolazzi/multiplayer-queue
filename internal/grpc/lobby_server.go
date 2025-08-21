package grpc

import (
	"context"
	"errors"
	"math/rand"

	"github.com/NicoPolazzi/multiplayer-queue/gen/lobby"
	"github.com/NicoPolazzi/multiplayer-queue/internal/models"
	lobbyrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/lobby"
	usrrepo "github.com/NicoPolazzi/multiplayer-queue/internal/repository/user"
	"github.com/google/uuid"
)

type LobbyServer struct {
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

	newLobby := &models.Lobby{
		LobbyID: uuid.New().String(),
		Name:    req.GetName(),
		Players: []models.User{*creator},
		Status:  models.LobbyStatusWaiting,
	}

	if err := s.lobbyRepo.Create(newLobby); err != nil {
		return nil, err
	}

	return toProtoLobby(newLobby), nil
}

func (s *LobbyServer) JoinLobby(ctx context.Context, req *lobby.JoinLobbyRequest) (*lobby.Lobby, error) {
	player, err := s.userRepo.FindByUsername(req.GetUsername())
	if err != nil {
		return nil, err
	}

	lobbyToJoin, err := s.lobbyRepo.FindByID(req.GetLobbyId())
	if err != nil {
		return nil, err
	}

	if len(lobbyToJoin.Players) >= 2 {
		return nil, errors.New("lobby is full")
	}

	if err := s.lobbyRepo.AddPlayer(lobbyToJoin, player); err != nil {
		return nil, err
	}

	if err := s.lobbyRepo.UpdateStatus(lobbyToJoin, models.LobbyStatusInProgress); err != nil {
		return nil, err
	}

	lobbyToJoin.Players = append(lobbyToJoin.Players, *player)
	lobbyToJoin.Status = models.LobbyStatusInProgress
	return toProtoLobby(lobbyToJoin), nil
}

func (s *LobbyServer) FinishGame(ctx context.Context, req *lobby.FinishGameRequest) (*lobby.Lobby, error) {
	gameLobby, err := s.lobbyRepo.FindByID(req.GetLobbyId())
	if err != nil {
		return nil, err
	}

	winnerIndex := rand.Intn(len(gameLobby.Players))
	winner := gameLobby.Players[winnerIndex]

	if err := s.lobbyRepo.UpdateWinner(gameLobby, winner.ID); err != nil {
		return nil, err
	}

	if err := s.lobbyRepo.UpdateStatus(gameLobby, models.LobbyStatusFinished); err != nil {
		return nil, err
	}

	gameLobby.Winner = &winner
	gameLobby.Status = models.LobbyStatusFinished
	return toProtoLobby(gameLobby), nil
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
	protoLobbies := make([]*lobby.Lobby, len(lobbies))
	for i, l := range lobbies {
		protoLobbies[i] = toProtoLobby(l)
	}
	return &lobby.ListAvailableLobbiesResponse{Lobbies: protoLobbies}, nil
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
