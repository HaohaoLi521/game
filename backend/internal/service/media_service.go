package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"this-is-pun/backend/internal/entity"
	"this-is-pun/backend/internal/model"
)

// MediaService 负责文件校验、MinIO 对象写入和资产元数据保存。
type MediaService struct {
	client *minio.Client
	bucket string
	repo   *model.MediaAssetRepository
}

func NewMediaService(client *minio.Client, bucket string, repo *model.MediaAssetRepository) *MediaService {
	return &MediaService{client: client, bucket: bucket, repo: repo}
}
func (s *MediaService) Upload(ctx context.Context, ownerID uint64, name, contentType string, size int64, body io.Reader) (*entity.MediaAsset, error) {
	if size <= 0 || size > 5<<20 || !strings.HasPrefix(contentType, "image/") {
		return nil, ErrInvalidRequest
	}
	key := fmt.Sprintf("uploads/%d/%s%s", ownerID, randomID(), filepath.Ext(name))
	if _, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}
	asset := &entity.MediaAsset{OwnerID: ownerID, ObjectKey: key, PublicURL: "/media/" + key, ContentType: contentType, Size: size}
	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, fmt.Errorf("save media asset: %w", err)
	}
	return asset, nil
}

// Open 读取私有 bucket 中的对象，由应用统一对外提供媒体 URL。
func (s *MediaService) Open(ctx context.Context, key string) (io.Reader, minio.ObjectInfo, error) {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if key == "" || strings.Contains(key, "..") {
		return nil, minio.ObjectInfo{}, ErrInvalidRequest
	}
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, fmt.Errorf("get object: %w", err)
	}
	info, err := object.Stat()
	if err != nil {
		_ = object.Close()
		return nil, minio.ObjectInfo{}, fmt.Errorf("stat object: %w", err)
	}
	return object, info, nil
}
