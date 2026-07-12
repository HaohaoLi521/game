package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// PlayerSubmissionHandler 是玩家投稿接口的 HTTP 入口。
type PlayerSubmissionHandler struct {
	service *service.PlayerSubmissionService
}

func NewPlayerSubmissionHandler(service *service.PlayerSubmissionService) *PlayerSubmissionHandler {
	return &PlayerSubmissionHandler{service: service}
}

// Create 创建一条归属于当前玩家的投稿。
func (h *PlayerSubmissionHandler) Create(c *gin.Context) {
	playerID, ok := currentPlayerID(c)
	if !ok {
		return
	}
	var req service.SubmissionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.service.Create(c.Request.Context(), playerID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

// List 返回当前玩家的投稿状态列表。
func (h *PlayerSubmissionHandler) List(c *gin.Context) {
	playerID, ok := currentPlayerID(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), playerID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, items)
}

// Get 返回当前玩家名下某条投稿的详情。
func (h *PlayerSubmissionHandler) Get(c *gin.Context) {
	playerID, ok := currentPlayerID(c)
	if !ok {
		return
	}
	submissionID, ok := parseID(c, "id")
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), playerID, submissionID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

func currentPlayerID(c *gin.Context) (uint64, bool) {
	playerID, err := strconv.ParseUint(c.GetString("UserID"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}
	return playerID, true
}
