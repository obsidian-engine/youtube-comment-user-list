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

// renderError は統一されたエラーレスポンスを返す
func renderError(w stdhttp.ResponseWriter, r *stdhttp.Request, code int, err string, message string) {
	render.Status(r, code)
	render.JSON(w, r, ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
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