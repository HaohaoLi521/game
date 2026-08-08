package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrRateLimitUnavailable = errors.New("rate limit store is unavailable")

// RateLimitRepository 使用 Redis 原子计数实现固定窗口限流。
type RateLimitRepository struct {
	client *redis.Client
}

// NewRateLimitRepository 创建限流仓储。
func NewRateLimitRepository(client *redis.Client) *RateLimitRepository {
	return &RateLimitRepository{client: client}
}

// Consume 增加窗口内计数并返回当前计数。
func (r *RateLimitRepository) Consume(ctx context.Context, scope, subject string, window time.Duration) (int, error) {
	if r == nil || r.client == nil {
		return 0, ErrRateLimitUnavailable
	}
	if window <= 0 {
		return 0, fmt.Errorf("rate limit window must be positive")
	}

	// INCR 与首次设置过期时间必须在同一段 Lua 中执行，避免并发请求留下永久 key。
	const script = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return count
`
	key := "this-is-pun:rate:" + scope + ":" + subject
	count, err := r.client.Eval(ctx, script, []string{key}, window.Milliseconds()).Int()
	if err != nil {
		return 0, fmt.Errorf("consume rate limit: %w", err)
	}
	return count, nil
}
