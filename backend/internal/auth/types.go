package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Session 多设备SSO会话信息
type Session struct {
	UserID       string    `json:"user_id"`
	LastActivity time.Time `json:"last_activity"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceID     string    `json:"device_id"`
	SessionID    string    `json:"session_id"`
	TokenHash    string    `json:"token_hash"`    // 当前Token的哈希值(用于精确匹配)
	RefreshToken string    `json:"refresh_token"` // 关联的刷新Token
	ExpiresAt    time.Time `json:"expires_at"`    // 会话过期时间
}

// LoginResult 登录结果
type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	SessionID    string `json:"session_id"`
	DeviceID     string `json:"device_id"`
}

// UserSessionList 用户会话列表
type UserSessionList struct {
	UserID   string    `json:"user_id"`
	Sessions []Session `json:"sessions"`
	Count    int       `json:"count"`
}

// SessionValidationResult 会话验证结果
type SessionValidationResult struct {
	IsValid bool          `json:"is_valid"`
	Session *Session      `json:"session"`
	Claims  jwt.MapClaims `json:"claims"`
	Error   string        `json:"error,omitempty"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string        // JWT密钥
	AccessTokenExpire  time.Duration // 访问Token过期时间
	RefreshTokenExpire time.Duration // 刷新Token过期时间
	AppID              string        // 应用ID
	DomainName         string        // 域名
}

// SSOConfig SSO配置
type SSOConfig struct {
	EnableSingleLogin bool // 是否启用单点登录
	SessionExpireTime int  // 会话过期时间(小时)
	EnableAsyncUpdate bool // 是否启用异步更新
	BatchSize         int  // 批量操作大小
	EnableDegradation bool // 是否启用 Redis 故障降级（降级为无状态 JWT 验证）
}
