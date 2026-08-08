package model

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ReadinessRepository 负责探测运行时依赖，不执行迁移或业务写入。
type ReadinessRepository struct {
	db     *gorm.DB
	redis  *redis.Client
	media  *minio.Client
	bucket string
}

// NewReadinessRepository 创建运行时依赖探针仓储。
func NewReadinessRepository(db *gorm.DB, redisClient *redis.Client, media *minio.Client, bucket string) *ReadinessRepository {
	return &ReadinessRepository{db: db, redis: redisClient, media: media, bucket: bucket}
}

// Check 返回每个依赖的错误；nil 表示该依赖检查通过。
func (r *ReadinessRepository) Check(ctx context.Context) map[string]error {
	result := map[string]error{"postgres": errors.New("postgres is not configured"), "redis": errors.New("redis is not configured"), "minio": errors.New("minio is not configured")}
	if r.db != nil {
		result["postgres"] = r.db.WithContext(ctx).Exec("SELECT 1").Error
	}
	if r.redis != nil {
		result["redis"] = r.redis.Ping(ctx).Err()
	}
	if r.media != nil && r.bucket != "" {
		exists, err := r.media.BucketExists(ctx, r.bucket)
		if err == nil && !exists {
			err = errors.New("bucket does not exist")
		}
		result["minio"] = err
	}
	return result
}
