package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/interfaces/http/response"
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

// StartMonitoring POST /api/monitoring/start を処理します
func (h *MonitoringHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("X-Request-ID"))
	if correlationID == "http-" {
		correlationID = fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	}

    h.logger.LogAPI(constants.LogLevelInfo, "Start monitoring request received", "", correlationID, map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	var req StartMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.LogError(constants.LogLevelError, "Invalid request body", "", correlationID, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "Invalid request body", correlationID)
		return
	}

	// リクエストを検証
	if req.VideoInput == "" {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "video_input is required", correlationID)
		return
	}

	if req.MaxUsers <= 0 {
		req.MaxUsers = constants.DefaultMaxUsers
	}

	// 監視を開始
	session, err := h.chatMonitoringUC.StartMonitoring(r.Context(), req.VideoInput, req.MaxUsers)
    if err != nil {
        h.logger.LogError(constants.LogLevelError, "Failed to start monitoring", req.VideoInput, correlationID, err, nil)
        response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), correlationID)
        return
    }

    h.logger.LogAPI(constants.LogLevelInfo, "Start monitoring response", session.VideoID, correlationID, map[string]interface{}{
		"success": true,
		"videoID": session.VideoID,
	})

	response.RenderStartMonitoring(w, r, session.VideoID, "Monitoring started successfully")
}

// StopMonitoring DELETE /api/monitoring/stop（または POST 互換）を処理します
func (h *MonitoringHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))

    h.logger.LogAPI(constants.LogLevelInfo, "Stop monitoring request received", "", correlationID, map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	err := h.chatMonitoringUC.StopMonitoring()
	if err != nil {
		// 既に停止済みの場合は正常レスポンスを返す
		if err.Error() == "no active monitoring session" {
            h.logger.LogAPI(constants.LogLevelInfo, "Monitoring already stopped", "", correlationID, nil)
			response.RenderSuccessWithCorrelation(w, r, map[string]string{
				"message": "Monitoring already stopped",
			}, correlationID)
			return
		}

		// その他のエラーの場合はエラーレスポンス
        h.logger.LogError(constants.LogLevelError, "Failed to stop monitoring", "", correlationID, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	response.RenderSuccessWithCorrelation(w, r, map[string]string{
		"message": "Monitoring stopped successfully",
	}, correlationID)
}

// GetUserList GET /api/monitoring/{videoId}/users を処理します
// 互換性のために復活（UIの一部がスナップショット取得に使用）。
func (h *MonitoringHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
    correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))
    videoID := chi.URLParam(r, "videoId")
    if videoID == "" {
        response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "video ID is required", correlationID)
        return
    }

    h.logger.LogAPI(constants.LogLevelInfo, "Get user list request received", videoID, correlationID, nil)

    users, err := h.chatMonitoringUC.GetUserList(r.Context(), videoID)
    if err != nil {
        h.logger.LogError(constants.LogLevelError, "Failed to get user list", videoID, correlationID, err, nil)
        response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), correlationID)
        return
    }

    response.RenderUserList(w, r, users, len(users))
}

// GetActiveVideoID 現在監視中のvideoIDを取得します
func (h *MonitoringHandler) GetActiveVideoID(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))

	videoID, isActive, exists := h.chatMonitoringUC.GetActiveVideoID()
	if !exists {
    h.logger.LogAPI(constants.LogLevelInfo, "No monitoring session found", "", correlationID, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusNotFound, "No monitoring session found", correlationID)
		return
	}

    h.logger.LogAPI(constants.LogLevelInfo, "Video ID retrieved", videoID, correlationID, map[string]interface{}{
		"isActive": isActive,
	})

	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{
		"videoId":  videoID,
		"isActive": isActive,
	}, correlationID)
}

// GetVideoStatus GET /api/monitoring/{videoId}/status を処理します
func (h *MonitoringHandler) GetVideoStatus(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	videoID := chi.URLParam(r, "videoId")

	if videoID == "" {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "video ID is required", correlationID)
		return
	}

    h.logger.LogAPI(constants.LogLevelInfo, "Get video status request received", videoID, correlationID, nil)

	status, err := h.chatMonitoringUC.GetVideoStatus(r.Context(), videoID)
    if err != nil {
        h.logger.LogError(constants.LogLevelError, "Failed to get video status", videoID, correlationID, err, nil)
        response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), correlationID)
        return
    }

	response.RenderSuccess(w, r, map[string]interface{}{
		"videoID": videoID,
		"status":  status,
	})
}

// GetActiveUserList GET /api/monitoring/users を処理します（アクティブセッションのユーザー一覧を直接返す）
// 現行UIはSSEでユーザー一覧取得のため、/api/monitoring/users は削除しました

// AutoEndToggleRequest 自動終了検知の切替リクエスト
type AutoEndToggleRequest struct {
	Enabled bool `json:"enabled"`
}

// GetAutoEndSetting GET /api/monitoring/auto-end 現在の自動終了検知状態を返します
func (h *MonitoringHandler) GetAutoEndSetting(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	videoID, enabled, err := h.chatMonitoringUC.IsAutoEndEnabled()
	if err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusNotFound, err.Error(), correlationID)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{
		"videoId": videoID,
		"enabled": enabled,
	}, correlationID)
}

// SetAutoEndSetting POST /api/monitoring/auto-end 自動終了検知の有効/無効を設定します
func (h *MonitoringHandler) SetAutoEndSetting(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("http-%s", r.Header.Get("requestId"))
	var req AutoEndToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "Invalid request body", correlationID)
		return
	}
	videoID, err := h.chatMonitoringUC.SetAutoEndEnabled(req.Enabled)
	if err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, err.Error(), correlationID)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{
		"videoId": videoID,
		"enabled": req.Enabled,
	}, correlationID)
}
