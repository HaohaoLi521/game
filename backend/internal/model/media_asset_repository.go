package model

import (
	"context"
	"gorm.io/gorm"
	"this-is-pun/backend/internal/entity"
)

// MediaAssetRepository 负责上传资产元数据持久化。
type MediaAssetRepository struct{ db *gorm.DB }

func NewMediaAssetRepository(db *gorm.DB) (*MediaAssetRepository, error) {
	if err := db.AutoMigrate(&entity.MediaAsset{}); err != nil {
		return nil, err
	}
	return &MediaAssetRepository{db: db}, nil
}
func (r *MediaAssetRepository) Create(ctx context.Context, asset *entity.MediaAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}
