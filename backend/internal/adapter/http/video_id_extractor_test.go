package http

import (
	"testing"
)

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "ライブチャットURL（通常）",
			input:    "https://www.youtube.com/live_chat?is_popout=1&v=Qw3tyIFqKrg",
			expected: "Qw3tyIFqKrg",
			hasError: false,
		},
		{
			name:     "YouTube Studio ライブチャットURL",
			input:    "https://studio.youtube.com/live_chat?is_popout=1&v=bIRpAmqwbvs",
			expected: "bIRpAmqwbvs",
			hasError: false,
		},
		{
			name:     "通常の動画URL",
			input:    "https://www.youtube.com/watch?v=Qw3tyIFqKrg",
			expected: "Qw3tyIFqKrg",
			hasError: false,
		},
		{
			name:     "短縮URL",
			input:    "https://youtu.be/Qw3tyIFqKrg",
			expected: "Qw3tyIFqKrg",
			hasError: false,
		},
		{
			name:     "video_idのみ（既存動作維持）",
			input:    "Qw3tyIFqKrg",
			expected: "Qw3tyIFqKrg",
			hasError: false,
		},
		{
			name:     "ライブチャットURL（パラメータ順序違い）",
			input:    "https://www.youtube.com/live_chat?v=Qw3tyIFqKrg&is_popout=1",
			expected: "Qw3tyIFqKrg",
			hasError: false,
		},
		{
			name:     "埋め込みURL",
			input:    "https://www.youtube.com/embed/Qw3tyIFqKrg",
			expected: "Qw3tyIFqKrg",
			hasError: false,
		},
		{
			name:     "無効なURL",
			input:    "invalid-url",
			expected: "invalid-url",
			hasError: false,
		},
		{
			name:     "YouTube以外のURL",
			input:    "https://example.com/watch?v=Qw3tyIFqKrg",
			expected: "",
			hasError: true,
		},
		{
			name:     "空文字",
			input:    "",
			expected: "",
			hasError: true,
		},
		{
			name:     "vパラメータなし",
			input:    "https://www.youtube.com/live_chat?is_popout=1",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractVideoID(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("期待するエラーが発生しませんでした。input: %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v, input: %s", err, tt.input)
				}
				if result != tt.expected {
					t.Errorf("期待値: %s, 実際: %s, input: %s", tt.expected, result, tt.input)
				}
			}
		})
	}
}