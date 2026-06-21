package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/go-chi/render"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// ErrorResponse はAPIエラーレスポンスの標準形式
type ErrorResponse struct {
	Error    string      `json:"error"`
	Message  string      `json:"message,omitempty"`
	Code     string      `json:"code,omitempty"` // 機械可読エラーコード (quota_exceeded 等)
	HTTPCode int         `json:"httpCode"`       // HTTP ステータスコード (旧 Code から rename)
	Logs     []LogDetail `json:"logs,omitempty"` // Collector entries
}

// ErrorConfig はrenderError関数の引数をまとめる設定構造体
type ErrorConfig struct {
	ResponseWriter stdhttp.ResponseWriter
	Request        *stdhttp.Request
	HTTPCode       int
	Error          string
	Message        string
	Code           string             // 機械可読コード (省略可)
	Collector      *logging.Collector // nil 許容
}

// renderErrorWithConfig は統一されたエラーレスポンスを返す（新しいAPI）
func renderErrorWithConfig(config ErrorConfig) {
	resp := ErrorResponse{
		Error:    config.Error,
		Message:  config.Message,
		Code:     config.Code,
		HTTPCode: config.HTTPCode,
	}
	if config.Collector != nil {
		resp.Logs = collectLogs(config.Collector)
	}
	render.Status(config.Request, config.HTTPCode)
	render.JSON(config.ResponseWriter, config.Request, resp)
}

// renderError は統一されたエラーレスポンスを返す（後方互換性のために残存）
func renderError(w stdhttp.ResponseWriter, r *stdhttp.Request, code int, err string, message string) {
	renderErrorWithCollector(w, r, code, err, message, nil)
}

// renderErrorWithCollector はcollector付きエラーレスポンスを返す
func renderErrorWithCollector(w stdhttp.ResponseWriter, r *stdhttp.Request, code int, err string, message string, collector *logging.Collector) {
	renderErrorWithConfig(ErrorConfig{
		ResponseWriter: w,
		Request:        r,
		HTTPCode:       code,
		Error:          err,
		Message:        message,
		Collector:      collector,
	})
}

// renderInternalError は内部エラー用のヘルパー
func renderInternalError(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderErrorWithCollector(w, r, StatusInternalServerError, "internal_error", message, nil)
}

// renderInternalErrorWithCollector は内部エラー用のヘルパー (collector 付き)
func renderInternalErrorWithCollector(w stdhttp.ResponseWriter, r *stdhttp.Request, message string, collector *logging.Collector) {
	renderErrorWithCollector(w, r, StatusInternalServerError, "internal_error", message, collector)
}

// renderBadRequest はバッドリクエスト用のヘルパー
func renderBadRequest(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderErrorWithCollector(w, r, StatusBadRequest, "bad_request", message, nil)
}

// renderBadRequestWithCollector はバッドリクエスト用のヘルパー (collector 付き)
func renderBadRequestWithCollector(w stdhttp.ResponseWriter, r *stdhttp.Request, message string, collector *logging.Collector) {
	renderErrorWithCollector(w, r, StatusBadRequest, "bad_request", message, collector)
}

// renderBadGateway はバッドゲートウェイ用のヘルパー
func renderBadGateway(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderErrorWithCollector(w, r, StatusBadGateway, "bad_gateway", message, nil)
}

// renderUsecaseError は usecase 層 error を ErrorResponse に変換する。
// domain.APIError を含む場合は機械可読 Code + 対応 HTTP status に振り分ける。
// それ以外は fallbackStatus + fallbackErr ("bad_gateway" 等) を使う。
func renderUsecaseError(w stdhttp.ResponseWriter, r *stdhttp.Request, err error, message string, collector *logging.Collector, fallbackStatus int, fallbackErr string) {
	var apiErr *domain.APIError
	if errors.As(err, &apiErr) {
		status := httpStatusFor(apiErr.Code)
		renderErrorWithConfig(ErrorConfig{
			ResponseWriter: w,
			Request:        r,
			HTTPCode:       status,
			Error:          fallbackErr,
			Message:        message,
			Code:           string(apiErr.Code),
			Collector:      collector,
		})
		return
	}
	renderErrorWithCollector(w, r, fallbackStatus, fallbackErr, message, collector)
}

func httpStatusFor(code domain.APIErrorCode) int {
	switch code {
	case domain.ErrCodeQuotaExceeded, domain.ErrCodeRateLimited:
		return stdhttp.StatusTooManyRequests
	case domain.ErrCodeVideoNotFound:
		return stdhttp.StatusNotFound
	case domain.ErrCodeLiveChatEnded:
		return stdhttp.StatusGone
	case domain.ErrCodeAuthFailed:
		return stdhttp.StatusUnauthorized
	case domain.ErrCodeConflict:
		return StatusConflict
	case domain.ErrCodeInvalidArgument:
		return StatusBadRequest
	default:
		return StatusBadGateway
	}
}
