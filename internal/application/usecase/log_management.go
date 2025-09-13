package usecase

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    "strings"

    "github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
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
	logger    repository.Logger
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
func NewLogManagementUseCase(logger repository.Logger, maxLogEntries int) *LogManagementUseCase {
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
    // レベル名を統一（WARN→WARNING など）
    var lvl string
    switch strings.ToUpper(level) {
    case "WARN", constants.LogLevelWarning:
        lvl = constants.LogLevelWarning
    case constants.LogLevelError:
        lvl = constants.LogLevelError
    case constants.LogLevelDebug:
        lvl = constants.LogLevelDebug
    default:
        lvl = constants.LogLevelInfo
    }

    entry := LogEntry{
        Timestamp:     time.Now().Format(constants.TimeFormatLog),
        Level:         lvl,
        Component:     component,
        Event:         event,
        Message:       message,
        VideoID:       videoID,
        CorrelationID: correlationID,
        Context:       context,
    }

	uc.logBuffer.add(entry)

    // ここでは基盤ロガーへは再出力しない（循環呼び出しを防止）
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

// GetLogStats ログイベントの統計情報を返します
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

// ClearLogs バッファからすべてのログエントリをクリアします
func (uc *LogManagementUseCase) ClearLogs(ctx context.Context) error {
	uc.logBuffer.mu.Lock()
	defer uc.logBuffer.mu.Unlock()

	// バッファをリセット
	uc.logBuffer.entries = make([]LogEntry, uc.logBuffer.maxSize)
	uc.logBuffer.current = 0

	// クリアアクションをログに記録
    uc.AddLogEntry(constants.LogLevelInfo, "log_management", "logs_cleared", "Log buffer cleared by user request", "", "", nil)

	return nil
}

// ExportLogs ログをJSON形式でエクスポートします
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

// add 循環バッファにエントリを追加します
func (lb *LogBuffer) add(entry LogEntry) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.entries[lb.current] = entry
	lb.current = (lb.current + 1) % lb.maxSize
}

// matchesFilters ログエントリが指定されたフィルタに一致するかチェックします
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
