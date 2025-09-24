package http

import (
	"net/http"

	"github.com/go-chi/render"
)

// RenderJSONResponse は統一されたJSON応答を送信します
func RenderJSONResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	render.JSON(w, r, data)
}

// RenderErrorResponse はエラー応答を送信します
func RenderErrorResponse(w http.ResponseWriter, r *http.Request, errorType, message string, status int) {
	errorResponse := ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    status,
	}
	RenderJSONResponse(w, r, status, errorResponse)
}

// RenderSuccessResponse は成功応答を送信します
func RenderSuccessResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	RenderJSONResponse(w, r, status, data)
}

// RenderBadRequestError はBad Requestエラーを送信します
func RenderBadRequestError(w http.ResponseWriter, r *http.Request, message string) {
	RenderErrorResponse(w, r, "bad_request", message, StatusBadRequest)
}

// RenderInternalServerError はInternal Server Errorを送信します  
func RenderInternalServerError(w http.ResponseWriter, r *http.Request, message string) {
	RenderErrorResponse(w, r, "internal_error", message, StatusInternalServerError)
}

// RenderBadGatewayError はBad Gateway Errorを送信します
func RenderBadGatewayError(w http.ResponseWriter, r *http.Request, message string) {
	RenderErrorResponse(w, r, "bad_gateway", message, StatusBadGateway)
}

// RenderNotFoundError はNot Found Errorを送信します
func RenderNotFoundError(w http.ResponseWriter, r *http.Request, message string) {
	RenderErrorResponse(w, r, "not_found", message, http.StatusNotFound)
}