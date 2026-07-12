package auth

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// HashToken 对Token进行哈希
func HashToken(tokenString string) string {
	hash := md5.Sum([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

// GetClientIP 获取客户端IP地址
func GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	if xRealIP != "" {
		return xRealIP
	}

	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if remoteAddr != "" {
		// 去掉端口号
		if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
			return remoteAddr[:idx]
		}
		return remoteAddr
	}

	return ""
}
