package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis键前缀常量
var (
	SSOKeyNamespacePrefix = ""
	SSOGlobalPrefix       = "sso:"
	// 用户会话集合键: sso:user:{user_id} -> Set{session_id1, session_id2, ...}
	SSOUserSessionsPrefix = "sso:user:"
	// 会话详情键: sso:session:{session_id} -> session_info
	SSOSessionPrefix = "sso:session:"
	// 设备会话映射: sso:device:{user_id}:{device_id} -> session_id
	SSODeviceSessionPrefix = "sso:device:"
	// Token会话映射: sso:token:{token_hash} -> session_id（用于精确匹配）
	SSOTokenSessionPrefix = "sso:token:"
	// Token黑名单键: sso:blacklist:{token_hash} -> "1"
	SSOBlacklistPrefix = "sso:blacklist:"
	// 刷新Token键: sso:refresh:{refresh_token} -> session_id
	SSORefreshTokenPrefix = "sso:refresh:"
	// 用户刷新Token集合: sso:user_refresh:{user_id} -> Set{refresh_token1, refresh_token2, ...}
	SSOUserRefreshPrefix = "sso:user_refresh:"
	// 会话Token映射: sso:session_token:{session_id} -> token_hash
	SSOSessionTokenPrefix = "sso:session_token:"
)

func SetRedisKeyNamespacePrefix(prefix string) {
	ns := strings.TrimSpace(prefix)
	if ns != "" && !strings.HasSuffix(ns, ":") {
		ns += ":"
	}
	SSOKeyNamespacePrefix = ns
	SSOGlobalPrefix = ns + "sso:"
	SSOUserSessionsPrefix = SSOGlobalPrefix + "user:"
	SSOSessionPrefix = SSOGlobalPrefix + "session:"
	SSODeviceSessionPrefix = SSOGlobalPrefix + "device:"
	SSOTokenSessionPrefix = SSOGlobalPrefix + "token:"
	SSOBlacklistPrefix = SSOGlobalPrefix + "blacklist:"
	SSORefreshTokenPrefix = SSOGlobalPrefix + "refresh:"
	SSOUserRefreshPrefix = SSOGlobalPrefix + "user_refresh:"
	SSOSessionTokenPrefix = SSOGlobalPrefix + "session_token:"
}

// SSOManager SSO管理器
type SSOManager struct {
	ctx       context.Context
	batchSize int
	rdb       redis.Cmdable
	config    SSOConfig
}

// NewSSOManager 创建SSO管理器
func NewSSOManager(rdb redis.Cmdable, config SSOConfig) *SSOManager {
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}
	if config.SessionExpireTime <= 0 {
		config.SessionExpireTime = 72 // 默认72小时
	}

	return &SSOManager{
		ctx:       context.Background(),
		batchSize: config.BatchSize,
		rdb:       rdb,
		config:    config,
	}
}

// WithContext 返回一个使用指定 context 的 SSOManager 副本, fix
// 用于需要超时控制的场景，例如：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
//	defer cancel()
//	session, err := ssoManager.WithContext(ctx).GetSession(sessionID)
func (m *SSOManager) WithContext(ctx context.Context) *SSOManager {
	return &SSOManager{
		ctx:       ctx,
		batchSize: m.batchSize,
		rdb:       m.rdb,
		config:    m.config,
	}
}

