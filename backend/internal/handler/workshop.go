package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// WorkshopHandler 是创意工坊 HTTP 入口。
type WorkshopHandler struct{ service *service.WorkshopService }

func NewWorkshopHandler(service *service.WorkshopService) *WorkshopHandler {
	return &WorkshopHandler{service: service}
}

// List 返回公开的未下架题目卡片，支持 category、difficulty、page、page_size 查询参数。
func (h *WorkshopHandler) List(c *gin.Context) {
	difficulty, err := queryInt(c, "difficulty", 0)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid difficulty")
		return
	}
	page, err := queryInt(c, "page", 1)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page")
		return
	}
	pageSize, err := queryInt(c, "page_size", 12)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page_size")
		return
	}
	pageResult, err := h.service.List(c.Request.Context(), c.Query("category"), difficulty, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, pageResult)
}

func queryInt(c *gin.Context, key string, fallback int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
