package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestModuleLogger_Info(t *testing.T) {
	// バッファーでログ出力をキャプチャ
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// モジュールロガーを作成
	moduleLogger := NewModuleLogger("TEST_MODULE", logger)

	// Info ログを実行
	moduleLogger.Info("test message", "key", "value")

	// 出力を検証
	output := buf.String()
	if !strings.Contains(output, "TEST_MODULE") {
		t.Errorf("Expected log to contain 'TEST_MODULE', got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected log to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected log to contain 'key=value', got: %s", output)
	}
}

func TestModuleLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	moduleLogger := NewModuleLogger("ERROR_MODULE", logger)
	
	// Error ログを実行
	moduleLogger.Error("error occurred", "error_code", 500)

	output := buf.String()
	if !strings.Contains(output, "ERROR_MODULE") {
		t.Errorf("Expected log to contain 'ERROR_MODULE', got: %s", output)
	}
	if !strings.Contains(output, "error occurred") {
		t.Errorf("Expected log to contain 'error occurred', got: %s", output)
	}
	if !strings.Contains(output, "error_code=500") {
		t.Errorf("Expected log to contain 'error_code=500', got: %s", output)
	}
}

func TestModuleLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	moduleLogger := NewModuleLogger("CONTEXT_MODULE", logger)
	
	// コンテキスト付きでログを実行
	type requestIDKey string
	ctx := context.WithValue(context.Background(), requestIDKey("request_id"), "abc123")
	moduleLogger.InfoContext(ctx, "processing request", "user_id", "user456")

	output := buf.String()
	if !strings.Contains(output, "CONTEXT_MODULE") {
		t.Errorf("Expected log to contain 'CONTEXT_MODULE', got: %s", output)
	}
	if !strings.Contains(output, "processing request") {
		t.Errorf("Expected log to contain 'processing request', got: %s", output)
	}
}

func TestModuleLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(ModuleLogger, string)
		expected string
	}{
		{
			name: "Debug level",
			logFunc: func(ml ModuleLogger, msg string) {
				ml.Debug(msg)
			},
			expected: "level=DEBUG",
		},
		{
			name: "Info level", 
			logFunc: func(ml ModuleLogger, msg string) {
				ml.Info(msg)
			},
			expected: "level=INFO",
		},
		{
			name: "Warn level",
			logFunc: func(ml ModuleLogger, msg string) {
				ml.Warn(msg)
			},
			expected: "level=WARN",
		},
		{
			name: "Error level",
			logFunc: func(ml ModuleLogger, msg string) {
				ml.Error(msg)
			},
			expected: "level=ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			moduleLogger := NewModuleLogger("LEVEL_TEST", logger)
			tt.logFunc(moduleLogger, "test message")

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected log to contain '%s', got: %s", tt.expected, output)
			}
		})
	}
}