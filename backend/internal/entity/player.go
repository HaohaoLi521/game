package entity

import "time"

// Player 是玩家账户的持久化实体，密码哈希不会直接暴露给接口响应。
type Player struct {
	ID           uint64    `gorm:"primaryKey"`
	Username     string    `gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// TableName 固定玩家账户表名。
func (Player) TableName() string { return "players" }

// PlayerProgress 是玩家对单题的服务端进度记录。
type PlayerProgress struct {
	ID        uint64    `gorm:"primaryKey"`
	PlayerID  uint64    `gorm:"not null;uniqueIndex:idx_player_puzzle"`
	PuzzleID  int64     `gorm:"not null;uniqueIndex:idx_player_puzzle"`
	SolvedAt  time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName 固定玩家进度表名。
func (PlayerProgress) TableName() string { return "player_progress" }
