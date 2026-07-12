package model

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"this-is-pun/backend/internal/entity"
)

// PlayerSubmissionRepository 负责玩家投稿归属关系的 GORM 访问。
type PlayerSubmissionRepository struct{ db *gorm.DB }

// NewPlayerSubmissionRepository 使用 entity 自动迁移投稿关联表。
func NewPlayerSubmissionRepository(db *gorm.DB) (*PlayerSubmissionRepository, error) {
	if err := db.AutoMigrate(&entity.PlayerSubmission{}); err != nil {
		return nil, err
	}
	return &PlayerSubmissionRepository{db: db}, nil
}

// Create 保存一条玩家投稿归属关系。
func (r *PlayerSubmissionRepository) Create(ctx context.Context, relation *entity.PlayerSubmission) error {
	return r.db.WithContext(ctx).Create(relation).Error
}

// List 返回玩家最新投稿在前的归属关系。
func (r *PlayerSubmissionRepository) List(ctx context.Context, playerID uint64) ([]entity.PlayerSubmission, error) {
	var relations []entity.PlayerSubmission
	err := r.db.WithContext(ctx).Where("player_id = ?", playerID).Order("submission_id DESC").Find(&relations).Error
	return relations, err
}

// Get 仅在投稿归属于该玩家时返回关联记录，防止越权读取投稿详情。
func (r *PlayerSubmissionRepository) Get(ctx context.Context, playerID uint64, submissionID int64) (*entity.PlayerSubmission, error) {
	var relation entity.PlayerSubmission
	err := r.db.WithContext(ctx).Where("player_id = ? AND submission_id = ?", playerID, submissionID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &relation, err
}
