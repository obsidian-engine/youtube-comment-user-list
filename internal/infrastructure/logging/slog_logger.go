// Package logging slog対応の構造化ログ機能を提供します
package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// SlogLogger slogを使用した構造化Logger実装
type SlogLogger struct {
	logger    *slog.Logger
	requestID string
	baseAttrs []slog.Attr
}

// NewSlogLogger 新しいslog対応ロガーを作成します
func NewSlogLogger() repository.Logger {
	// JSONハンドラーでslogを設定
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false, // 本番では false を推奨
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &SlogLogger{
		logger:    logger,
		baseAttrs: make([]slog.Attr, 0),
	}
}

// WithRequestID リクエストIDを含むロガーを返します
func (l *SlogLogger) WithRequestID(requestID string) repository.Logger {
	newAttrs := append(l.baseAttrs, slog.String("request_id", requestID))
	return &SlogLogger{
		logger:    l.logger,
		requestID: requestID,
		baseAttrs: newAttrs,
	}
}

// WithContext コンテキストを含むロガーを返します
func (l *SlogLogger) WithContext(ctx context.Context) repository.Logger {
	// contextから追加情報を取得する場合はここで実装
	return l
}

// LogStructured 構造化メッセージをログ出力します
func (l *SlogLogger) LogStructured(level, component, event, message, videoID, correlationID string, context map[string]interface{}) {
	attrs := l.buildAttrs(component, event, videoID, correlationID, context)
	l.logWithLevel(level, message, attrs...)
}

// LogAPI API関連のイベントをログ出力します
func (l *SlogLogger) LogAPI(level, message, videoID, correlationID string, context map[string]interface{}) {
	attrs := l.buildAttrs("api", "", videoID, correlationID, context)
	l.logWithLevel(level, message, attrs...)
}

// LogPoller ポーリング関連のイベントをログ出力します
func (l *SlogLogger) LogPoller(level, message, videoID, correlationID string, context map[string]interface{}) {
	attrs := l.buildAttrs("poller", "", videoID, correlationID, context)
	l.logWithLevel(level, message, attrs...)
}

// LogUser ユーザー関連のイベントをログ出力します
func (l *SlogLogger) LogUser(level, message, videoID, correlationID string, context map[string]interface{}) {
	attrs := l.buildAttrs("user", "", videoID, correlationID, context)
	l.logWithLevel(level, message, attrs...)
}

// LogError エラーイベントをログ出力します
func (l *SlogLogger) LogError(level, message, videoID, correlationID string, err error, context map[string]interface{}) {
	attrs := l.buildAttrs("error", "", videoID, correlationID, context)
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	l.logWithLevel(level, message, attrs...)
}

// buildAttrs 共通属性を構築します
func (l *SlogLogger) buildAttrs(component, event, videoID, correlationID string, context map[string]interface{}) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(l.baseAttrs)+6+len(context))

	// ベース属性を追加
	attrs = append(attrs, l.baseAttrs...)

	// 標準属性を追加
	if component != "" {
		attrs = append(attrs, slog.String("component", component))
	}
	if event != "" {
		attrs = append(attrs, slog.String("event", event))
	}
	if videoID != "" {
		attrs = append(attrs, slog.String("video_id", videoID))
	}
	if correlationID != "" {
		attrs = append(attrs, slog.String("correlation_id", correlationID))
	}

	// コンテキスト属性を追加
	for key, value := range context {
		attrs = append(attrs, slog.Any(key, value))
	}

	return attrs
}

// logWithLevel レベルに応じてログ出力します
func (l *SlogLogger) logWithLevel(level, message string, attrs ...slog.Attr) {
	ctx := context.Background()

	switch level {
	case "DEBUG":
		l.logger.LogAttrs(ctx, slog.LevelDebug, message, attrs...)
	case "INFO":
		l.logger.LogAttrs(ctx, slog.LevelInfo, message, attrs...)
	case "WARN", "WARNING":
		l.logger.LogAttrs(ctx, slog.LevelWarn, message, attrs...)
	case "ERROR":
		l.logger.LogAttrs(ctx, slog.LevelError, message, attrs...)
	case "FATAL":
		l.logger.LogAttrs(ctx, slog.LevelError, message, attrs...)
		// FATALの場合は終了処理が必要なら実装
	default:
		// デフォルトはINFO
		l.logger.LogAttrs(ctx, slog.LevelInfo, message, attrs...)
	}
}
