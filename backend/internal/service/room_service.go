package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"this-is-pun/backend/internal/entity"
	"this-is-pun/backend/internal/model"
)

const maxRoomPlayers = 8

// ErrDependencyUnavailable 表示 Redis 房间状态存储不可用。
var ErrDependencyUnavailable = errors.New("room dependency unavailable")

// RoomRepository 是房间服务所需的最小 Redis 仓储端口。
type RoomRepository interface {
	Create(context.Context, entity.Room) error
	Get(context.Context, string) (entity.Room, error)
	Update(context.Context, string, func(*entity.Room) error) (entity.Room, error)
}

// RoomService 编排房间生命周期和答题状态机。
type RoomService struct {
	repo RoomRepository
	game *GameService
}

func NewRoomService(repo RoomRepository, game *GameService) *RoomService {
	return &RoomService{repo: repo, game: game}
}

// CreateRoomInput 是创建房间请求。
type CreateRoomInput struct {
	PlayerID    string
	PlayerName  string
	PuzzleSetID int64
}

// JoinRoomInput 是加入房间请求。
type JoinRoomInput struct {
	RoomID     string
	PlayerID   string
	PlayerName string
}

// ReadyRoomInput 是准备状态请求。
type ReadyRoomInput struct{ RoomID, PlayerID string }

// StartRoomInput 是开始游戏请求。
type StartRoomInput struct{ RoomID, PlayerID string }

// AnswerRoomInput 是联机答题请求。
type AnswerRoomInput struct {
	RoomID     string
	PlayerID   string
	Answer     string
	AnswerMode model.AnswerMode
	ElapsedMS  int64
}

func (s *RoomService) Create(ctx context.Context, input CreateRoomInput) (entity.Room, error) {
	if strings.TrimSpace(input.PlayerID) == "" {
		return entity.Room{}, ErrUnauthorized
	}
	setID := input.PuzzleSetID
	if setID < 1 {
		setID = 1
	}
	puzzles, err := s.game.ListPublicPuzzlesBySet(setID)
	if err != nil {
		return entity.Room{}, err
	}
	if len(puzzles) == 0 {
		return entity.Room{}, ErrInvalidRequest
	}
	id := newRoomID()
	now := time.Now().UTC()
	room := entity.Room{ID: id, HostID: input.PlayerID, Status: entity.RoomStatusWaiting, Players: []entity.RoomPlayer{{ID: input.PlayerID, Name: fallbackName(input.PlayerName, input.PlayerID), Present: true}}, PuzzleIDs: make([]int64, 0, len(puzzles)), CurrentPuzzle: puzzles[0].ID, CreatedAt: now, UpdatedAt: now}
	for _, puzzle := range puzzles {
		room.PuzzleIDs = append(room.PuzzleIDs, puzzle.ID)
	}
	if err := s.repo.Create(ctx, room); err != nil {
		return entity.Room{}, mapRoomError(err)
	}
	return room, nil
}

func (s *RoomService) Join(ctx context.Context, input JoinRoomInput) (entity.Room, error) {
	if strings.TrimSpace(input.RoomID) == "" || strings.TrimSpace(input.PlayerID) == "" {
		return entity.Room{}, ErrInvalidRequest
	}
	room, err := s.repo.Update(ctx, input.RoomID, func(room *entity.Room) error {
		if room.Status == entity.RoomStatusFinished {
			return ErrInvalidRequest
		}
		for i := range room.Players {
			if room.Players[i].ID == input.PlayerID {
				room.Players[i].Present = true
				room.Players[i].Name = fallbackName(input.PlayerName, input.PlayerID)
				return nil
			}
		}
		if len(room.Players) >= maxRoomPlayers {
			return ErrAlreadyExists
		}
		room.Players = append(room.Players, entity.RoomPlayer{ID: input.PlayerID, Name: fallbackName(input.PlayerName, input.PlayerID), Present: true})
		return nil
	})
	if err != nil {
		return entity.Room{}, mapRoomError(err)
	}
	return room, nil
}

