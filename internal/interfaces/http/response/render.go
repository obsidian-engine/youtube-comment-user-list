// Package response HTTP レスポンス用の共通構造体とヘルパー関数を提供します
package response

import (
	"net/http"

	"github.com/go-chi/render"
)

// APIResponse 標準的なAPIレスポンス構造体
type APIResponse struct {
	Success       bool        `json:"success"`
	Data          interface{} `json:"data,omitempty"`
	Error         string      `json:"error,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// Render render.Renderer インターフェースを実装
func (a *APIResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Content-Type は render パッケージが自動設定
	return nil
}

// RenderSuccess 成功レスポンスを返す
func RenderSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	render.JSON(w, r, &APIResponse{
		Success: true,
		Data:    data,
	})
}

// RenderSuccessWithCorrelation 成功レスポンス（相関ID付き）を返す
func RenderSuccessWithCorrelation(w http.ResponseWriter, r *http.Request, data interface{}, correlationID string) {
	render.JSON(w, r, &APIResponse{
		Success:       true,
		Data:          data,
		CorrelationID: correlationID,
	})
}

// RenderErrorWithCorrelation エラーレスポンス（相関ID付き）を返す
func RenderErrorWithCorrelation(w http.ResponseWriter, r *http.Request, status int, message, correlationID string) {
	render.Status(r, status)
	render.JSON(w, r, &APIResponse{
		Success:       false,
		Error:         message,
		CorrelationID: correlationID,
	})
}

// UserListResponse ユーザーリスト専用レスポンス
type UserListResponse struct {
	Success bool        `json:"success"`
	Users   interface{} `json:"users"`
	Count   int         `json:"count"`
	Error   string      `json:"error,omitempty"`
}

// Render render.Renderer インターフェースを実装
func (u *UserListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// RenderUserList ユーザーリストレスポンスを返す
func RenderUserList(w http.ResponseWriter, r *http.Request, users interface{}, count int) {
	render.JSON(w, r, &UserListResponse{
		Success: true,
		Users:   users,
		Count:   count,
	})
}

// StartMonitoringResponse 監視開始専用レスポンス
type StartMonitoringResponse struct {
	Success bool   `json:"success"`
	VideoID string `json:"video_id"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Render render.Renderer インターフェースを実装
func (s *StartMonitoringResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// RenderStartMonitoring 監視開始レスポンスを返す
func RenderStartMonitoring(w http.ResponseWriter, r *http.Request, videoID, message string) {
	render.JSON(w, r, &StartMonitoringResponse{
		Success: true,
		VideoID: videoID,
		Message: message,
	})
}
