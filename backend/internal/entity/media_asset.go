package entity

import "time"

// MediaAsset 是上传文件的元数据实体，二进制内容由 MinIO 保存。
type MediaAsset struct {
	ID          uint64    `gorm:"primaryKey"`
	OwnerID     uint64    `gorm:"index;not null"`
	ObjectKey   string    `gorm:"size:255;uniqueIndex;not null"`
	PublicURL   string    `gorm:"size:1024;not null"`
	ContentType string    `gorm:"size:128;not null"`
	Size        int64     `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (MediaAsset) TableName() string { return "media_assets" }
