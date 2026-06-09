package domain_test

import (
	"errors"
	"testing"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *domain.APIError
		want string
	}{
		{
			name: "quota_exceeded",
			err:  &domain.APIError{Code: domain.ErrCodeQuotaExceeded, Message: "daily quota exceeded"},
			want: "quota_exceeded: daily quota exceeded",
		},
		{
			name: "video_not_found",
			err:  &domain.APIError{Code: domain.ErrCodeVideoNotFound, Message: "video abc123 not found"},
			want: "video_not_found: video abc123 not found",
		},
		{
			name: "live_chat_ended",
			err:  &domain.APIError{Code: domain.ErrCodeLiveChatEnded, Message: "live chat has ended"},
			want: "live_chat_ended: live chat has ended",
		},
		{
			name: "auth_failed",
			err:  &domain.APIError{Code: domain.ErrCodeAuthFailed, Message: "invalid API key"},
			want: "auth_failed: invalid API key",
		},
		{
			name: "rate_limited",
			err:  &domain.APIError{Code: domain.ErrCodeRateLimited, Message: "too many requests"},
			want: "rate_limited: too many requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	sentinel := errors.New("original error")
	apiErr := &domain.APIError{
		Code:    domain.ErrCodeAuthFailed,
		Message: "auth failed",
		Wrapped: sentinel,
	}

	// errors.Is で wrapped error を透過的に検出できる
	if !errors.Is(apiErr, sentinel) {
		t.Error("errors.Is(apiErr, sentinel) = false, want true")
	}

	// Unwrap() が wrapped error を返す
	if got := apiErr.Unwrap(); got != sentinel {
		t.Errorf("Unwrap() = %v, want %v", got, sentinel)
	}
}

func TestAPIError_Unwrap_Nil(t *testing.T) {
	apiErr := &domain.APIError{
		Code:    domain.ErrCodeVideoNotFound,
		Message: "not found",
		Wrapped: nil,
	}

	if got := apiErr.Unwrap(); got != nil {
		t.Errorf("Unwrap() = %v, want nil", got)
	}
}

func TestAPIError_ErrorsAs(t *testing.T) {
	wrapped := errors.New("upstream error")
	apiErr := &domain.APIError{
		Code:    domain.ErrCodeQuotaExceeded,
		Message: "quota exceeded",
		Wrapped: wrapped,
	}

	var target *domain.APIError
	if !errors.As(apiErr, &target) {
		t.Fatal("errors.As(apiErr, &target) = false, want true")
	}
	if target.Code != domain.ErrCodeQuotaExceeded {
		t.Errorf("Code = %q, want %q", target.Code, domain.ErrCodeQuotaExceeded)
	}
}
