package model

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"this-is-pun/backend/internal/entity"
)

// PuzzleArchiveRepository 负责既有 puzzles 表的归档查询和恢复。
type PuzzleArchiveRepository struct{ db *gorm.DB }

func NewPuzzleArchiveRepository(db *gorm.DB) *PuzzleArchiveRepository {
	return &PuzzleArchiveRepository{db: db}
}

func (r *PuzzleArchiveRepository) ListArchived(ctx context.Context) ([]entity.PuzzleArchive, error) {
	var items []entity.PuzzleArchive
	err := r.db.WithContext(ctx).Where("archived_at IS NOT NULL").Order("archived_at DESC, id DESC").Find(&items).Error
	return items, err
}

func (r *PuzzleArchiveRepository) Restore(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&entity.PuzzleArchive{}).Where("id = ? AND archived_at IS NOT NULL", id).Updates(map[string]any{"archived_at": nil, "updated_at": gorm.Expr("NOW()")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("puzzle not found or already active")
	}
	return nil
}
