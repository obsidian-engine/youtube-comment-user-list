package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// MonitoringHandler チャット監視のHTTPリクエストを処理します
type MonitoringHandler struct {
	chatMonitoringUC *usecase.ChatMonitoringUseCase
	logger           service.Logger
}

// NewMonitoringHandler 新しい監視ハンドラーを作成します
func NewMonitoringHandler(
	chatMonitoringUC *usecase.ChatMonitoringUseCase,
	logger service.Logger,
) *MonitoringHandler {
	return &MonitoringHandler{
		chatMonitoringUC: chatMonitoringUC,
		logger:           logger,
	}
}

// StartMonitoringRequest 監視開始のリクエストボディを表します
type StartMonitoringRequest struct {
	VideoInput string `json:"video_input"`
	MaxUsers   int    `json:"max_users"`
}

// StartMonitoringResponse the response for starting monitoringを表します
type StartMonitoringResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// UserListResponse the user list responseを表します
type UserListResponse struct {
	Success bool           `json:"success"`
	Users   []*entity.User `json:"users"`
	Count   int            `json:"count"`
	Error   string         `json:"error,omitempty"`
}

// StartMonitoring handles POST /api/monitoring/start
func (h *MonitoringHandler) StartMonitoring(c *gin.Context) {
	correlationID := fmt.Sprintf("http-%s", c.GetHeader("X-Request-ID"))
	if correlationID == "http-" {
		correlationID = fmt.Sprintf("http-%s", c.GetString("requestId"))
	}

	h.logger.LogAPI("INFO", "Start monitoring request received", "", correlationID, map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	})

	var req StartMonitoringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError("ERROR", "Invalid request body", "", correlationID, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "correlationID": correlationID})
		return
	}

	// Validate request
	if req.VideoInput == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video_input is required", "correlationID": correlationID})
		return
	}

	if req.MaxUsers <= 0 {
		req.MaxUsers = constants.DefaultMaxUsers
	}

	// Start monitoring
	session, err := h.chatMonitoringUC.StartMonitoring(c.Request.Context(), req.VideoInput, req.MaxUsers)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to start monitoring", req.VideoInput, correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	response := StartMonitoringResponse{
		Success: true,
		VideoID: session.VideoID,
		Message: "Monitoring started successfully",
	}

	h.logger.LogAPI("INFO", "Start monitoring response", session.VideoID, correlationID, map[string]interface{}{
		"success": response.Success,
		"videoID": response.VideoID,
	})

	c.JSON(http.StatusOK, response)
}

// StopMonitoring handles POST /api/monitoring/stop/{videoId}
func (h *MonitoringHandler) StopMonitoring(c *gin.Context) {
	correlationID := fmt.Sprintf("http-%s", c.GetString("requestId"))
	videoID := c.Param("videoId")

	h.logger.LogAPI("INFO", "Stop monitoring request received", videoID, correlationID, map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	})

	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video ID is required", "correlationID": correlationID})
		return
	}

	err := h.chatMonitoringUC.StopMonitoring(videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to stop monitoring", videoID, correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Monitoring stopped successfully",
		"videoID": videoID,
	})
}

// GetUserList handles GET /api/monitoring/{videoId}/users
func (h *MonitoringHandler) GetUserList(c *gin.Context) {
	correlationID := fmt.Sprintf("http-%s", c.GetString("requestId"))
	videoID := c.Param("videoId")

	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video ID is required", "correlationID": correlationID})
		return
	}

	h.logger.LogAPI("INFO", "Get user list request received", videoID, correlationID, nil)

	users, err := h.chatMonitoringUC.GetUserList(c.Request.Context(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get user list", videoID, correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	response := UserListResponse{
		Success: true,
		Users:   users,
		Count:   len(users),
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoStatus handles GET /api/monitoring/{videoId}/status
func (h *MonitoringHandler) GetVideoStatus(c *gin.Context) {
	correlationID := fmt.Sprintf("http-%s", c.GetString("requestId"))
	videoID := c.Param("videoId")

	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video ID is required", "correlationID": correlationID})
		return
	}

	h.logger.LogAPI("INFO", "Get video status request received", videoID, correlationID, nil)

	status, err := h.chatMonitoringUC.GetVideoStatus(c.Request.Context(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get video status", videoID, correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"videoID": videoID,
		"status":  status,
	})
}

// GetActiveVideos handles GET /api/monitoring/active
func (h *MonitoringHandler) GetActiveVideos(c *gin.Context) {
	correlationID := fmt.Sprintf("http-%s", c.GetString("requestId"))

	h.logger.LogAPI("INFO", "Get active videos request received", "", correlationID, nil)

	activeVideos := h.chatMonitoringUC.GetActiveVideos()
	// No error handling needed for GetActiveVideos as it doesn't return an error

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"videos":  activeVideos,
		"count":   len(activeVideos),
	})
}

// Helper methods
