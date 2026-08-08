package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"this-is-pun/backend/internal/entity"
	"this-is-pun/backend/internal/model"
)

// WorkshopQuery 是创意工坊查询所需的最小仓储端口。
type WorkshopQuery interface {
	List(context.Context, model.WorkshopFilter) ([]entity.WorkshopPuzzle, int64, error)
}

// WorkshopItem 是工坊卡片响应 DTO，不暴露 GORM 投影。
type WorkshopItem struct {
	ID          int64             `json:"id"`
	PuzzleSetID int64             `json:"puzzle_set_id"`
	AuthorName  string            `json:"author_name"`
	HintImages  []model.HintImage `json:"hint_images"`
	Category    string            `json:"category"`
	Difficulty  int               `json:"difficulty"`
}

// WorkshopPage 是工坊分页响应 DTO。
type WorkshopPage struct {
	Items    []WorkshopItem `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Total    int64          `json:"total"`
}

// WorkshopService 负责工坊过滤、分页和响应映射。
type WorkshopService struct{ repo WorkshopQuery }

func NewWorkshopService(repo WorkshopQuery) *WorkshopService { return &WorkshopService{repo: repo} }

func (s *WorkshopService) List(ctx context.Context, category string, difficulty, page, pageSize int) (WorkshopPage, error) {
	category = strings.TrimSpace(category)
	if page < 1 || pageSize < 1 || pageSize > 50 || difficulty < 0 || difficulty > 5 {
		return WorkshopPage{}, ErrInvalidRequest
	}
	rows, total, err := s.repo.List(ctx, model.WorkshopFilter{Category: category, Difficulty: difficulty, Limit: pageSize, Offset: (page - 1) * pageSize})
	if err != nil {
		return WorkshopPage{}, fmt.Errorf("list workshop: %w", err)
	}
	items := make([]WorkshopItem, 0, len(rows))
	for _, row := range rows {
		var hints []model.HintImage
		if len(row.HintImages) > 0 && string(row.HintImages) != "null" {
			if err := json.Unmarshal(row.HintImages, &hints); err != nil {
				return WorkshopPage{}, fmt.Errorf("decode workshop hints: %w", err)
			}
		}
		items = append(items, WorkshopItem{ID: row.ID, PuzzleSetID: row.PuzzleSetID, AuthorName: row.AuthorName, HintImages: hints, Category: row.Category, Difficulty: row.Difficulty})
	}
	return WorkshopPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}
