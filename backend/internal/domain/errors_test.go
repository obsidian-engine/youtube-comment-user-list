package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name          string
		err           ValidationError
		expectedParts []string
	}{
		{
			name: "Without value",
			err:  NewValidationError("email", "invalid format"),
			expectedParts: []string{
				"validation failed",
				"field 'email'",
				"invalid format",
			},
		},
		{
			name: "With value",
			err:  NewValidationErrorWithValue("age", "must be positive", "-5"),
			expectedParts: []string{
				"validation failed",
				"field 'age'",
				"must be positive",
				"value: -5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := tt.err.Error()
			for _, part := range tt.expectedParts {
				if !strings.Contains(errorMsg, part) {
					t.Errorf("Error message should contain '%s', got: %s", part, errorMsg)
				}
			}
		})
	}
}

func TestYouTubeAPIError_Error(t *testing.T) {
	tests := []struct {
		name          string
		err           YouTubeAPIError
		expectedParts []string
	}{
		{
			name: "Without API code",
			err:  NewYouTubeAPIError("GetLiveChat", "rate limit exceeded", true),
			expectedParts: []string{
				"youtube api error",
				"GetLiveChat",
				"rate limit exceeded",
				"retry: true",
			},
		},
		{
			name: "With API code",
			err:  NewYouTubeAPIErrorWithCode("GetVideo", "not found", 404, false),
			expectedParts: []string{
				"youtube api error",
				"GetVideo",
				"not found",
				"code: 404",
				"retry: false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := tt.err.Error()
			for _, part := range tt.expectedParts {
				if !strings.Contains(errorMsg, part) {
					t.Errorf("Error message should contain '%s', got: %s", part, errorMsg)
				}
			}
		})
	}
}

func TestBusinessLogicError_Error(t *testing.T) {
	tests := []struct {
		name          string
		err           BusinessLogicError
		expectedParts []string
	}{
		{
			name: "Without context",
			err:  NewBusinessLogicError("SwitchVideo", "live stream not active"),
			expectedParts: []string{
				"business logic error",
				"SwitchVideo",
				"live stream not active",
			},
		},
		{
			name: "With context",
			err:  NewBusinessLogicErrorWithContext("Pull", "no messages available", "video_id=abc123"),
			expectedParts: []string{
				"business logic error",
				"Pull",
				"no messages available",
				"context: video_id=abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := tt.err.Error()
			for _, part := range tt.expectedParts {
				if !strings.Contains(errorMsg, part) {
					t.Errorf("Error message should contain '%s', got: %s", part, errorMsg)
				}
			}
		})
	}
}

func TestConfigurationError_Error(t *testing.T) {
	err := NewConfigurationError("youtube_api_key", "key is empty", true)
	
	errorMsg := err.Error()
	expectedParts := []string{
		"configuration error",
		"'youtube_api_key'",
		"key is empty",
		"fatal: true",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errorMsg, part) {
			t.Errorf("Error message should contain '%s', got: %s", part, errorMsg)
		}
	}
}

func TestCustomErrors_JSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		err  interface{}
	}{
		{
			name: "ValidationError",
			err:  NewValidationErrorWithValue("email", "invalid format", "not-an-email"),
		},
		{
			name: "YouTubeAPIError",
			err:  NewYouTubeAPIErrorWithCode("GetVideo", "not found", 404, false),
		},
		{
			name: "BusinessLogicError",
			err:  NewBusinessLogicErrorWithContext("Pull", "no messages", "video_id=123"),
		},
		{
			name: "ConfigurationError",
			err:  NewConfigurationError("api_key", "missing", true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.err)
			if err != nil {
				t.Errorf("Failed to marshal %s: %v", tt.name, err)
			}

			// デシリアライズしてみる
			var result map[string]interface{}
			if err := json.Unmarshal(jsonData, &result); err != nil {
				t.Errorf("Failed to unmarshal %s: %v", tt.name, err)
			}

			// 基本フィールドが存在することを確認
			if len(result) == 0 {
				t.Errorf("Serialized %s should not be empty", tt.name)
			}
		})
	}
}

func TestErrorTypes_ImplementErrorInterface(t *testing.T) {
	// 全てのカスタムエラー型がerror interfaceを実装していることを確認
	var _ error = ValidationError{}
	var _ error = YouTubeAPIError{}
	var _ error = BusinessLogicError{}
	var _ error = ConfigurationError{}

	// 実際にError()メソッドが呼び出せることを確認
	errors := []error{
		NewValidationError("test", "test message"),
		NewYouTubeAPIError("test", "test message", false),
		NewBusinessLogicError("test", "test reason"),
		NewConfigurationError("test", "test message", false),
	}

	for i, err := range errors {
		if errorStr := err.Error(); errorStr == "" {
			t.Errorf("Error %d returned empty string", i)
		}
	}
}