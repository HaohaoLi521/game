package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

// MediaHandler 是媒体上传 HTTP 入口。
type MediaHandler struct{ service *service.MediaService }

func NewMediaHandler(service *service.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

// UploadResponse 是上传结果 DTO。
type UploadResponse struct {
	ID          uint64 `json:"id"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func (h *MediaHandler) Upload(c *gin.Context) {
	owner, err := strconv.ParseUint(c.GetString("UserID"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file is required")
		return
	}
	src, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid file")
		return
	}
	defer src.Close()
	asset, err := h.service.Upload(c.Request.Context(), owner, file.Filename, file.Header.Get("Content-Type"), file.Size, src)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, UploadResponse{asset.ID, asset.PublicURL, asset.ContentType, asset.Size})
}
