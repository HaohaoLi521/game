package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// AdminSubmissionHandler 是管理端批量审核的 HTTP 入口。
type AdminSubmissionHandler struct {
	service *service.AdminSubmissionService
}

func NewAdminSubmissionHandler(service *service.AdminSubmissionService) *AdminSubmissionHandler {
	return &AdminSubmissionHandler{service: service}
}

// BatchReviewRequest 是批量审核请求 DTO。
type BatchReviewRequest struct {
	SubmissionIDs []int64                   `json:"submission_ids"`
	Action        service.BatchReviewAction `json:"action"`
	ReviewNote    string                    `json:"review_note"`
}

// BatchReview 对多条待审核投稿执行同一种审核动作。
func (h *AdminSubmissionHandler) BatchReview(c *gin.Context) {
	var req BatchReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	items, err := h.service.BatchReview(service.BatchReviewInput{SubmissionIDs: req.SubmissionIDs, Action: req.Action, ReviewNote: req.ReviewNote, Reviewer: c.GetString("admin_username")})
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, items)
}
