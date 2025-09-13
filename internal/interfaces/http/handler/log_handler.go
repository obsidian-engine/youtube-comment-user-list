// Package handler HTTPハンドラーの実装を提供します
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/interfaces/http/response"
)

// LogHandler ログ管理用のハンドラー
type LogHandler struct {
	logManagementUC *usecase.LogManagementUseCase
	logger          repository.Logger
}

// NewLogHandler 新しいログハンドラーを作成します
func NewLogHandler(logManagementUC *usecase.LogManagementUseCase, logger repository.Logger) *LogHandler {
	return &LogHandler{logManagementUC: logManagementUC, logger: logger}
}

// buildFilters クエリからフィルタ生成
func (h *LogHandler) buildFilters(r *http.Request) usecase.LogFilters {
	q := r.URL.Query()
	limit := 300
	if ls := q.Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 {
			limit = v
		}
	}
	return usecase.LogFilters{
		Level:         strings.TrimSpace(q.Get("level")),
		Component:     strings.TrimSpace(q.Get("component")),
		VideoID:       strings.TrimSpace(q.Get("video_id")),
		CorrelationID: strings.TrimSpace(q.Get("correlation_id")),
		Limit:         limit,
	}
}

// GetLogs GET /api/logs 取得 or stats / export
func (h *LogHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	q := r.URL.Query()

	// stats モード
	if q.Get("stats") == "1" {
		stats, err := h.logManagementUC.GetLogStats(r.Context())
		if err != nil {
			h.logger.LogError(constants.LogLevelError, "Failed to get log stats", "", cid, err, nil)
			response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
			return
		}
		response.RenderSuccessWithCorrelation(w, r, stats, cid)
		return
	}

	// export モード
	if q.Get("export") == "1" {
		filters := h.buildFilters(r)
		data, err := h.logManagementUC.ExportLogs(r.Context(), filters)
		if err != nil {
			h.logger.LogError(constants.LogLevelError, "Failed to export logs", "", cid, err, nil)
			response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
			return
		}
		filename := "logs.json"
		if filters.VideoID != "" {
			filename = "logs_" + filters.VideoID + ".json"
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		_, _ = w.Write([]byte(data))
		return
	}

	// 通常取得
	filters := h.buildFilters(r)
	logs, err := h.logManagementUC.GetRecentLogs(r.Context(), filters)
	if err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to get logs", "", cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	}, cid)
}

// ClearLogs DELETE /api/logs
func (h *LogHandler) ClearLogs(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	if err := h.logManagementUC.ClearLogs(r.Context()); err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to clear logs", "", cid, err, nil)
		response.RenderErrorWithCorrelation(w, r, http.StatusInternalServerError, err.Error(), cid)
		return
	}
	response.RenderSuccessWithCorrelation(w, r, map[string]string{"message": "All logs cleared"}, cid)
}
