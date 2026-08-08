package entity

import "time"

// PuzzleArchive 是既有 puzzles 表的管理端归档投影，不负责创建新表。
type PuzzleArchive struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	PuzzleSetID int64      `gorm:"column:puzzle_set_id"`
	Answer      string     `gorm:"column:answer"`
	Category    string     `gorm:"column:category"`
	ArchivedAt  *time.Time `gorm:"column:archived_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (PuzzleArchive) TableName() string { return "puzzles" }
