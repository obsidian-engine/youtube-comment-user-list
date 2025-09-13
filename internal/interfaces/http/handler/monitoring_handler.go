package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/interfaces/http/response"
)

// MonitoringHandler 監視機能のHTTPハンドラー
type MonitoringHandler struct {
	chatMonitoringUC *usecase.ChatMonitoringUseCase
	logger           repository.Logger
}

func NewMonitoringHandler(chatMonitoringUC *usecase.ChatMonitoringUseCase, logger repository.Logger) *MonitoringHandler {
	return &MonitoringHandler{chatMonitoringUC: chatMonitoringUC, logger: logger}
}

// StartMonitoringRequest ...
type StartMonitoringRequest struct {
	VideoInput string `json:"video_input"`
	MaxUsers   int    `json:"max_users,omitempty"`
}

// ResumeMonitoringRequest ...
type ResumeMonitoringRequest struct {
	VideoInput string `json:"video_input,omitempty"`
	MaxUsers   int    `json:"max_users,omitempty"`
}

// AutoEndToggleRequest ...
type AutoEndToggleRequest struct {
	Enabled bool `json:"enabled"`
}

// StartMonitoring POST /api/monitoring/start
func (h *MonitoringHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	h.logger.LogAPI(constants.LogLevelInfo, "Start monitoring request received", "", cid, map[string]interface{}{"method": r.Method, "path": r.URL.Path})
	var req StartMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogError(constants.LogLevelError, "Invalid request body", "", cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "Invalid request body", cid)
		return
	}
	if req.VideoInput == "" {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "video_input is required", cid)
		return
	}
	if req.MaxUsers <= 0 {
		req.MaxUsers = constants.DefaultMaxUsers
	}
	session, err := h.chatMonitoringUC.StartMonitoring(r.Context(), req.VideoInput, req.MaxUsers)
	if err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to start monitoring", req.VideoInput, cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
		return
	}
	h.logger.LogAPI(constants.LogLevelInfo, "Start monitoring response", session.VideoID, cid, map[string]interface{}{"success": true, "videoID": session.VideoID})
	response.RenderStartMonitoring(w, r, session.VideoID, "Monitoring started successfully")
}

// StopMonitoring DELETE /api/monitoring/stop
func (h *MonitoringHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	h.logger.LogAPI(constants.LogLevelInfo, "Stop monitoring request received", "", cid, map[string]interface{}{"method": r.Method, "path": r.URL.Path})
	if err := h.chatMonitoringUC.StopMonitoring(); err != nil {
		if err.Error() == "no active monitoring session" {
			h.logger.LogAPI(constants.LogLevelInfo, "Monitoring already stopped", "", cid, nil)
			response.RenderSuccessWithCorrelation(w, r, map[string]string{"message": "Monitoring already stopped"}, cid)
			return
		}
		h.logger.LogError(constants.LogLevelError, "Failed to stop monitoring", "", cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]string{"message": "Monitoring stopped successfully"}, cid)
}

// ResumeMonitoring POST /api/monitoring/resume
func (h *MonitoringHandler) ResumeMonitoring(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	h.logger.LogAPI(constants.LogLevelInfo, "Resume monitoring request received", "", cid, map[string]interface{}{"method": r.Method, "path": r.URL.Path})
	var req ResumeMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		h.logger.LogError(constants.LogLevelError, "Invalid request body", "", cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "Invalid request body", cid)
		return
	}
	if videoID, isActive, exists := h.chatMonitoringUC.GetActiveVideoID(); exists && isActive {
		h.logger.LogAPI(constants.LogLevelInfo, "Monitoring already active", videoID, cid, nil)
		response.RenderStartMonitoring(w, r, videoID, "Monitoring already active")
		return
	}
	maxUsers := req.MaxUsers
	if maxUsers <= 0 {
		maxUsers = constants.DefaultMaxUsers
	}
	if req.VideoInput == "" {
		session, err := h.chatMonitoringUC.ResumeMonitoring(r.Context(), maxUsers)
		if err != nil {
			response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, err.Error(), cid)
			return
		}
		response.RenderStartMonitoring(w, r, session.VideoID, "Monitoring resumed successfully")
		return
	}
	session, err := h.chatMonitoringUC.StartMonitoring(r.Context(), req.VideoInput, maxUsers)
	if err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, err.Error(), cid)
		return
	}
	response.RenderStartMonitoring(w, r, session.VideoID, "Monitoring started successfully")
}

// GetUserList GET /api/monitoring/{videoId}/users
func (h *MonitoringHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	videoID := chi.URLParam(r, "videoId")
	if videoID == "" {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "video ID is required", cid)
		return
	}
	h.logger.LogAPI(constants.LogLevelInfo, "Get user list request received", videoID, cid, nil)
	users, err := h.chatMonitoringUC.GetUserList(r.Context(), videoID)
	if err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to get user list", videoID, cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
		return
	}
	response.RenderUserList(w, r, users, len(users))
}

// GetActiveVideoID GET /api/monitoring/active
func (h *MonitoringHandler) GetActiveVideoID(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	videoID, isActive, exists := h.chatMonitoringUC.GetActiveVideoID()
	if !exists {
		h.logger.LogAPI(constants.LogLevelInfo, "No monitoring session found", "", cid, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusNotFound, "No monitoring session found", cid)
		return
	}
	h.logger.LogAPI(constants.LogLevelInfo, "Video ID retrieved", videoID, cid, map[string]interface{}{"isActive": isActive})
	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{"videoId": videoID, "isActive": isActive}, cid)
}

// GetVideoStatus GET /api/monitoring/{videoId}/status
func (h *MonitoringHandler) GetVideoStatus(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	videoID := chi.URLParam(r, "videoId")
	if videoID == "" {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "video ID is required", cid)
		return
	}
	h.logger.LogAPI(constants.LogLevelInfo, "Get video status request received", videoID, cid, nil)
	status, err := h.chatMonitoringUC.GetVideoStatus(r.Context(), videoID)
	if err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to get video status", videoID, cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
		return
	}
	response.RenderSuccess(w, r, map[string]interface{}{"videoID": videoID, "status": status})
}

// GetAutoEndSetting GET /api/monitoring/auto-end
func (h *MonitoringHandler) GetAutoEndSetting(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	videoID, enabled, err := h.chatMonitoringUC.IsAutoEndEnabled()
	if err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusNotFound, err.Error(), cid)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{"videoId": videoID, "enabled": enabled}, cid)
}

// SetAutoEndSetting POST /api/monitoring/auto-end
func (h *MonitoringHandler) SetAutoEndSetting(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	var req AutoEndToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, "Invalid request body", cid)
		return
	}
	videoID, err := h.chatMonitoringUC.SetAutoEndEnabled(req.Enabled)
	if err != nil {
		response.RenderErrorWithCorrelation(w, r, http.StatusBadRequest, err.Error(), cid)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{"videoId": videoID, "enabled": req.Enabled}, cid)
}
