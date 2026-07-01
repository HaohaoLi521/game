package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/model"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

type PuzzleHandler struct {
	game *service.GameService
}

func NewPuzzleHandler(game *service.GameService) *PuzzleHandler {
	return &PuzzleHandler{game: game}
}

func (h *PuzzleHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok", "service": "this-is-pun"})
}

func (h *PuzzleHandler) ListPuzzleSets(c *gin.Context) {
	response.OK(c, h.game.ListPuzzleSets())
}

func (h *PuzzleHandler) GetPuzzleSet(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	set, err := h.game.GetPuzzleSet(id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, set)
}

func (h *PuzzleHandler) ListPuzzlesBySet(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	puzzles, err := h.game.ListPublicPuzzlesBySet(id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, puzzles)
}

func (h *PuzzleHandler) GetPuzzle(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	puzzle, err := h.game.GetPublicPuzzle(id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, puzzle)
}

func (h *PuzzleHandler) CheckAnswer(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req service.CheckAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.game.CheckAnswer(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PuzzleHandler) GetHint(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req service.HintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	hint, err := h.game.GetHint(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, hint)
}

func (h *PuzzleHandler) CreateSubmission(c *gin.Context) {
	var req service.SubmissionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	submission, err := h.game.CreateSubmission(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, submission)
}

func (h *PuzzleHandler) RegisterAdmin(c *gin.Context) {
	var req service.AdminAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.game.RegisterAdmin(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PuzzleHandler) LoginAdmin(c *gin.Context) {
	var req service.AdminAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.game.LoginAdmin(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PuzzleHandler) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		session, ok := h.game.AdminSession(token)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Set("admin_username", session.Username)
		c.Next()
	}
}

func (h *PuzzleHandler) ListAdminPuzzleSets(c *gin.Context) {
	response.OK(c, h.game.ListPuzzleSets())
}

func (h *PuzzleHandler) ListAdminPuzzles(c *gin.Context) {
	puzzles, err := h.game.ListAdminPuzzles()
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, puzzles)
}

func (h *PuzzleHandler) CreateAdminPuzzle(c *gin.Context) {
	var req service.PuzzleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	puzzle, err := h.game.CreatePuzzle(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, puzzle)
}

func (h *PuzzleHandler) UpdateAdminPuzzle(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req service.PuzzleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	puzzle, err := h.game.UpdatePuzzle(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, puzzle)
}

func (h *PuzzleHandler) DeleteAdminPuzzle(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.game.DeletePuzzle(id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *PuzzleHandler) ListAdminSubmissions(c *gin.Context) {
	status := model.SubmissionStatus(c.Query("status"))
	submissions, err := h.game.ListSubmissions(status)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, submissions)
}

func (h *PuzzleHandler) ApproveSubmission(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req service.SubmissionReviewRequest
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	submission, err := h.game.ApproveSubmission(id, c.GetString("admin_username"), req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, submission)
}

func (h *PuzzleHandler) RejectSubmission(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req service.SubmissionReviewRequest
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	submission, err := h.game.RejectSubmission(id, c.GetString("admin_username"), req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, submission)
}

func (h *PuzzleHandler) GetExplanation(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	explanation, err := h.game.GetExplanation(id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, explanation)
}

func parseID(c *gin.Context, key string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrNotFound) {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrInvalidRequest) {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	if errors.Is(err, service.ErrAlreadyExists) {
		response.Error(c, http.StatusConflict, "already exists")
		return
	}
	if errors.Is(err, service.ErrUnauthorized) {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	response.Error(c, http.StatusInternalServerError, "internal server error")
}
