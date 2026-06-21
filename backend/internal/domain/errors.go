package domain

import (
	"errors"
	"fmt"
)

// ErrNotFound はリソースが存在しない場合に返す sentinel error。
var ErrNotFound = errors.New("not found")

// APIErrorCode は YouTube API から返るエラー種別を表す。
type APIErrorCode string

const (
	ErrCodeQuotaExceeded   APIErrorCode = "quota_exceeded"
	ErrCodeVideoNotFound   APIErrorCode = "video_not_found"
	ErrCodeLiveChatEnded   APIErrorCode = "live_chat_ended"
	ErrCodeAuthFailed      APIErrorCode = "auth_failed"
	ErrCodeRateLimited     APIErrorCode = "rate_limited"
	ErrCodeConflict        APIErrorCode = "conflict"         // 現 state と矛盾する操作 (例: ACTIVE 中の Reserve)
	ErrCodeInvalidArgument APIErrorCode = "invalid_argument" // 入力不正 (例: 非 live video の Reserve)
)

// APIError は YouTube API エラーを機械可読コードとともに保持する。
// handler 側で errors.As(err, &apiErr) で取り出して ErrorResponse.Code に入れる。
type APIError struct {
	Code    APIErrorCode
	Message string
	Wrapped error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Wrapped
}
