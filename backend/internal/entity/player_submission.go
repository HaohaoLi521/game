package entity

import "time"

// PlayerSubmission 记录已登录玩家与旧投稿记录之间的归属关系。
type PlayerSubmission struct {
	ID           uint64    `gorm:"primaryKey"`
	PlayerID     uint64    `gorm:"not null;index:idx_player_submission_player_id"`
	SubmissionID int64     `gorm:"not null;uniqueIndex"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// TableName 固定玩家投稿关联表名。
func (PlayerSubmission) TableName() string { return "player_submissions" }
