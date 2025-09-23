package http

import (
	stdhttp "net/http"

	"github.com/go-chi/render"
)

// ErrorResponse はAPIエラーレスポンスの標準形式
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// ErrorConfig はrenderError関数の引数をまとめる設定構造体
type ErrorConfig struct {
	ResponseWriter stdhttp.ResponseWriter
	Request        *stdhttp.Request
	Code           int
	Error          string
	Message        string
}

// renderErrorWithConfig は統一されたエラーレスポンスを返す（新しいAPI）
func renderErrorWithConfig(config ErrorConfig) {
	render.Status(config.Request, config.Code)
	render.JSON(config.ResponseWriter, config.Request, ErrorResponse{
		Error:   config.Error,
		Message: config.Message,
		Code:    config.Code,
	})
}

// renderError は統一されたエラーレスポンスを返す（後方互換性のために残存）
func renderError(w stdhttp.ResponseWriter, r *stdhttp.Request, code int, err string, message string) {
	renderErrorWithConfig(ErrorConfig{
		ResponseWriter: w,
		Request:        r,
		Code:           code,
		Error:          err,
		Message:        message,
	})
}

// renderInternalError は内部エラー用のヘルパー
func renderInternalError(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderError(w, r, StatusInternalServerError, "internal_error", message)
}

// renderBadRequest はバッドリクエスト用のヘルパー
func renderBadRequest(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderError(w, r, StatusBadRequest, "bad_request", message)
}

// renderBadGateway はバッドゲートウェイ用のヘルパー
func renderBadGateway(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderError(w, r, StatusBadGateway, "bad_gateway", message)
}
