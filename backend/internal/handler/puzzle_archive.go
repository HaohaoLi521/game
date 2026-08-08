package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// PuzzleArchiveHandler 是管理端题目归档接口入口。
type PuzzleArchiveHandler struct{ service *service.PuzzleArchiveService }

func NewPuzzleArchiveHandler(service *service.PuzzleArchiveService) *PuzzleArchiveHandler {
	return &PuzzleArchiveHandler{service: service}
}

func (h *PuzzleArchiveHandler) ListArchived(c *gin.Context) {
	items, err := h.service.ListArchived(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list archived puzzles")
		return
	}
	response.OK(c, items)
}

func (h *PuzzleArchiveHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.Error(c, http.StatusBadRequest, "invalid puzzle id")
		return
	}
	if err = h.service.Restore(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"restored": true, "id": id})
}
