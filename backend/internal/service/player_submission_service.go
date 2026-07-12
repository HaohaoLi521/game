package service

import (
	"context"
	"fmt"
	"time"

	"this-is-pun/backend/internal/entity"
	"this-is-pun/backend/internal/model"
)

// PlayerSubmissionView 是玩家端投稿状态与详情的响应 DTO。
type PlayerSubmissionView struct {
	ID             int64                  `json:"id"`
	Status         model.SubmissionStatus `json:"status"`
	ReviewNote     string                 `json:"review_note"`
	Answer         string                 `json:"answer"`
	Category       string                 `json:"category"`
	Difficulty     int                    `json:"difficulty"`
	HintImages     []model.HintImage      `json:"hint_images"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	ApprovedPuzzle *model.PuzzlePublic    `json:"approved_puzzle,omitempty"`
}

// PlayerSubmissionService 编排投稿创建、归属持久化和玩家权限校验。
type PlayerSubmissionService struct {
	repo *model.PlayerSubmissionRepository
	game *GameService
}

func NewPlayerSubmissionService(repo *model.PlayerSubmissionRepository, game *GameService) *PlayerSubmissionService {
	return &PlayerSubmissionService{repo: repo, game: game}
}

// Create 先沿用既有投稿业务，再将结果关联到当前登录玩家。
func (s *PlayerSubmissionService) Create(ctx context.Context, playerID uint64, input SubmissionInput) (PlayerSubmissionView, error) {
	if playerID == 0 {
		return PlayerSubmissionView{}, ErrUnauthorized
	}
	submission, err := s.game.CreateSubmission(input)
	if err != nil {
		return PlayerSubmissionView{}, err
	}
	if err := s.repo.Create(ctx, &entity.PlayerSubmission{PlayerID: playerID, SubmissionID: submission.ID}); err != nil {
		return PlayerSubmissionView{}, fmt.Errorf("save player submission: %w", err)
	}
	return toPlayerSubmissionView(submission), nil
}

// List 只返回当前玩家名下投稿，状态由既有投稿记录作为唯一事实来源。
func (s *PlayerSubmissionService) List(ctx context.Context, playerID uint64) ([]PlayerSubmissionView, error) {
	if playerID == 0 {
		return nil, ErrUnauthorized
	}
	relations, err := s.repo.List(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player submissions: %w", err)
	}
	items := make([]PlayerSubmissionView, 0, len(relations))
	for _, relation := range relations {
		submission, err := s.game.repo.GetSubmission(relation.SubmissionID)
		if err != nil {
			return nil, convertRepoError(err)
		}
		items = append(items, toPlayerSubmissionView(submission))
	}
	return items, nil
}

// Get 校验投稿归属后读取详情。
func (s *PlayerSubmissionService) Get(ctx context.Context, playerID uint64, submissionID int64) (PlayerSubmissionView, error) {
	if playerID == 0 || submissionID < 1 {
		return PlayerSubmissionView{}, ErrInvalidRequest
	}
	relation, err := s.repo.Get(ctx, playerID, submissionID)
	if err != nil {
		return PlayerSubmissionView{}, fmt.Errorf("find player submission: %w", err)
	}
	if relation == nil {
		return PlayerSubmissionView{}, ErrNotFound
	}
	submission, err := s.game.repo.GetSubmission(relation.SubmissionID)
	if err != nil {
		return PlayerSubmissionView{}, convertRepoError(err)
	}
	return toPlayerSubmissionView(submission), nil
}

func toPlayerSubmissionView(submission model.PuzzleSubmission) PlayerSubmissionView {
	return PlayerSubmissionView{ID: submission.ID, Status: submission.Status, ReviewNote: submission.ReviewNote, Answer: submission.Answer, Category: submission.Category, Difficulty: submission.Difficulty, HintImages: submission.HintImages, CreatedAt: submission.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: submission.UpdatedAt.UTC().Format(time.RFC3339), ApprovedPuzzle: submission.ApprovedPuzzle}
}
