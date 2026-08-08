package model

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"this-is-pun/backend/internal/entity"
)

// WorkshopFilter 是工坊查询条件，分页边界由 service 层负责校验。
type WorkshopFilter struct {
	Category   string
	Difficulty int
	Limit      int
	Offset     int
}

// WorkshopRepository 负责创意工坊题目投影查询。
type WorkshopRepository struct{ db *gorm.DB }

func NewWorkshopRepository(db *gorm.DB) *WorkshopRepository { return &WorkshopRepository{db: db} }

// List 查询未下架题目，并返回总数供前端分页。
func (r *WorkshopRepository) List(ctx context.Context, filter WorkshopFilter) ([]entity.WorkshopPuzzle, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.WorkshopPuzzle{}).Where("archived_at IS NULL")
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Difficulty > 0 {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count workshop puzzles: %w", err)
	}
	var items []entity.WorkshopPuzzle
	if err := query.Select("id, puzzle_set_id, author_name, hint_images, category, difficulty, sort_order").Order("sort_order, id").Limit(filter.Limit).Offset(filter.Offset).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list workshop puzzles: %w", err)
	}
	return items, total, nil
}
