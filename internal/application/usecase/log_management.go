package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// LogEntry 詳細情報を含む構造化ログエントリを表します
type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`
	Level         string                 `json:"level"`
	Component     string                 `json:"component,omitempty"`
	Event         string                 `json:"event,omitempty"`
	Message       string                 `json:"message"`
	VideoID       string                 `json:"video_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// LogManagementUseCase アプリケーションのログワークフローを処理します
type LogManagementUseCase struct {
	logger    service.Logger
	logBuffer *LogBuffer
}

// LogBuffer 最近のログエントリの循環バッファを維持します
type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	current int
}

// NewLogManagementUseCase 新しいLogManagementUseCaseを作成します
func NewLogManagementUseCase(logger service.Logger, maxLogEntries int) *LogManagementUseCase {
	return &LogManagementUseCase{
		logger: logger,
		logBuffer: &LogBuffer{
			entries: make([]LogEntry, maxLogEntries),
			maxSize: maxLogEntries,
			current: 0,
		},
	}
}

// AddLogEntry 新しいログエントリをバッファに追加します
func (uc *LogManagementUseCase) AddLogEntry(level, component, event, message, videoID, correlationID string, context map[string]interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().Format(constants.TimeFormatLog),
		Level:         level,
		Component:     component,
		Event:         event,
		Message:       message,
		VideoID:       videoID,
		CorrelationID: correlationID,
		Context:       context,
	}

	uc.logBuffer.add(entry)

	// 基盤のロガーにもログ出力
	uc.logger.LogStructured(level, component, event, message, videoID, correlationID, context)
}

// GetRecentLogs オプションのフィルタリング付きで最近のログエントリを返します
func (uc *LogManagementUseCase) GetRecentLogs(ctx context.Context, filters LogFilters) ([]LogEntry, error) {
	uc.logBuffer.mu.RLock()
	defer uc.logBuffer.mu.RUnlock()

	var filteredLogs []LogEntry

	// 空でないエントリを全て取得
	for i := 0; i < uc.logBuffer.maxSize; i++ {
		idx := (uc.logBuffer.current - 1 - i + uc.logBuffer.maxSize) % uc.logBuffer.maxSize
		entry := uc.logBuffer.entries[idx]

		// 空のエントリをスキップ（バッファがまだ満杯でない場合）
		if entry.Timestamp == "" {
			continue
		}

		// フィルタを適用
		if uc.matchesFilters(entry, filters) {
			filteredLogs = append(filteredLogs, entry)
		}

		// 結果を制限
		if filters.Limit > 0 && len(filteredLogs) >= filters.Limit {
			break
		}
	}

	return filteredLogs, nil
}

// LogFilters ログ取得のフィルタリング基準を定義します
type LogFilters struct {
	Level         string // ログレベルでフィルタ
	VideoID       string // 動画IDでフィルタ
	Component     string // コンポーネントでフィルタ
	CorrelationID string // 相関IDでフィルタ
	Limit         int    // 返すエントリの最大数
}

// GetLogStats returns statistics about logged events
func (uc *LogManagementUseCase) GetLogStats(ctx context.Context) (map[string]interface{}, error) {
	uc.logBuffer.mu.RLock()
	defer uc.logBuffer.mu.RUnlock()

	stats := map[string]interface{}{
		"totalEntries":    0,
		"levelCounts":     make(map[string]int),
		"componentCounts": make(map[string]int),
		"videoIdCounts":   make(map[string]int),
	}

	levelCounts := make(map[string]int)
	componentCounts := make(map[string]int)
	videoIdCounts := make(map[string]int)
	totalEntries := 0

	for i := 0; i < uc.logBuffer.maxSize; i++ {
		entry := uc.logBuffer.entries[i]
		if entry.Timestamp == "" {
			continue
		}

		totalEntries++
		levelCounts[entry.Level]++
		if entry.Component != "" {
			componentCounts[entry.Component]++
		}
		if entry.VideoID != "" {
			videoIdCounts[entry.VideoID]++
		}
	}

	stats["totalEntries"] = totalEntries
	stats["levelCounts"] = levelCounts
	stats["componentCounts"] = componentCounts
	stats["videoIdCounts"] = videoIdCounts

	return stats, nil
}

// ClearLogs clears all log entries from the buffer
func (uc *LogManagementUseCase) ClearLogs(ctx context.Context) error {
	uc.logBuffer.mu.Lock()
	defer uc.logBuffer.mu.Unlock()

	// Reset the buffer
	uc.logBuffer.entries = make([]LogEntry, uc.logBuffer.maxSize)
	uc.logBuffer.current = 0

	// Log the clear action
	uc.AddLogEntry("INFO", "log_management", "logs_cleared", "Log buffer cleared by user request", "", "", nil)

	return nil
}

// ExportLogs exports logs in JSON format
func (uc *LogManagementUseCase) ExportLogs(ctx context.Context, filters LogFilters) (string, error) {
	logs, err := uc.GetRecentLogs(ctx, filters)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	jsonData, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal logs to JSON: %w", err)
	}

	return string(jsonData), nil
}

// add adds an entry to the circular buffer
func (lb *LogBuffer) add(entry LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.entries[lb.current] = entry
	lb.current = (lb.current + 1) % lb.maxSize
}

// matchesFilters checks if a log entry matches the given filters
func (uc *LogManagementUseCase) matchesFilters(entry LogEntry, filters LogFilters) bool {
	if filters.Level != "" && entry.Level != filters.Level {
		return false
	}

	if filters.VideoID != "" && entry.VideoID != filters.VideoID {
		return false
	}

	if filters.Component != "" && entry.Component != filters.Component {
		return false
	}

	if filters.CorrelationID != "" && entry.CorrelationID != filters.CorrelationID {
		return false
	}

	return true
}

// LogAPI logs API-related events with structured data
func (uc *LogManagementUseCase) LogAPI(level, message, videoID, correlationID string, context map[string]interface{}) {
	uc.AddLogEntry(level, "api", "api_call", message, videoID, correlationID, context)
}

// LogPoller logs polling-related events
func (uc *LogManagementUseCase) LogPoller(level, message, videoID, correlationID string, context map[string]interface{}) {
	uc.AddLogEntry(level, "poller", "polling_event", message, videoID, correlationID, context)
}

// LogUser logs user-related events
func (uc *LogManagementUseCase) LogUser(level, message, videoID, correlationID string, context map[string]interface{}) {
	uc.AddLogEntry(level, "user", "user_event", message, videoID, correlationID, context)
}

// LogError logs error events with error details
func (uc *LogManagementUseCase) LogError(level, message, videoID, correlationID string, err error, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]interface{})
	}
	if err != nil {
		context["error"] = err.Error()
	}
	uc.AddLogEntry(level, "error", "error_occurred", message, videoID, correlationID, context)
}
