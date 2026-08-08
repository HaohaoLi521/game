package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrRateLimitStoreUnavailable = errors.New("rate limit store unavailable")

// RateLimitRule 描述一个接口限流窗口。
type RateLimitRule struct {
	Limit  int
	Window time.Duration
}

// RateLimitDecision 是限流中间件所需的最小决策结果。
type RateLimitDecision struct {
	Limit      int
	Allowed    bool
	Remaining  int
	RetryAfter int
}

// RateLimitQuery 是限流服务依赖的最小仓储接口。
type RateLimitQuery interface {
	Consume(context.Context, string, string, time.Duration) (int, error)
}

// RateLimitService 负责按业务范围执行 Redis 固定窗口限流。
type RateLimitService struct {
	repo  RateLimitQuery
	rules map[string]RateLimitRule
}

// NewRateLimitService 创建限流服务并复制规则，避免运行时被外部 map 修改。
func NewRateLimitService(repo RateLimitQuery, rules map[string]RateLimitRule) *RateLimitService {
	copied := make(map[string]RateLimitRule, len(rules))
	for name, rule := range rules {
		copied[name] = rule
	}
	return &RateLimitService{repo: repo, rules: copied}
}

// DefaultRateLimitRules 返回默认的生产保护阈值。
func DefaultRateLimitRules() map[string]RateLimitRule {
	return map[string]RateLimitRule{
		"player-auth":   {Limit: 10, Window: time.Minute},
		"puzzle-answer": {Limit: 60, Window: time.Minute},
		"room-action":   {Limit: 30, Window: time.Minute},
	}
}

// Allow 判断当前主体是否还能访问指定业务范围。
func (s *RateLimitService) Allow(ctx context.Context, scope, subject string) (RateLimitDecision, error) {
	rule, ok := s.rules[scope]
	if !ok || rule.Limit <= 0 || rule.Window <= 0 {
		return RateLimitDecision{Allowed: true}, nil
	}
	if s.repo == nil {
		return RateLimitDecision{}, ErrRateLimitStoreUnavailable
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "anonymous"
	}
	// 将 IP、用户 ID 等主体做摘要，避免 Redis key 泄露原始标识并控制 key 长度。
	digest := sha256.Sum256([]byte(subject))
	count, err := s.repo.Consume(ctx, scope, hex.EncodeToString(digest[:]), rule.Window)
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("allow %s: %w", scope, err)
	}
	remaining := rule.Limit - count
	if remaining < 0 {
		remaining = 0
	}
	decision := RateLimitDecision{Limit: rule.Limit, Allowed: count <= rule.Limit, Remaining: remaining, RetryAfter: int(rule.Window.Seconds())}
	return decision, nil
}
