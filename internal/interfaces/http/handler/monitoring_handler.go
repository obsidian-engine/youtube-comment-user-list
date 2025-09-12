package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

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
func (h *MonitoringHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("X-Request-ID"))
	if correlationID == "http-" {
		correlationID = fmt.Sprintf("http-%d", r.Context().Value("requestId"))
	}

	h.logger.LogAPI("INFO", "Start monitoring request received", "", correlationID, map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	if r.Method != http.MethodPost {
		h.respondWithError(w, "Method not allowed", correlationID)
		return
	}

	var req StartMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogError("ERROR", "Invalid request body", "", correlationID, err, nil)
		h.respondWithError(w, "Invalid request body", correlationID)
		return
	}

	// Validate request
	if req.VideoInput == "" {
		h.respondWithError(w, "video_input is required", correlationID)
		return
	}

	if req.MaxUsers <= 0 {
		req.MaxUsers = constants.DefaultMaxUsers
	}

	// Start monitoring
	session, err := h.chatMonitoringUC.StartMonitoring(r.Context(), req.VideoInput, req.MaxUsers)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to start monitoring", req.VideoInput, correlationID, err, nil)
		h.respondWithError(w, err.Error(), correlationID)
		return
	}

	response := StartMonitoringResponse{
		Success: true,
		VideoID: session.VideoID,
		Message: "Monitoring started successfully",
	}

	h.respondWithJSON(w, response, session.VideoID, correlationID)
}

// StopMonitoring handles POST /api/monitoring/stop/{videoId}
func (h *MonitoringHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodPost {
		h.respondWithError(w, "Method not allowed", correlationID)
		return
	}

	// Extract video ID from URL path
	videoID := h.extractVideoIDFromPath(r.URL.Path)
	if videoID == "" {
		h.respondWithError(w, "video ID is required", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Stop monitoring request received", videoID, correlationID, nil)

	if err := h.chatMonitoringUC.StopMonitoring(videoID); err != nil {
		h.logger.LogError("ERROR", "Failed to stop monitoring", videoID, correlationID, err, nil)
		h.respondWithError(w, err.Error(), correlationID)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Monitoring stopped successfully",
	}

	h.respondWithJSON(w, response, videoID, correlationID)
}

// GetUserList handles GET /api/monitoring/{videoId}/users
func (h *MonitoringHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodGet {
		h.respondWithError(w, "Method not allowed", correlationID)
		return
	}

	// Extract video ID from URL path
	videoID := h.extractVideoIDFromPath(r.URL.Path)
	if videoID == "" {
		h.respondWithError(w, "video ID is required", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get user list request received", videoID, correlationID, nil)

	users, err := h.chatMonitoringUC.GetUserList(r.Context(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get user list", videoID, correlationID, err, nil)
		h.respondWithError(w, err.Error(), correlationID)
		return
	}

	response := UserListResponse{
		Success: true,
		Users:   users,
		Count:   len(users),
	}

	h.respondWithJSON(w, response, videoID, correlationID)
}

// GetVideoStatus handles GET /api/monitoring/{videoId}/status
func (h *MonitoringHandler) GetVideoStatus(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodGet {
		h.respondWithError(w, "Method not allowed", correlationID)
		return
	}

	// Extract video ID from URL path
	videoID := h.extractVideoIDFromPath(r.URL.Path)
	if videoID == "" {
		h.respondWithError(w, "video ID is required", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get video status request received", videoID, correlationID, nil)

	status, err := h.chatMonitoringUC.GetVideoStatus(r.Context(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get video status", videoID, correlationID, err, nil)
		h.respondWithError(w, err.Error(), correlationID)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"status":  status,
	}

	h.respondWithJSON(w, response, videoID, correlationID)
}

// GetActiveVideos handles GET /api/monitoring/active
func (h *MonitoringHandler) GetActiveVideos(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodGet {
		h.respondWithError(w, "Method not allowed", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get active videos request received", "", correlationID, nil)

	videos := h.chatMonitoringUC.GetActiveVideos()

	response := map[string]interface{}{
		"success": true,
		"videos":  videos,
		"count":   len(videos),
	}

	h.respondWithJSON(w, response, "", correlationID)
}

// Helper methods

func (h *MonitoringHandler) extractVideoIDFromPath(path string) string {
	// Simple path parsing - in a real implementation, you might use a router like gorilla/mux
	// Expected paths: /api/monitoring/{videoId}/users, /api/monitoring/{videoId}/status, etc.
	parts := splitPath(path)
	if len(parts) >= constants.MinPathPartsForAPI && parts[1] == "api" && parts[2] == "monitoring" {
		if len(parts) >= constants.MinPathPartsForVideoID {
			return parts[3]
		}
	}
	return ""
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func (h *MonitoringHandler) respondWithJSON(w http.ResponseWriter, data interface{}, videoID, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON response", videoID, correlationID, err, nil)
	}

	h.logger.LogAPI("INFO", "Response sent", videoID, correlationID, map[string]interface{}{
		"statusCode": http.StatusOK,
	})
}

func (h *MonitoringHandler) respondWithError(w http.ResponseWriter, message, correlationID string) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.LogError("ERROR", "Failed to encode error response", "", correlationID, err, nil)
	}

	h.logger.LogAPI("ERROR", "Error response sent", "", correlationID, map[string]interface{}{
		"statusCode": http.StatusOK,
		"error":      message,
	})
}