// CreateSession 创建用户会话
func (m *SSOManager) CreateSession(session *Session) error {
	if session == nil {
		return errors.New("会话信息不能为空")
	}

	// 验证必要字段
	if session.UserID == "" || session.SessionID == "" || session.DeviceID == "" {
		return errors.New("用户ID、会话ID和设备ID不能为空")
	}

	// 1. 处理单点登录逻辑
	if m.config.EnableSingleLogin {
		// 单点登录：清除用户所有旧会话
		if err := m.ClearUserAllSessions(session.UserID, "单点登录模式"); err != nil {
			logger.Warn("清除用户旧会话失败", zap.String("user_id", session.UserID), zap.Error(err))
		}
	} else {
		// 多设备登录：检查是否有会话过期，如果过期则删除该会话及相关的映射
		go func() {
			sessionIDs, err := m.rdb.SMembers(m.ctx, SSOUserSessionsPrefix+session.UserID).Result()
			if err != nil {
				logger.Warn("获取用户会话列表失败", zap.String("user_id", session.UserID), zap.Error(err))
				return
			}
			if len(sessionIDs) > 0 {
				for _, sessionID := range sessionIDs {
					_, err := m.GetSession(sessionID)
					if err != nil {
						logger.Warn("获取会话详情失败", zap.String("session_id", sessionID), zap.Error(err))
					}
				}
			}
		}()
		// 多设备登录：检查同设备是否已存在会话，如果存在则替换
		//if err := m.ClearDeviceSession(session.UserID, session.DeviceID, "同设备新登录"); err != nil {
		//	logger.Warn("清除同设备旧会话失败",
		//		zap.String("user_id", session.UserID),
		//		zap.String("device_id", session.DeviceID),
		//		zap.Error(err))
		//}
	}

	// 2. 设置会话过期时间
	if session.ExpiresAt.IsZero() {
		expireDuration := time.Hour * time.Duration(m.config.SessionExpireTime)
		session.ExpiresAt = time.Now().Add(expireDuration)
	}

	// 3. 序列化会话信息
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化会话信息失败: %v", err)
	}

	// 4. 计算过期时间
	expiration := time.Until(session.ExpiresAt)
	if expiration <= 0 {
		return errors.New("会话已过期")
	}

	// 5. 使用Pipeline批量操作Redis - 事务性操作
	pipe := m.rdb.Pipeline()

	// 批量添加Redis操作
	userSessionsKey := SSOUserSessionsPrefix + session.UserID
	sessionKey := SSOSessionPrefix + session.SessionID
	deviceSessionKey := SSODeviceSessionPrefix + session.UserID + ":" + session.DeviceID

	// 添加会话到用户会话集合
	pipe.SAdd(m.ctx, userSessionsKey, session.SessionID)
	pipe.Expire(m.ctx, userSessionsKey, expiration)

	// 存储会话详情
	pipe.Set(m.ctx, sessionKey, string(sessionData), expiration)

	// 存储设备会话映射
	pipe.Set(m.ctx, deviceSessionKey, session.SessionID, expiration)

	// 如果有Token哈希，建立Token到会话的映射
	if session.TokenHash != "" {
		tokenSessionKey := SSOTokenSessionPrefix + session.TokenHash
		pipe.Set(m.ctx, tokenSessionKey, session.SessionID, expiration)

		// 建立会话到Token的反向映射
		sessionTokenKey := SSOSessionTokenPrefix + session.SessionID
		pipe.Set(m.ctx, sessionTokenKey, session.TokenHash, expiration)
	}

	// 如果有刷新Token，建立映射
	//if session.RefreshToken != "" {
	//	refreshTokenKey := SSORefreshTokenPrefix + session.RefreshToken
	//	pipe.Set(m.ctx, refreshTokenKey, session.SessionID, expiration)
	//
	//	// 添加到用户刷新Token集合
	//	userRefreshKey := SSOUserRefreshPrefix + session.UserID
	//	pipe.SAdd(m.ctx, userRefreshKey, session.RefreshToken)
	//	pipe.Expire(m.ctx, userRefreshKey, expiration)
	//}

	// 6. 执行批量操作
	_, err = pipe.Exec(m.ctx)
	if err != nil {
		return fmt.Errorf("批量创建多设备会话失败: %v", err)
	}

	logger.Info("创建SSO会话成功",
		zap.String("user_id", session.UserID),
		zap.String("session_id", session.SessionID),
		zap.String("device_id", session.DeviceID),
		zap.String("ip_address", session.IPAddress))

	return nil
}

// GetSession 获取单个会话详情
func (m *SSOManager) GetSession(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("会话ID不能为空")
	}

	sessionKey := SSOSessionPrefix + sessionID
	sessionData, err := m.rdb.Get(m.ctx, sessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // 会话不存在
		}
		return nil, fmt.Errorf("获取会话详情失败: %v", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return nil, fmt.Errorf("反序列化会话信息失败: %v", err)
	}

	// 检查会话是否过期
	if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
		// 会话已过期，异步清理
		go func() {
			if err := m.deleteExpiredSession(sessionID, session, "会话过期"); err != nil {
				logger.Warn("清理过期会话失败", zap.String("session_id", sessionID), zap.Error(err))
			}
		}()
		return nil, nil
	}

	return &session, nil
}

// GetSessionByTokenHash 根据Token哈希获取会话
func (m *SSOManager) GetSessionByTokenHash(tokenHash string) (*Session, error) {
	if tokenHash == "" {
		return nil, errors.New("Token哈希不能为空")
	}

	// 1. 从Token映射获取会话ID
	tokenSessionKey := SSOTokenSessionPrefix + tokenHash
	sessionID, err := m.rdb.Get(m.ctx, tokenSessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("获取Token对应会话ID失败: %v", err)
	}

	// 2. 获取会话详情
	return m.GetSession(sessionID)
}

