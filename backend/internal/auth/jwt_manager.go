package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager JWT管理器
type JWTManager struct {
	config JWTConfig
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(config JWTConfig) *JWTManager {
	if config.AccessTokenExpire <= 0 {
		config.AccessTokenExpire = time.Hour * 72 // 默认72小时
	}
	if config.RefreshTokenExpire <= 0 {
		config.RefreshTokenExpire = time.Hour * 72 // 默认72小时
	}

	return &JWTManager{
		config: config,
	}
}

// GenerateAccessToken 生成访问Token
// claimsCustomizer 用于自定义Claims
func (j *JWTManager) GenerateAccessToken(userID string, claimsCustomizer func(jwt.MapClaims)) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": j.config.AppID,
		"sub": j.config.DomainName,
		"aud": userID,
		"exp": now.Add(j.config.AccessTokenExpire).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"jti": newID(),
	}

	// 应用自定义Claims
	if claimsCustomizer != nil {
		claimsCustomizer(claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// GenerateRefreshToken 生成刷新Token
func (j *JWTManager) GenerateRefreshToken() (string, error) {
	return newID(), nil
}

// ParseToken 解析Token
func (j *JWTManager) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
