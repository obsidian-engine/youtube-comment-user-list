package logging

import (
	"context"
	"fmt"
	"log/slog"
)

// ModuleLogger は特定のモジュール用の構造化ログ機能を提供します
type ModuleLogger struct {
	module string
	logger *slog.Logger
}

// NewModuleLogger は新しいModuleLoggerインスタンスを作成します
func NewModuleLogger(module string, logger *slog.Logger) ModuleLogger {
	return ModuleLogger{
		module: module,
		logger: logger,
	}
}

// Info はINFOレベルのログを出力します
func (ml ModuleLogger) Info(msg string, args ...any) {
	ml.logger.Info(fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// InfoContext はコンテキスト付きでINFOレベルのログを出力します
func (ml ModuleLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	ml.logger.InfoContext(ctx, fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// Error はERRORレベルのログを出力します
func (ml ModuleLogger) Error(msg string, args ...any) {
	ml.logger.Error(fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// ErrorContext はコンテキスト付きでERRORレベルのログを出力します
func (ml ModuleLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	ml.logger.ErrorContext(ctx, fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// Debug はDEBUGレベルのログを出力します
func (ml ModuleLogger) Debug(msg string, args ...any) {
	ml.logger.Debug(fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// DebugContext はコンテキスト付きでDEBUGレベルのログを出力します
func (ml ModuleLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	ml.logger.DebugContext(ctx, fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// Warn はWARNレベルのログを出力します
func (ml ModuleLogger) Warn(msg string, args ...any) {
	ml.logger.Warn(fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}

// WarnContext はコンテキスト付きでWARNレベルのログを出力します
func (ml ModuleLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	ml.logger.WarnContext(ctx, fmt.Sprintf("[%s] %s", ml.module, msg), args...)
}