// DeleteSession 删除指定会话
func (m *SSOManager) DeleteSession(sessionID, reason string) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	// 1. 直接从Redis获取会话详情
	sessionKey := SSOSessionPrefix + sessionID
	sessionData, err := m.rdb.Get(m.ctx, sessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil // 会话不存在，认为删除成功
		}
		return fmt.Errorf("获取会话详情失败: %v", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		// 如果反序列化失败，仍然尝试删除会话key，避免脏数据
		logger.Warn("反序列化会话信息失败，将强制删除会话",
			zap.String("session_id", sessionID),
			zap.Error(err))
		m.rdb.Del(m.ctx, sessionKey)
		return nil
	}

	return m.deleteExpiredSession(sessionID, session, reason)
}

// deleteExpiredSession 删除过期会话的专用函数
func (m *SSOManager) deleteExpiredSession(sessionID string, session Session, reason string) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	// 使用Pipeline批量删除
	pipe := m.rdb.Pipeline()

	// 删除会话详情
	sessionKey := SSOSessionPrefix + sessionID
	pipe.Del(m.ctx, sessionKey)

	// 从用户会话集合中移除
	userSessionsKey := SSOUserSessionsPrefix + session.UserID
	pipe.SRem(m.ctx, userSessionsKey, sessionID)

	// 删除设备会话映射
	deviceSessionKey := SSODeviceSessionPrefix + session.UserID + ":" + session.DeviceID
	pipe.Del(m.ctx, deviceSessionKey)

	// 删除Token会话映射
	if session.TokenHash != "" {
		tokenSessionKey := SSOTokenSessionPrefix + session.TokenHash
		pipe.Del(m.ctx, tokenSessionKey)

		// 删除会话Token映射
		sessionTokenKey := SSOSessionTokenPrefix + sessionID
		pipe.Del(m.ctx, sessionTokenKey)

		// 将Token加入黑名单
		blacklistKey := SSOBlacklistPrefix + session.TokenHash
		// 注意：这里需要知道JWT的过期时间，如果没有配置对象传递进来，可以给一个默认值
		// 或者从SSOConfig中获取一个BlacklistExpireTime
		expiration := time.Hour * 72 // 默认72小时
		pipe.Set(m.ctx, blacklistKey, "1", expiration)
	}

	// 删除刷新Token映射
	if session.RefreshToken != "" {
		refreshKey := SSORefreshTokenPrefix + session.RefreshToken
		pipe.Del(m.ctx, refreshKey)

		// 从用户刷新Token集合中移除
		userRefreshKey := SSOUserRefreshPrefix + session.UserID
		pipe.SRem(m.ctx, userRefreshKey, session.RefreshToken)
	}

	// 执行Pipeline
	_, err := pipe.Exec(m.ctx)
	if err != nil {
		return fmt.Errorf("删除会话失败: %v", err)
	}

	logger.Info("删除会话成功",
		zap.String("user_id", session.UserID),
		zap.String("session_id", sessionID),
		zap.String("reason", reason))

	return nil
}

