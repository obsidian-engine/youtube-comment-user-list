// Package logging 構造化ログ機能を提供します
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
)

// StructuredLogger 構造化ログ機能を持つLoggerインターフェースを実装します
type StructuredLogger struct {
	logger *log.Logger
}

// NewStructuredLogger 新しい構造化ロガーを作成します
func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		logger: log.New(os.Stdout, "", 0), // フォーマットを独自に処理するためプレフィックス/フラグなし
	}
}

// LogEntry 構造化ログエントリを表します
type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`
	Level         string                 `json:"level"`
	Component     string                 `json:"component,omitempty"`
	Event         string                 `json:"event,omitempty"`
	Message       string                 `json:"message"`
	VideoID       string                 `json:"video_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

// LogStructured 構造化メッセージをログ出力します
func (l *StructuredLogger) LogStructured(level, component, event, message, videoID, correlationID string, context map[string]interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().Format(constants.TimeFormatISO8601),
		Level:         level,
		Component:     component,
		Event:         event,
		Message:       message,
		VideoID:       videoID,
		CorrelationID: correlationID,
		Context:       context,
	}

	l.logEntry(entry)
}

// LogAPI API関連のイベントをログ出力します
func (l *StructuredLogger) LogAPI(level, message, videoID, correlationID string, context map[string]interface{}) {
	l.LogStructured(level, "api", "api_call", message, videoID, correlationID, context)
}

// LogPoller ポーリング関連のイベントをログ出力します
func (l *StructuredLogger) LogPoller(level, message, videoID, correlationID string, context map[string]interface{}) {
	l.LogStructured(level, "poller", "polling_event", message, videoID, correlationID, context)
}

// LogUser ユーザー関連のイベントをログ出力します
func (l *StructuredLogger) LogUser(level, message, videoID, correlationID string, context map[string]interface{}) {
	l.LogStructured(level, "user", "user_event", message, videoID, correlationID, context)
}

// LogError エラーイベントをログ出力します
func (l *StructuredLogger) LogError(level, message, videoID, correlationID string, err error, context map[string]interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().Format(constants.TimeFormatISO8601),
		Level:         level,
		Component:     "error",
		Event:         "error_occurred",
		Message:       message,
		VideoID:       videoID,
		CorrelationID: correlationID,
		Context:       context,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.logEntry(entry)
}

// logEntry ログエントリをJSON形式で出力します
func (l *StructuredLogger) logEntry(entry LogEntry) {
	// コンソール出力の場合、開発用により読みやすい形式を使用
	if l.isConsoleOutput() {
		l.logConsole(entry)
	} else {
		l.logJSON(entry)
	}
}

// logConsole コンソール用に人間が読みやすい形式でログを出力します
func (l *StructuredLogger) logConsole(entry LogEntry) {
	// フォーマット: [TIMESTAMP] LEVEL [COMPONENT] MESSAGE (videoId=..., correlationId=...)
	var output string

	timestamp := entry.Timestamp
	if entry.Level != "" {
		output = fmt.Sprintf("[%s] %s", timestamp, entry.Level)
	} else {
		output = fmt.Sprintf("[%s]", timestamp)
	}

	if entry.Component != "" {
		output = fmt.Sprintf("%s [%s]", output, entry.Component)
	}

	if entry.Event != "" {
		output = fmt.Sprintf("%s [%s]", output, entry.Event)
	}

	output = fmt.Sprintf("%s %s", output, entry.Message)

	// メタデータを追加
	var metadata []string
	if entry.VideoID != "" {
		metadata = append(metadata, fmt.Sprintf("videoId=%s", entry.VideoID))
	}
	if entry.CorrelationID != "" {
		metadata = append(metadata, fmt.Sprintf("correlationId=%s", entry.CorrelationID))
	}
	if entry.Error != "" {
		metadata = append(metadata, fmt.Sprintf("error=%s", entry.Error))
	}

	if len(metadata) > 0 {
		output = fmt.Sprintf("%s (%s)", output, fmt.Sprintf("%v", metadata))
	}

	// 存在する場合はコンテキストを追加
	if len(entry.Context) > 0 {
		contextJSON, _ := json.Marshal(entry.Context)
		output = fmt.Sprintf("%s context=%s", output, string(contextJSON))
	}

	l.logger.Println(output)
}

// logJSON 構造化ログシステム用にJSON形式でログを出力します
func (l *StructuredLogger) logJSON(entry LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// JSONマーシャリングが失敗した場合は簡単なログにフォールバック
		l.logger.Printf("JSON marshal error: %v, original message: %s", err, entry.Message)
		return
	}

	l.logger.Println(string(jsonBytes))
}

// isConsoleOutput コンソールフレンドリーなフォーマットを使用するかを判定します
// 環境変数や他の設定をチェックするように拡張できます
func (l *StructuredLogger) isConsoleOutput() bool {
	// 現在は開発用に常にコンソール出力を使用
	// 本番環境ではJSON形式を使用したい場合があります
	return true
}
