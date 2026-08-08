package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// ReadinessHandler 是服务就绪检查的 HTTP 入口。
type ReadinessHandler struct{ service *service.ReadinessService }

// NewReadinessHandler 创建就绪检查 HTTP 处理器。
func NewReadinessHandler(service *service.ReadinessService) *ReadinessHandler {
	return &ReadinessHandler{service: service}
}

func (h *ReadinessHandler) Check(c *gin.Context) {
	result := h.service.Check(c.Request.Context())
	if !result.Ready {
		// 依赖不可用时仍返回详细状态，便于编排系统和运维定位故障。
		c.JSON(http.StatusServiceUnavailable, response.Body{Data: result, Error: "dependencies unavailable"})
		return
	}
	response.OK(c, result)
}
