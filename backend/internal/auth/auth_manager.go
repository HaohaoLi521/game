package auth

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AuthManager 认证管理器
type AuthManager struct {
	*JWTManager
	*SSOManager
}

// NewAuthManager 创建认证管理器
func NewAuthManager(rdb redis.Cmdable, jwtConfig JWTConfig, ssoConfig SSOConfig) *AuthManager {
	return &AuthManager{
		JWTManager: NewJWTManager(jwtConfig),
		SSOManager: NewSSOManager(rdb, ssoConfig),
	}
}

// Login 通用登录逻辑
func (a *AuthManager) Login(userID string, clientIP, userAgent string, claimsCustomizer func(jwt.MapClaims)) (*LoginResult, error) {
	// 1. 生成Token对
	accessToken, err := a.GenerateAccessToken(userID, claimsCustomizer)
	if err != nil {
		return nil, fmt.Errorf("生成访问Token失败: %v", err)
	}

	refreshToken, err := a.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成刷新Token失败: %v", err)
	}

	// 2. 创建会话
	sessionID := newID()
	deviceID := a.generateDeviceID(clientIP, userAgent)
	tokenHash := HashToken(accessToken)

	session := &Session{
		UserID:       userID,
		LastActivity: time.Now(),
		IPAddress:    clientIP,
		UserAgent:    userAgent,
		DeviceID:     deviceID,
		SessionID:    sessionID,
		TokenHash:    tokenHash,
		RefreshToken: refreshToken,
	}

	if err := a.CreateSession(session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(a.JWTManager.config.AccessTokenExpire / time.Second),
		SessionID:    sessionID,
		DeviceID:     deviceID,
	}, nil
}

// ValidateToken 验证Token, has 故障降级
func (a *AuthManager) ValidateToken(tokenString string) (*SessionValidationResult, error) {
	if tokenString == "" {
		return &SessionValidationResult{
			IsValid: false,
			Error:   "Token不能为空",
		}, nil
	}

	// 1. 解析 JWT，校验签名和过期时间
	claims, err := a.ParseToken(tokenString)
	if err != nil {
		return &SessionValidationResult{
			IsValid: false,
			Error:   "Token无效: " + err.Error(),
		}, nil
	}

	// 2. 检查黑名单
	tokenHash := HashToken(tokenString)

	// 2.1 检查黑名单
	blacklistKey := SSOBlacklistPrefix + tokenHash
	exist, err := a.rdb.Exists(a.ctx, blacklistKey).Result()
	if err != nil {
		// Redis 查询失败
		if a.SSOManager.config.EnableDegradation && isRedisError(err) {
			logger.Error("Redis故障（黑名单检查），启用降级策略", zap.Error(err))
			// 降级：无法检查黑名单，但允许通过
			return a.buildDegradedResult(claims), nil
		}
		// 如果未启用降级，或者不是 Redis 连接错误，则拒绝
		return &SessionValidationResult{
			IsValid: false,
			Error:   "验证失败: " + err.Error(),
		}, nil
	}
	if exist > 0 {
		return &SessionValidationResult{
			IsValid: false,
			Error:   "Token已被撤销",
		}, nil
	}

	// 2.2 检查会话
	session, err := a.GetSessionByTokenHash(tokenHash)
	if err != nil {
		// Redis 查询失败
		if a.SSOManager.config.EnableDegradation && isRedisError(err) {
			logger.Error("Redis故障（Session检查），启用降级策略", zap.Error(err))
			// 降级：无法检查 Session，但允许通过
			return a.buildDegradedResult(claims), nil
		}
		return &SessionValidationResult{
			IsValid: false,
			Error:   "获取会话失败",
		}, nil
	}
	if session == nil {
		a.AddTokenToBlacklist(tokenString)
		return &SessionValidationResult{
			IsValid: false,
			Error:   "会话不存在或已过期",
		}, nil
	}

	// 3. 验证用户ID一致性
	tokenUserID, ok := claims["aud"].(string)
	if !ok || tokenUserID != session.UserID {
		a.AddTokenToBlacklist(tokenString)
		return &SessionValidationResult{
			IsValid: false,
			Error:   "Token用户不匹配",
		}, nil
	}

	// 4. 异步更新活动时间
	go func() {
		_ = a.UpdateSessionActivity(session.SessionID)
	}()

	return &SessionValidationResult{
		IsValid: true,
		Session: session,
		Claims:  claims,
	}, nil
}

