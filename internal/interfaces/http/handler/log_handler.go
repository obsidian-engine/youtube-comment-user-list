// Package handler HTTPハンドラーの実装を提供します
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
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
func (h *LogHandler) GetLogs(c *gin.Context) {
	correlationID := fmt.Sprintf("logs-%s", c.GetString("requestId"))

	h.logger.LogAPI("INFO", "Get logs request received", "", correlationID, nil)

	// Parse query parameters for filtering
	filters := h.parseLogFiltersFromGin(c)

	logs, err := h.logManagementUC.GetRecentLogs(c.Request.Context(), filters)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get logs", "", correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
		"filters": filters,
	})
}

// GetLogStats handles GET /api/logs/stats
func (h *LogHandler) GetLogStats(c *gin.Context) {
	correlationID := fmt.Sprintf("logs-stats-%s", c.GetString("requestId"))

	h.logger.LogAPI("INFO", "Get log stats request received", "", correlationID, nil)

	stats, err := h.logManagementUC.GetLogStats(c.Request.Context())
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get log stats", "", correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	// Extract values from map
	totalEntries, _ := stats["totalEntries"].(int)
	levelCounts, _ := stats["levelCounts"].(map[string]int)

	var errorCount, warningCount int
	if levelCounts != nil {
		errorCount = levelCounts["ERROR"]
		warningCount = levelCounts["WARNING"]
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"total":    totalEntries,
		"errors":   errorCount,
		"warnings": warningCount,
	})
}

// ClearLogs handles DELETE /api/logs
func (h *LogHandler) ClearLogs(c *gin.Context) {
	correlationID := fmt.Sprintf("logs-clear-%s", c.GetString("requestId"))

	h.logger.LogAPI("INFO", "Clear logs request received", "", correlationID, nil)

	err := h.logManagementUC.ClearLogs(c.Request.Context())
	if err != nil {
		h.logger.LogError("ERROR", "Failed to clear logs", "", correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All logs cleared successfully",
	})
}

// ExportLogs handles GET /api/logs/export
func (h *LogHandler) ExportLogs(c *gin.Context) {
	correlationID := fmt.Sprintf("logs-export-%s", c.GetString("requestId"))

	h.logger.LogAPI("INFO", "Export logs request received", "", correlationID, nil)

	// Parse query parameters for filtering
	filters := h.parseLogFiltersFromGin(c)

	exportData, err := h.logManagementUC.ExportLogs(c.Request.Context(), filters)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to export logs", "", correlationID, err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "correlationID": correlationID})
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=\"logs_export.json\"")

	c.String(http.StatusOK, exportData)
}

// parseLogFiltersFromGin Ginからクエリパラメータを解析してログフィルターを作成します
func (h *LogHandler) parseLogFiltersFromGin(c *gin.Context) usecase.LogFilters {
	filters := usecase.LogFilters{}

	if level := c.Query("level"); level != "" {
		filters.Level = level
	}

	if videoID := c.Query("video_id"); videoID != "" {
		filters.VideoID = videoID
	}

	if component := c.Query("component"); component != "" {
		filters.Component = component
	}

	if correlationID := c.Query("correlation_id"); correlationID != "" {
		filters.CorrelationID = correlationID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	// Default limit if not specified
	if filters.Limit == constants.MinValidLimit {
		filters.Limit = constants.DefaultLogDisplayLimit
	}

	return filters
}

// Helper methods
