package auth

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	AuthManager     *AuthManager
	NoAuthPaths     []string
	LocalIPs        []string
	ClaimsExtractor func(*gin.Context, jwt.MapClaims)           // 自定义Claims提取逻辑
	SuccessCallback func(*gin.Context, *Session, jwt.MapClaims) // 验证成功后的回调
}

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(config MiddlewareConfig) gin.HandlerFunc {
	// 默认本地IP
	//if len(config.LocalIPs) == 0 {
	//	config.LocalIPs = []string{
	//		"127.0.0.1", "localhost", "::1", "192.168.", "10.0.", "172.16.",
	//	}
	//}

	return func(c *gin.Context) {
		// 1. 检查免登录路径
		for _, path := range config.NoAuthPaths {
			if matchNoAuthPath(c.Request.RequestURI, path) {
				// 检查本地IP访问
				//if isLocalRequest(c, config.LocalIPs) {
				c.Next()
				return
				//	return
				//}

				// 为了稳妥，我们先不直接return，而是跳过 Auth 检查
				//goto SkipAuth
			}
		}

		// 2. 执行认证
		{
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				authHeader = c.Query("access_token")
			}
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  "authorization required",
				})
				c.Abort()
				return
			}

			// 支持 Bearer 前缀也支持直接传 Token
			tokenString := authHeader
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}

			// 验证 Token
			res, err := config.AuthManager.ValidateToken(tokenString)
			if err != nil {
				logger.Debug("Token验证失败", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  "invalid token",
				})
				c.Abort()
				return
			}

			if !res.IsValid {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": http.StatusUnauthorized,
					"msg":  res.Error,
				})
				c.Abort()
				return
			}

			// 设置上下文
			c.Set("UserID", res.Session.UserID)
			c.Set("SessionID", res.Session.SessionID)

			// 调用自定义提取器
			if config.ClaimsExtractor != nil {
				config.ClaimsExtractor(c, res.Claims)
			}

			// 调用成功回调
			if config.SuccessCallback != nil {
				config.SuccessCallback(c, res.Session, res.Claims)
			}
			c.Next()
		}

		//SkipAuth:
		//	c.Next()
	}
}

func matchNoAuthPath(requestURI, noAuthPath string) bool {
	requestPath := requestURI
	if idx := strings.IndexByte(requestPath, '?'); idx >= 0 {
		requestPath = requestPath[:idx]
	}
	if noAuthPath == "" || !strings.Contains(requestPath, noAuthPath) {
		return false
	}

	searchFrom := 0
	for {
		idx := strings.Index(requestPath[searchFrom:], noAuthPath)
		if idx < 0 {
			return false
		}
		idx += searchFrom
		end := idx + len(noAuthPath)
		if end == len(requestPath) || requestPath[end] == '/' {
			return true
		}
		searchFrom = idx + 1
		if searchFrom >= len(requestPath) {
			return false
		}
	}
}

// isLocalRequest 检查是否为本地请求
func isLocalRequest(c *gin.Context, localIPs []string) bool {
	remoteIP := c.RemoteIP()
	for _, ip := range localIPs {
		if strings.HasSuffix(ip, ".") {
			// 子网匹配
			if strings.HasPrefix(remoteIP, ip) {
				return true
			}
		} else {
			// 精确匹配
			if remoteIP == ip || (net.ParseIP(remoteIP) != nil && net.ParseIP(remoteIP).Equal(net.ParseIP(ip))) {
				return true
			}
		}
	}
	return false
}
