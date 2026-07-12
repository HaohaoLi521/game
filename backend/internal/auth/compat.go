package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	"go.uber.org/zap"
)

// authLogger 兼容旧认证代码的日志调用，避免引入外部基础库。
type authLogger struct{}

func (authLogger) Debug(message string, _ ...zap.Field) { slog.Debug(message) }
func (authLogger) Info(message string, _ ...zap.Field)  { slog.Info(message) }
func (authLogger) Warn(message string, _ ...zap.Field)  { slog.Warn(message) }
func (authLogger) Error(message string, _ ...zap.Field) { slog.Error(message) }

var logger authLogger

// newID 为 JWT 标识和会话标识生成随机值。
func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes[:])
}
