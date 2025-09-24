package domain

import "fmt"

// ValidationError はバリデーションエラーを表すカスタムエラー型
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (v ValidationError) Error() string {
	if v.Value != "" {
		return fmt.Sprintf("validation failed for field '%s': %s (value: %s)", v.Field, v.Message, v.Value)
	}
	return fmt.Sprintf("validation failed for field '%s': %s", v.Field, v.Message)
}

// YouTubeAPIError はYouTube API関連のエラーを表すカスタムエラー型
type YouTubeAPIError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	APICode   int    `json:"api_code,omitempty"`
	Retry     bool   `json:"retry"`
}

func (y YouTubeAPIError) Error() string {
	if y.APICode != 0 {
		return fmt.Sprintf("youtube api error during %s: %s (code: %d, retry: %t)", 
			y.Operation, y.Message, y.APICode, y.Retry)
	}
	return fmt.Sprintf("youtube api error during %s: %s (retry: %t)", 
		y.Operation, y.Message, y.Retry)
}

// BusinessLogicError はビジネスロジックエラーを表すカスタムエラー型
type BusinessLogicError struct {
	Operation string `json:"operation"`
	Reason    string `json:"reason"`
	Context   string `json:"context,omitempty"`
}

func (b BusinessLogicError) Error() string {
	if b.Context != "" {
		return fmt.Sprintf("business logic error in %s: %s (context: %s)", 
			b.Operation, b.Reason, b.Context)
	}
	return fmt.Sprintf("business logic error in %s: %s", b.Operation, b.Reason)
}

// ConfigurationError は設定エラーを表すカスタムエラー型
type ConfigurationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Fatal   bool   `json:"fatal"`
}

func (c ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error for '%s': %s (fatal: %t)", 
		c.Field, c.Message, c.Fatal)
}

// エラーファクトリー関数

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

func NewValidationErrorWithValue(field, message, value string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

func NewYouTubeAPIError(operation, message string, retry bool) YouTubeAPIError {
	return YouTubeAPIError{
		Operation: operation,
		Message:   message,
		Retry:     retry,
	}
}

func NewYouTubeAPIErrorWithCode(operation, message string, apiCode int, retry bool) YouTubeAPIError {
	return YouTubeAPIError{
		Operation: operation,
		Message:   message,
		APICode:   apiCode,
		Retry:     retry,
	}
}

func NewBusinessLogicError(operation, reason string) BusinessLogicError {
	return BusinessLogicError{
		Operation: operation,
		Reason:    reason,
	}
}

func NewBusinessLogicErrorWithContext(operation, reason, context string) BusinessLogicError {
	return BusinessLogicError{
		Operation: operation,
		Reason:    reason,
		Context:   context,
	}
}

func NewConfigurationError(field, message string, fatal bool) ConfigurationError {
	return ConfigurationError{
		Field:   field,
		Message: message,
		Fatal:   fatal,
	}
}