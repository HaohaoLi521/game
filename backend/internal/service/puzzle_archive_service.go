package service

import (
	"context"
	"fmt"
	"time"

	"this-is-pun/backend/internal/model"
)

// PuzzleArchiveService 编排管理端题目归档查询和恢复。
// PuzzleArchiveView 是管理端归档题目的响应 DTO。
type PuzzleArchiveView struct {
	ID          int64      `json:"id"`
	PuzzleSetID int64      `json:"puzzle_set_id"`
	Answer      string     `json:"answer"`
	Category    string     `json:"category"`
	ArchivedAt  *time.Time `json:"archived_at"`
}

type PuzzleArchiveService struct {
	repo *model.PuzzleArchiveRepository
}

func NewPuzzleArchiveService(repo *model.PuzzleArchiveRepository) *PuzzleArchiveService {
	return &PuzzleArchiveService{repo: repo}
}

func (s *PuzzleArchiveService) ListArchived(ctx context.Context) ([]PuzzleArchiveView, error) {
	items, err := s.repo.ListArchived(ctx)
	if err != nil {
		return nil, fmt.Errorf("list archived puzzles: %w", err)
	}
	result := make([]PuzzleArchiveView, 0, len(items))
	for _, item := range items {
		result = append(result, PuzzleArchiveView{ID: item.ID, PuzzleSetID: item.PuzzleSetID, Answer: item.Answer, Category: item.Category, ArchivedAt: item.ArchivedAt})
	}
	return result, nil
}

func (s *PuzzleArchiveService) Restore(ctx context.Context, id int64) error {
	if id < 1 {
		return ErrInvalidRequest
	}
	if err := s.repo.Restore(ctx, id); err != nil {
		return fmt.Errorf("restore puzzle: %w", err)
	}
	return nil
}
