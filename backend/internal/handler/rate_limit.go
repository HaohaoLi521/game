package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// RateLimitHandler 将限流服务适配为 Gin 路由中间件。
type RateLimitHandler struct {
	service *service.RateLimitService
}

// NewRateLimitHandler 创建限流中间件适配器。
func NewRateLimitHandler(service *service.RateLimitService) *RateLimitHandler {
	return &RateLimitHandler{service: service}
}

// Middleware 按 scope 限制请求频率；Redis 故障时返回 503，避免静默失去保护。
func (h *RateLimitHandler) Middleware(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject := c.GetString("UserID")
		if strings.TrimSpace(subject) == "" {
			subject = c.ClientIP()
		}
		decision, err := h.service.Allow(c.Request.Context(), scope, subject)
		if err != nil {
			response.Error(c, http.StatusServiceUnavailable, "rate limit unavailable")
			c.Abort()
			return
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(decision.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
		if !decision.Allowed {
			c.Header("Retry-After", strconv.Itoa(decision.RetryAfter))
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}
