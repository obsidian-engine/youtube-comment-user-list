package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// MonitoringHandler チャット監視のHTTPリクエストを処理します
type MonitoringHandler struct {
	chatMonitoringUC *usecase.ChatMonitoringUseCase
	logger           repository.Logger
}

// NewMonitoringHandler 新しい監視ハンドラーを作成します
func NewMonitoringHandler(
	chatMonitoringUC *usecase.ChatMonitoringUseCase,
	logger repository.Logger,
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

// StartMonitoringResponse 監視開始のレスポンスを表します
type StartMonitoringResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// UserListResponse ユーザーリストのレスポンスを表します
type UserListResponse struct {
	Success bool           `json:"success"`
	Users   []*entity.User `json:"users"`
	Count   int            `json:"count"`
	Error   string         `json:"error,omitempty"`
}

// StartMonitoring POST /api/monitoring/start を処理します
func (h *MonitoringHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("X-Request-ID"))
	if correlationID == "http-" {
		correlationID = fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	}

	h.logger.LogAPI("INFO", "Start monitoring request received", "", correlationID, map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	var req StartMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogError("ERROR", "Invalid request body", "", correlationID, err, nil)
		h.writeJSONError(w, http.StatusBadRequest, "Invalid request body", correlationID)
		return
	}

	// リクエストを検証
	if req.VideoInput == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video_input is required", correlationID)
		return
	}

	if req.MaxUsers <= 0 {
		req.MaxUsers = constants.DefaultMaxUsers
	}

	// 監視を開始
	session, err := h.chatMonitoringUC.StartMonitoring(r.Context(), req.VideoInput, req.MaxUsers)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to start monitoring", req.VideoInput, correlationID, err, nil)
		h.writeJSONError(w, http.StatusInternalServerError, err.Error(), correlationID)
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

	h.writeJSON(w, http.StatusOK, response)
}

// StopMonitoring POST /api/monitoring/stop を処理します
func (h *MonitoringHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))

	h.logger.LogAPI("INFO", "Stop monitoring request received", "", correlationID, map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	err := h.chatMonitoringUC.StopMonitoring()
	if err != nil {
		// 既に停止済みの場合は正常レスポンスを返す
		if err.Error() == "no active monitoring session" {
			h.logger.LogAPI("INFO", "Monitoring already stopped", "", correlationID, nil)
			h.writeJSON(w, http.StatusOK, map[string]interface{}{
				"success":       true,
				"message":       "Monitoring already stopped",
				"correlationID": correlationID,
			})
			return
		}

		// その他のエラーの場合はエラーレスポンス
		h.logger.LogError("ERROR", "Failed to stop monitoring", "", correlationID, err, nil)
		h.writeJSONError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       "Monitoring stopped successfully",
		"correlationID": correlationID,
	})
}

// GetUserList GET /api/monitoring/{videoId}/users を処理します
func (h *MonitoringHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	videoID := chi.URLParam(r, "videoId")

	if videoID == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video ID is required", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get user list request received", videoID, correlationID, nil)

	users, err := h.chatMonitoringUC.GetUserList(r.Context(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get user list", videoID, correlationID, err, nil)
		h.logger.LogAPI("DEBUG", "Sending error response", videoID, correlationID, map[string]interface{}{
			"statusCode": http.StatusInternalServerError,
			"error":      err.Error(),
		})
		h.writeJSONError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	response := UserListResponse{
		Success: true,
		Users:   users,
		Count:   len(users),
	}

	// デバッグ用：レスポンスをログ出力
	h.logger.LogAPI("DEBUG", "Sending user list response", videoID, correlationID, map[string]interface{}{
		"userCount": len(users),
		"success":   true,
	})

	h.writeJSON(w, http.StatusOK, response)
}

// GetActiveVideoID 現在監視中のvideoIDを取得します
func (h *MonitoringHandler) GetActiveVideoID(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))

	videoID, isActive, exists := h.chatMonitoringUC.GetActiveVideoID()
	if !exists {
		h.logger.LogAPI("INFO", "No monitoring session found", "", correlationID, nil)
		h.writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"error":         "No monitoring session found",
			"message":       "no active or previous video",
			"correlationID": correlationID,
		})
		return
	}

	h.logger.LogAPI("INFO", "Video ID retrieved", videoID, correlationID, map[string]interface{}{
		"isActive": isActive,
	})
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"videoId":       videoID,
		"isActive":      isActive,
		"correlationID": correlationID,
	})
}

// GetVideoStatus GET /api/monitoring/{videoId}/status を処理します
func (h *MonitoringHandler) GetVideoStatus(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	videoID := chi.URLParam(r, "videoId")

	if videoID == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video ID is required", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get video status request received", videoID, correlationID, nil)

	status, err := h.chatMonitoringUC.GetVideoStatus(r.Context(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get video status", videoID, correlationID, err, nil)
		h.writeJSONError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"videoID": videoID,
		"status":  status,
	})
}

// writeJSON はJSONレスポンスを書き込みます
func (h *MonitoringHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON response", "", "", err, nil)
	}
}

// writeJSONError はJSONエラーレスポンスを書き込みます
func (h *MonitoringHandler) writeJSONError(w http.ResponseWriter, statusCode int, message, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":         message,
		"correlationID": correlationID,
	}); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON error response", "", correlationID, err, nil)
	}
}