// buildDegradedResult 构造降级验证结果（仅基于 JWT Claims）
func (a *AuthManager) buildDegradedResult(claims jwt.MapClaims) *SessionValidationResult {
	userID, _ := claims["aud"].(string)
	return &SessionValidationResult{
		IsValid: true,
		Claims:  claims,
		Session: &Session{
			UserID: userID,
			// 其他字段无法从 JWT 恢复，保持零值
		},
	}
}

// isRedisError 判断是否为 Redis 连接/网络错误
func isRedisError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 常见的 Redis 连接错误关键词
	return strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "i/o timeout") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "no route to host") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "EOF")
}

// RefreshToken 刷新Token
func (a *AuthManager) RefreshToken(refreshTokenString, clientIP, userAgent string, claimsCustomizer func(jwt.MapClaims)) (*LoginResult, error) {
	// 1. 查找刷新Token对应的会话
	refreshKey := SSORefreshTokenPrefix + refreshTokenString
	sessionID, err := a.rdb.Get(a.ctx, refreshKey).Result()
	if err != nil || sessionID == "" {
		return nil, errors.New("刷新Token无效或已过期")
	}

	session, err := a.GetSession(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("会话不存在")
	}

	if session.RefreshToken != refreshTokenString {
		return nil, errors.New("刷新Token不匹配")
	}

	// 2. 生成新Token
	newAccessToken, err := a.GenerateAccessToken(session.UserID, claimsCustomizer)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := a.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 3. 更新会话
	oldTokenHash := session.TokenHash
	oldRefreshToken := session.RefreshToken // 其实就是输入参数

	// 将旧Token加入黑名单
	if oldTokenHash != "" {
		a.AddTokenToBlacklistByHash(oldTokenHash)
	}

	// 更新Redis映射
	newTokenHash := HashToken(newAccessToken)

	// 更新内存对象
	session.TokenHash = newTokenHash
	session.RefreshToken = newRefreshToken
	session.LastActivity = time.Now()
	session.IPAddress = clientIP
	session.UserAgent = userAgent

	// 这里需要单独实现一个UpdateSessionAfterRefresh，或者手动操作Pipeline
	// 为了复用代码，我们在 SSOManager 中添加 UpdateSessionAfterRefresh
	if err := a.SSOManager.updateSessionAfterRefresh(session, oldTokenHash, oldRefreshToken, newTokenHash, newRefreshToken); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(a.JWTManager.config.AccessTokenExpire / time.Second),
		SessionID:    session.SessionID,
		DeviceID:     session.DeviceID,
	}, nil
}

// Logout 登出
func (a *AuthManager) Logout(tokenString, sessionID string) error {
	if tokenString != "" {
		a.AddTokenToBlacklist(tokenString)
		if sessionID == "" {
			tokenHash := HashToken(tokenString)
			if s, err := a.GetSessionByTokenHash(tokenHash); err == nil && s != nil {
				sessionID = s.SessionID
			}
		}
	}

	if sessionID != "" {
		return a.DeleteSession(sessionID, "用户指定登出")
	}
	return nil
}

// AddTokenToBlacklist token加入黑名单
func (a *AuthManager) AddTokenToBlacklist(tokenString string) {
	if tokenString == "" {
		return
	}
	hash := HashToken(tokenString)
	a.AddTokenToBlacklistByHash(hash)
}

// AddTokenToBlacklistByHash hash加入黑名单
func (a *AuthManager) AddTokenToBlacklistByHash(tokenHash string) {
	key := SSOBlacklistPrefix + tokenHash
	// 使用访问Token的时间作为黑名单时间
	expiration := a.JWTManager.config.AccessTokenExpire
	if expiration <= 0 {
		expiration = time.Hour * 72 // 保底
	}

	if err := a.rdb.Set(a.ctx, key, "1", expiration).Err(); err != nil {
		logger.Error("添加黑名单失败", zap.Error(err))
	}
}

// generateDeviceID 生成设备及
func (a *AuthManager) generateDeviceID(clientIP, userAgent string) string {
	data := fmt.Sprintf("%s:%s", clientIP, userAgent)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// LogoutAllDevicesByUserID 根据用户ID登出所有设备
func (a *AuthManager) LogoutAllDevicesByUserID(userID, reason string) error {
	return a.SSOManager.ClearUserAllSessions(userID, reason)
}
