package model

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"this-is-pun/backend/internal/entity"
)

// PlayerRepository 负责玩家账户与进度的 GORM 数据访问。
type PlayerRepository struct{ db *gorm.DB }

// NewPlayerRepository 初始化玩家表结构，表结构仅由 entity 自动迁移维护。
func NewPlayerRepository(db *gorm.DB) (*PlayerRepository, error) {
	if err := db.AutoMigrate(&entity.Player{}, &entity.PlayerProgress{}); err != nil {
		return nil, err
	}
	return &PlayerRepository{db: db}, nil
}

// Create 创建玩家账户。
func (r *PlayerRepository) Create(ctx context.Context, player *entity.Player) error {
	return r.db.WithContext(ctx).Create(player).Error
}

// FindByUsername 按用户名查询玩家。
func (r *PlayerRepository) FindByUsername(ctx context.Context, username string) (*entity.Player, error) {
	var player entity.Player
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&player).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &player, err
}
