// Package logging slog対応の構造化ログ機能を提供します
package logging

import (
    "context"
    "log/slog"
    "os"
    "strings"

    "github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// SlogLogger slogを使用した構造化Logger実装
type SlogLogger struct {
    logger    *slog.Logger
    requestID string
    baseAttrs []slog.Attr
    sink      func(level, component, event, message, videoID, correlationID string, context map[string]interface{})
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

// NewSlogLoggerWithLevel ログレベル指定付きのロガーを作成します
func NewSlogLoggerWithLevel(levelStr string) repository.Logger {
    // JSONハンドラーでslogを設定
    lvl := parseLevel(levelStr)
    opts := &slog.HandlerOptions{
        Level:     lvl,
        AddSource: false,
    }
    handler := slog.NewJSONHandler(os.Stdout, opts)
    logger := slog.New(handler)
    return &SlogLogger{logger: logger, baseAttrs: make([]slog.Attr, 0)}
}

// SetSink ログ集約先（循環バッファ等）を登録します
func (l *SlogLogger) SetSink(fn func(level, component, event, message, videoID, correlationID string, context map[string]interface{})) {
    l.sink = fn
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
    norm := normalizeLevelName(level)
    attrs := l.buildAttrs(component, event, videoID, correlationID, context)
    l.logWithLevel(norm, message, attrs...)
    if l.sink != nil {
        l.sink(norm, component, event, message, videoID, correlationID, context)
    }
}

// LogAPI API関連のイベントをログ出力します
func (l *SlogLogger) LogAPI(level, message, videoID, correlationID string, context map[string]interface{}) {
    norm := normalizeLevelName(level)
    attrs := l.buildAttrs("api", "", videoID, correlationID, context)
    l.logWithLevel(norm, message, attrs...)
    if l.sink != nil {
        l.sink(norm, "api", "", message, videoID, correlationID, context)
    }
}

// LogPoller ポーリング関連のイベントをログ出力します
func (l *SlogLogger) LogPoller(level, message, videoID, correlationID string, context map[string]interface{}) {
    norm := normalizeLevelName(level)
    attrs := l.buildAttrs("poller", "", videoID, correlationID, context)
    l.logWithLevel(norm, message, attrs...)
    if l.sink != nil {
        l.sink(norm, "poller", "", message, videoID, correlationID, context)
    }
}

// LogUser ユーザー関連のイベントをログ出力します
func (l *SlogLogger) LogUser(level, message, videoID, correlationID string, context map[string]interface{}) {
    norm := normalizeLevelName(level)
    attrs := l.buildAttrs("user", "", videoID, correlationID, context)
    l.logWithLevel(norm, message, attrs...)
    if l.sink != nil {
        l.sink(norm, "user", "", message, videoID, correlationID, context)
    }
}

// LogError エラーイベントをログ出力します
func (l *SlogLogger) LogError(level, message, videoID, correlationID string, err error, context map[string]interface{}) {
    norm := normalizeLevelName(level)
    attrs := l.buildAttrs("error", "", videoID, correlationID, context)
    if err != nil {
        attrs = append(attrs, slog.String("error", err.Error()))
    }
    l.logWithLevel(norm, message, attrs...)
    if l.sink != nil {
        if context == nil {
            context = map[string]interface{}{}
        }
        context["error"] = err
        l.sink(norm, "error", "", message, videoID, correlationID, context)
    }
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
    case constants.LogLevelDebug:
        l.logger.LogAttrs(ctx, slog.LevelDebug, message, attrs...)
    case constants.LogLevelInfo:
        l.logger.LogAttrs(ctx, slog.LevelInfo, message, attrs...)
    case constants.LogLevelWarning:
        l.logger.LogAttrs(ctx, slog.LevelWarn, message, attrs...)
    case constants.LogLevelError:
        l.logger.LogAttrs(ctx, slog.LevelError, message, attrs...)
    case constants.LogLevelFatal:
        l.logger.LogAttrs(ctx, slog.LevelError, message, attrs...)
        // FATALの場合は終了処理が必要なら実装
    default:
        // デフォルトはINFO
        l.logger.LogAttrs(ctx, slog.LevelInfo, message, attrs...)
    }
}

// normalizeLevelName レベル名を統一（WARN→WARNING など）
func normalizeLevelName(level string) string {
    switch strings.ToUpper(level) {
    case "WARN", constants.LogLevelWarning:
        return constants.LogLevelWarning
    case constants.LogLevelError:
        return constants.LogLevelError
    case constants.LogLevelDebug:
        return constants.LogLevelDebug
    default:
        return constants.LogLevelInfo
    }
}

// parseLevel 文字列から slog レベルを決定
func parseLevel(level string) slog.Leveler {
    switch strings.ToUpper(level) {
    case constants.LogLevelDebug:
        return slog.LevelDebug
    case "WARN", constants.LogLevelWarning:
        return slog.LevelWarn
    case constants.LogLevelError:
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
