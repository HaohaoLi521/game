package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// PlayerHandler 是玩家账户的 HTTP 入口。
type PlayerHandler struct{ service *service.PlayerService }

func NewPlayerHandler(service *service.PlayerService) *PlayerHandler {
	return &PlayerHandler{service: service}
}

type PlayerAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type PlayerAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	PlayerID     uint64 `json:"player_id"`
	Username     string `json:"username"`
}

func (h *PlayerHandler) Register(c *gin.Context) { h.auth(c, true) }
func (h *PlayerHandler) Login(c *gin.Context)    { h.auth(c, false) }
func (h *PlayerHandler) auth(c *gin.Context, register bool) {
	var req PlayerAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	var token string
	var refresh string
	var expires int
	var playerID uint64
	var username string
	var err error
	if register {
		r, p, e := h.service.Register(c.Request.Context(), req.Username, req.Password, c.ClientIP(), c.GetHeader("User-Agent"))
		err = e
		if r != nil {
			token = r.AccessToken
			refresh = r.RefreshToken
			expires = r.ExpiresIn
		}
		if p != nil {
			playerID = p.ID
			username = p.Username
		}
	} else {
		r, p, e := h.service.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP(), c.GetHeader("User-Agent"))
		err = e
		if r != nil {
			token = r.AccessToken
			refresh = r.RefreshToken
			expires = r.ExpiresIn
		}
		if p != nil {
			playerID = p.ID
			username = p.Username
		}
	}
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, PlayerAuthResponse{token, refresh, expires, playerID, username})
}
func (h *PlayerHandler) Progress(c *gin.Context) {
	id, err := strconv.ParseUint(c.GetString("UserID"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	progress, err := h.service.Progress(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, progress)
}
func (h *PlayerHandler) MarkSolved(c *gin.Context) {
	uid, err := strconv.ParseUint(c.GetString("UserID"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err = h.service.MarkSolved(c.Request.Context(), uid, pid); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"saved": true})
}