// ClearUserAllSessions 清除用户所有会话
func (m *SSOManager) ClearUserAllSessions(userID, reason string) error {
	if userID == "" {
		return errors.New("用户ID不能为空")
	}

	// 1. 获取用户所有会话ID
	userSessionsKey := SSOUserSessionsPrefix + userID
	sessionIDs, err := m.rdb.SMembers(m.ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("获取用户会话列表失败: %v", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// 2. 批量删除
	for _, sessionID := range sessionIDs {
		// 获取会话详情以便完整清理
		session, err := m.GetSession(sessionID)
		if err != nil {
			logger.Warn("获取会话详情失败", zap.String("session_id", sessionID), zap.Error(err))
			continue
		}
		if session != nil {
			if err := m.deleteExpiredSession(sessionID, *session, reason); err != nil {
				logger.Warn("删除会话失败", zap.String("session_id", sessionID), zap.Error(err))
			}
		}
	}

	// 3. 确保集合清空
	m.rdb.Del(m.ctx, userSessionsKey)
	m.rdb.Del(m.ctx, SSOUserRefreshPrefix+userID)

	return nil
}

// UpdateSessionActivity 更新会话活动时间
func (m *SSOManager) UpdateSessionActivity(sessionID string) error {
	if sessionID == "" {
		return errors.New("会话ID不能为空")
	}

	// 1. 获取会话
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("会话不存在")
	}

	// 2. 更新最后活动时间
	session.LastActivity = time.Now()

	// 3. 更新Redis
	updateRedis := func() {
		sessionData, err := json.Marshal(session)
		if err != nil {
			logger.Error("序列化会话信息失败", zap.Error(err))
			return
		}

		sessionKey := SSOSessionPrefix + sessionID
		expiration := time.Until(session.ExpiresAt)
		if expiration <= 0 {
			m.deleteExpiredSession(sessionID, *session, "更新时发现会话过期")
			return
		}

		if err := m.rdb.Set(context.Background(), sessionKey, string(sessionData), expiration).Err(); err != nil {
			logger.Error("更新会话活动时间失败", zap.Error(err))
		}
	}

	if m.config.EnableAsyncUpdate {
		go updateRedis()
	} else {
		updateRedis()
	}

	return nil
}

// ClearDeviceSession 清除指定设备的会话
func (m *SSOManager) ClearDeviceSession(userID, deviceID, reason string) error {
	if userID == "" || deviceID == "" {
		return errors.New("用户ID和设备ID不能为空")
	}

	// 1. 获取设备当前会话ID
	deviceSessionKey := SSODeviceSessionPrefix + userID + ":" + deviceID
	sessionID, err := m.rdb.Get(m.ctx, deviceSessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("获取设备会话ID失败: %v", err)
	}

	// 2. 删除会话
	return m.DeleteSession(sessionID, reason)
}

// updateSessionAfterRefresh 刷新Token后更新会话信息
func (m *SSOManager) updateSessionAfterRefresh(session *Session, oldTokenHash, oldRefreshToken, newTokenHash, newRefreshToken string) error {
	if session == nil {
		return errors.New("会话信息不能为空")
	}
	if newTokenHash == "" || newRefreshToken == "" {
		return errors.New("新Token信息不能为空")
	}

	expiration := time.Until(session.ExpiresAt)
	if expiration <= 0 {
		return errors.New("会话已过期")
	}

	// 序列化更新后的会话信息
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化会话信息失败: %v", err)
	}

	// 使用Pipeline批量更新 - 确保原子性
	pipe := m.rdb.Pipeline()

	// 1. 更新会话详情
	sessionKey := SSOSessionPrefix + session.SessionID
	pipe.Set(m.ctx, sessionKey, string(sessionData), expiration)

	// 2. 处理Token映射的更新
	// 删除旧的Token映射
	if oldTokenHash != "" && oldTokenHash != newTokenHash {
		oldTokenSessionKey := SSOTokenSessionPrefix + oldTokenHash
		pipe.Del(m.ctx, oldTokenSessionKey)
	}

	// 添加新的Token映射
	newTokenSessionKey := SSOTokenSessionPrefix + newTokenHash
	pipe.Set(m.ctx, newTokenSessionKey, session.SessionID, expiration)

	// 更新会话Token反向映射
	sessionTokenKey := SSOSessionTokenPrefix + session.SessionID
	pipe.Set(m.ctx, sessionTokenKey, newTokenHash, expiration)

	// 3. 处理刷新Token映射的更新
	// 删除旧的刷新Token映射
	if oldRefreshToken != "" && oldRefreshToken != newRefreshToken {
		oldRefreshKey := SSORefreshTokenPrefix + oldRefreshToken
		pipe.Del(m.ctx, oldRefreshKey)

		// 从用户刷新Token集合中移除旧Token
		userRefreshKey := SSOUserRefreshPrefix + session.UserID
		pipe.SRem(m.ctx, userRefreshKey, oldRefreshToken)
	}

	// 添加新的刷新Token映射
	newRefreshKey := SSORefreshTokenPrefix + newRefreshToken
	pipe.Set(m.ctx, newRefreshKey, session.SessionID, expiration)

	// 添加新刷新Token到用户集合
	userRefreshKey := SSOUserRefreshPrefix + session.UserID
	pipe.SAdd(m.ctx, userRefreshKey, newRefreshToken)
	pipe.Expire(m.ctx, userRefreshKey, expiration)

	// 4. 执行批量更新
	_, err = pipe.Exec(m.ctx)
	if err != nil {
		return fmt.Errorf("批量更新会话信息失败: %v", err)
	}

	return nil
}