func (s *RoomService) Ready(ctx context.Context, input ReadyRoomInput) (entity.Room, error) {
	return s.updatePlayer(ctx, input.RoomID, input.PlayerID, func(player *entity.RoomPlayer, room *entity.Room) error {
		if room.Status != entity.RoomStatusWaiting && room.Status != entity.RoomStatusReady {
			return ErrInvalidRequest
		}
		player.Ready = true
		allReady := len(room.Players) > 0
		for _, item := range room.Players {
			if !item.Present || !item.Ready {
				allReady = false
				break
			}
		}
		if allReady {
			room.Status = entity.RoomStatusReady
		}
		return nil
	})
}

// Leave 将连接断开的玩家标记为离线，保留席位以支持短时间重连。
func (s *RoomService) Leave(ctx context.Context, roomID, playerID string) (entity.Room, error) {
	return s.updatePlayer(ctx, roomID, playerID, func(player *entity.RoomPlayer, room *entity.Room) error { player.Present = false; return nil })
}

func (s *RoomService) Start(ctx context.Context, input StartRoomInput) (entity.Room, error) {
	room, err := s.repo.Update(ctx, input.RoomID, func(room *entity.Room) error {
		if room.HostID != input.PlayerID || room.Status != entity.RoomStatusReady {
			return ErrInvalidRequest
		}
		for _, player := range room.Players {
			if !player.Ready {
				return ErrInvalidRequest
			}
		}
		room.Status = entity.RoomStatusPlaying
		return nil
	})
	if err != nil {
		return entity.Room{}, mapRoomError(err)
	}
	return room, nil
}

// Answer 校验答案并只允许每名玩家成功提交一次。
func (s *RoomService) Answer(ctx context.Context, input AnswerRoomInput) (entity.Room, CheckAnswerResult, error) {
	room, err := s.repo.Get(ctx, input.RoomID)
	if err != nil {
		return entity.Room{}, CheckAnswerResult{}, mapRoomError(err)
	}
	if room.Status != entity.RoomStatusPlaying {
		return entity.Room{}, CheckAnswerResult{}, ErrInvalidRequest
	}
	puzzle, err := s.game.GetPublicPuzzle(room.CurrentPuzzle)
	if err != nil {
		return entity.Room{}, CheckAnswerResult{}, err
	}
	result, err := s.game.CheckAnswer(room.CurrentPuzzle, CheckAnswerRequest{AttemptID: puzzle.AttemptID, Answer: input.Answer, AnswerMode: input.AnswerMode, ElapsedMS: input.ElapsedMS})
	if err != nil {
		return entity.Room{}, CheckAnswerResult{}, err
	}
	if !result.Correct {
		return room, result, nil
	}
	updated, err := s.repo.Update(ctx, input.RoomID, func(room *entity.Room) error {
		if room.Status != entity.RoomStatusPlaying {
			return ErrInvalidRequest
		}
		for i := range room.Players {
			if room.Players[i].ID == input.PlayerID {
				if room.Players[i].Answered {
					return ErrAlreadyExists
				}
				room.Players[i].Answered = true
				room.Players[i].Score += result.Score
				allAnswered := true
				for _, item := range room.Players {
					if !item.Answered {
						allAnswered = false
						break
					}
				}
				if allAnswered {
					room.Status = entity.RoomStatusFinished
				}
				return nil
			}
		}
		return ErrUnauthorized
	})
	if err != nil {
		return entity.Room{}, result, mapRoomError(err)
	}
	return updated, result, nil
}

func (s *RoomService) updatePlayer(ctx context.Context, roomID, playerID string, mutate func(*entity.RoomPlayer, *entity.Room) error) (entity.Room, error) {
	if strings.TrimSpace(roomID) == "" || strings.TrimSpace(playerID) == "" {
		return entity.Room{}, ErrInvalidRequest
	}
	room, err := s.repo.Update(ctx, roomID, func(room *entity.Room) error {
		for i := range room.Players {
			if room.Players[i].ID == playerID {
				return mutate(&room.Players[i], room)
			}
		}
		return ErrUnauthorized
	})
	if err != nil {
		return entity.Room{}, mapRoomError(err)
	}
	return room, nil
}

func mapRoomError(err error) error {
	if errors.Is(err, model.ErrRoomNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, model.ErrRoomUnavailable) {
		return ErrDependencyUnavailable
	}
	return err
}

func fallbackName(name, id string) string {
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	return "玩家" + id
}

func newRoomID() string {
	var bytes [4]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}
