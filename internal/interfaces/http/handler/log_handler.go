// Package handler HTTPハンドラーの実装を提供します
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// LogHandler ログ管理のHTTPリクエストを処理します
type LogHandler struct {
	logManagementUC *usecase.LogManagementUseCase
	logger          repository.Logger
}

// Handle 統合ログエンドポイント
// - GET /api/logs?stats=1 で統計
// - GET /api/logs?export=1 でエクスポート
// - GET /api/logs でログ一覧
// - DELETE /api/logs で全クリア
func (h *LogHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		if q.Get("stats") == "1" {
			h.GetLogStats(w, r)
			return
		}
		if q.Get("export") == "1" {
			h.ExportLogs(w, r)
			return
		}
		h.GetLogs(w, r)
		return
	case http.MethodDelete:
		h.ClearLogs(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("method not allowed"))
	}
}

// NewLogHandler 新しいログハンドラーを作成します
func NewLogHandler(
	logManagementUC *usecase.LogManagementUseCase,
	logger repository.Logger,
) *LogHandler {
	return &LogHandler{
		logManagementUC: logManagementUC,
		logger:          logger,
	}
}

// GetLogs GET /api/logs を処理します
func (h *LogHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("logs-%s", r.Header.Get("requestId"))

	h.logger.LogAPI("INFO", "Get logs request received", "", correlationID, nil)

	// フィルタリング用のクエリパラメータを解析
	filters := h.parseLogFilters(r)

	logs, err := h.logManagementUC.GetRecentLogs(r.Context(), filters)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get logs", "", correlationID, err, nil)
		h.writeJSONError(w, err.Error(), correlationID)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
		"filters": filters,
	})
}

// GetLogStats GET /api/logs/stats を処理します
func (h *LogHandler) GetLogStats(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("logs-stats-%s", r.Header.Get("requestId"))

	h.logger.LogAPI("INFO", "Get log stats request received", "", correlationID, nil)

	stats, err := h.logManagementUC.GetLogStats(r.Context())
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get log stats", "", correlationID, err, nil)
		h.writeJSONError(w, err.Error(), correlationID)
		return
	}

	// マップから値を抽出
	totalEntries, _ := stats["totalEntries"].(int)
	levelCounts, _ := stats["levelCounts"].(map[string]int)

	var errorCount, warningCount int
	if levelCounts != nil {
		errorCount = levelCounts["ERROR"]
		warningCount = levelCounts["WARNING"]
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"total":    totalEntries,
		"errors":   errorCount,
		"warnings": warningCount,
	})
}

// ClearLogs DELETE /api/logs を処理します
func (h *LogHandler) ClearLogs(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("logs-clear-%s", r.Header.Get("requestId"))

	h.logger.LogAPI("INFO", "Clear logs request received", "", correlationID, nil)

	err := h.logManagementUC.ClearLogs(r.Context())
	if err != nil {
		h.logger.LogError("ERROR", "Failed to clear logs", "", correlationID, err, nil)
		h.writeJSONError(w, err.Error(), correlationID)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "All logs cleared successfully",
	})
}

// ExportLogs GET /api/logs/export を処理します
func (h *LogHandler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("logs-export-%s", r.Header.Get("requestId"))

	h.logger.LogAPI("INFO", "Export logs request received", "", correlationID, nil)

	// フィルタリング用のクエリパラメータを解析
	filters := h.parseLogFilters(r)

	exportData, err := h.logManagementUC.ExportLogs(r.Context(), filters)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to export logs", "", correlationID, err, nil)
		h.writeJSONError(w, err.Error(), correlationID)
		return
	}

	// ファイルダウンロード用のヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"logs_export.json\"")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(exportData)); err != nil {
		h.logger.LogError("ERROR", "Failed to write export data", "", "", err, nil)
	}
}

// parseLogFilters HTTPリクエストからクエリパラメータを解析してログフィルターを作成します
func (h *LogHandler) parseLogFilters(r *http.Request) usecase.LogFilters {
	filters := usecase.LogFilters{}
	query := r.URL.Query()

	if level := query.Get("level"); level != "" {
		filters.Level = level
	}

	if videoID := query.Get("video_id"); videoID != "" {
		filters.VideoID = videoID
	}

	if component := query.Get("component"); component != "" {
		filters.Component = component
	}

	if correlationID := query.Get("correlation_id"); correlationID != "" {
		filters.CorrelationID = correlationID
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	// 指定されていない場合のデフォルト制限
	if filters.Limit == constants.MinValidLimit {
		filters.Limit = constants.DefaultLogDisplayLimit
	}

	return filters
}

// writeJSON はJSONレスポンスを書き込みます
func (h *LogHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON response", "", "", err, nil)
	}
}

// writeJSONError はJSONエラーレスポンスを書き込みます
func (h *LogHandler) writeJSONError(w http.ResponseWriter, message, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":         message,
		"correlationID": correlationID,
	}); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON error response", "", correlationID, err, nil)
	}
}
