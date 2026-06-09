package http

import (
	stdhttp "net/http"

	"github.com/go-chi/render"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
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

// renderBadGateway はバッドゲートウェイ用のヘルパー
func renderBadGateway(w stdhttp.ResponseWriter, r *stdhttp.Request, message string) {
	renderErrorWithCollector(w, r, StatusBadGateway, "bad_gateway", message, nil)
}

// renderBadGatewayWithCollector はバッドゲートウェイ用のヘルパー (collector 付き)
func renderBadGatewayWithCollector(w stdhttp.ResponseWriter, r *stdhttp.Request, message string, collector *logging.Collector) {
	renderErrorWithCollector(w, r, StatusBadGateway, "bad_gateway", message, collector)
}
