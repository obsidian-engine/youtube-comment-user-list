// Package handler HTTPハンドラーの実装を提供します
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// LogHandler HTTP requests for log managementを処理します
type LogHandler struct {
	logManagementUC *usecase.LogManagementUseCase
	logger          service.Logger
}

// NewLogHandler 新しいlogを作成します handler
func NewLogHandler(
	logManagementUC *usecase.LogManagementUseCase,
	logger service.Logger,
) *LogHandler {
	return &LogHandler{
		logManagementUC: logManagementUC,
		logger:          logger,
	}
}

// GetLogs handles GET /api/logs
func (h *LogHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("logs-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get logs request received", "", correlationID, nil)

	// Parse query parameters for filtering
	filters := h.parseLogFilters(r)

	logs, err := h.logManagementUC.GetRecentLogs(r.Context(), filters)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get logs", "", correlationID, err, nil)
		h.respondWithError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
		"filters": filters,
	}

	h.respondWithJSON(w, http.StatusOK, response, "", correlationID)
}

// GetLogStats handles GET /api/logs/stats
func (h *LogHandler) GetLogStats(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("log-stats-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Get log stats request received", "", correlationID, nil)

	stats, err := h.logManagementUC.GetLogStats(r.Context())
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get log stats", "", correlationID, err, nil)
		h.respondWithError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	h.respondWithJSON(w, http.StatusOK, response, "", correlationID)
}

// ClearLogs handles DELETE /api/logs
func (h *LogHandler) ClearLogs(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("clear-logs-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodDelete {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Clear logs request received", "", correlationID, nil)

	if err := h.logManagementUC.ClearLogs(r.Context()); err != nil {
		h.logger.LogError("ERROR", "Failed to clear logs", "", correlationID, err, nil)
		h.respondWithError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logs cleared successfully",
	}

	h.respondWithJSON(w, http.StatusOK, response, "", correlationID)
}

// ExportLogs handles GET /api/logs/export
func (h *LogHandler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("export-logs-%d", r.Context().Value("requestId"))

	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", correlationID)
		return
	}

	h.logger.LogAPI("INFO", "Export logs request received", "", correlationID, nil)

	// Parse query parameters for filtering
	filters := h.parseLogFilters(r)

	jsonData, err := h.logManagementUC.ExportLogs(r.Context(), filters)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to export logs", "", correlationID, err, nil)
		h.respondWithError(w, http.StatusInternalServerError, err.Error(), correlationID)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=logs.json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(jsonData)); err != nil {
		h.logger.LogError("ERROR", "Failed to write export response", "", correlationID, err, nil)
	}

	h.logger.LogAPI("INFO", "Logs exported successfully", "", correlationID, map[string]interface{}{
		"filters": filters,
	})
}

// parseLogFilters parses query parameters into log filters
func (h *LogHandler) parseLogFilters(r *http.Request) usecase.LogFilters {
	filters := usecase.LogFilters{}

	// Parse query parameters
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

	// Default limit if not specified
	if filters.Limit == 0 {
		filters.Limit = 100
	}

	return filters
}

// Helper methods

func (h *LogHandler) respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, videoID, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON response", videoID, correlationID, err, nil)
	}

	h.logger.LogAPI("DEBUG", "Response sent", videoID, correlationID, map[string]interface{}{
		"statusCode": statusCode,
	})
}

func (h *LogHandler) respondWithError(w http.ResponseWriter, statusCode int, message, correlationID string) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.LogError("ERROR", "Failed to encode error response", "", correlationID, err, nil)
	}

	h.logger.LogAPI("ERROR", "Error response sent", "", correlationID, map[string]interface{}{
		"statusCode": statusCode,
		"error":      message,
	})
}
