package service

import (
	"context"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"this-is-pun/backend/internal/auth"
	"this-is-pun/backend/internal/entity"
	"this-is-pun/backend/internal/model"
)

// PlayerService 处理玩家注册、登录和服务端进度。
type PlayerService struct {
	repo *model.PlayerRepository
	auth *auth.AuthManager
}

func NewPlayerService(repo *model.PlayerRepository, auth *auth.AuthManager) *PlayerService {
	return &PlayerService{repo: repo, auth: auth}
}
func (s *PlayerService) Register(ctx context.Context, username, password, ip, agent string) (*auth.LoginResult, *entity.Player, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if len(username) < 3 || len(password) < 6 {
		return nil, nil, ErrInvalidRequest
	}
	p, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	if p != nil {
		return nil, nil, ErrAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	p = &entity.Player{Username: username, PasswordHash: string(hash)}
	if err = s.repo.Create(ctx, p); err != nil {
		return nil, nil, err
	}
	token, err := s.auth.Login(strconv.FormatUint(p.ID, 10), ip, agent, nil)
	return token, p, err
}
func (s *PlayerService) Login(ctx context.Context, username, password, ip, agent string) (*auth.LoginResult, *entity.Player, error) {
	p, err := s.repo.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil || p == nil || bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password)) != nil {
		return nil, nil, ErrUnauthorized
	}
	token, err := s.auth.Login(strconv.FormatUint(p.ID, 10), ip, agent, nil)
	return token, p, err
}

// Progress 读取玩家已通关题目。
func (s *PlayerService) Progress(ctx context.Context, playerID uint64) ([]entity.PlayerProgress, error) {
	return s.repo.ListProgress(ctx, playerID)
}

// MarkSolved 保存一题通关状态。
func (s *PlayerService) MarkSolved(ctx context.Context, playerID uint64, puzzleID int64) error {
	if puzzleID < 1 {
		return ErrInvalidRequest
	}
	return s.repo.MarkSolved(ctx, playerID, puzzleID)
}
