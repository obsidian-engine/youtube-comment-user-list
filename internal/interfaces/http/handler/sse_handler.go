package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// SSEHandler リアルタイム更新用のServer-Sent Eventsを処理します
type SSEHandler struct {
	chatMonitoringUC *usecase.ChatMonitoringUseCase
	logger           repository.Logger
}

// NewSSEHandler 新しいSSEハンドラーを作成します
func NewSSEHandler(
	chatMonitoringUC *usecase.ChatMonitoringUseCase,
	logger repository.Logger,
) *SSEHandler {
	return &SSEHandler{
		chatMonitoringUC: chatMonitoringUC,
		logger:           logger,
	}
}

// SSEMessage Server-Sent Eventメッセージを表します
type SSEMessage struct {
	Type        string `json:"type"`
	VideoID     string `json:"video_id"`
	DisplayName string `json:"display_name"`
	ChannelID   string `json:"channel_id"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
}

// StreamMessages GET /api/sse/{videoId} を処理します
func (h *SSEHandler) StreamMessages(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("sse-%d", time.Now().Unix())
	videoID := chi.URLParam(r, "videoId")

	if videoID == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video ID is required")
		return
	}

	h.logger.LogAPI("INFO", "SSE stream request received", videoID, correlationID, map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	// SSEヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 監視セッションを取得
	_, exists := h.chatMonitoringUC.GetMonitoringSession(videoID)
	if !exists {
		h.logger.LogError("ERROR", "No active monitoring session found", videoID, correlationID, nil, nil)
		h.sendSSEMessage(w, "error", map[string]string{
			"message": "No active monitoring session for this video",
		}, videoID)
		return
	}

	// 初期接続メッセージを送信
	h.sendSSEMessage(w, "connected", map[string]interface{}{
		"message":     "Connected to live chat stream",
		"videoId":     videoID,
		"subscribers": 1,
	}, videoID)

	// 初期レスポンスをフラッシュ
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	h.logger.LogAPI("INFO", "SSE stream established", videoID, correlationID, nil)

	// クライアント切断時にキャンセルされるコンテキストを作成
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// 定期的なハートビートを送信し、メッセージを監視
	heartbeatTicker := time.NewTicker(constants.SSEHeartbeatInterval)
	defer heartbeatTicker.Stop()

	// 監視セッションからメッセージを受信するチャネル
	session, _ := h.chatMonitoringUC.GetMonitoringSession(videoID)
	messagesChan := session.MessagesChan

	for {
		select {
		case <-ctx.Done():
			h.logger.LogAPI("INFO", "SSE client disconnected", videoID, correlationID, nil)
			return

		case <-heartbeatTicker.C:
			// ハートビートを送信
			h.sendSSEMessage(w, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Format(constants.TimeFormatISO8601),
			}, videoID)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case message, ok := <-messagesChan:
			if !ok {
				// チャネルが閉じられ、監視が停止されました
				h.sendSSEMessage(w, "monitoring_stopped", map[string]string{
					"message": "Monitoring session ended",
				}, videoID)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				h.logger.LogAPI("INFO", "Monitoring session ended, closing SSE stream", videoID, correlationID, nil)
				return
			}

			// チャットメッセージをクライアントに送信
			h.sendChatMessage(w, &message, correlationID)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-time.After(constants.SSEConnectionTimeout):
			// タイムアウト - リソースリークを防ぐため接続を閉じる
			h.logger.LogAPI("INFO", "SSE connection timeout", videoID, correlationID, nil)
			h.sendSSEMessage(w, "timeout", map[string]string{
				"message": "Connection timeout",
			}, videoID)
			return
		}
	}
}

// 古いsendChatMessage関数は削除 - sendChatMessageToGinを使用

// sendChatMessage チャットメッセージ送信
func (h *SSEHandler) sendChatMessage(w http.ResponseWriter, message *entity.ChatMessage, _ string) {
	messageData := SSEMessage{
		Type:        "chat_message",
		VideoID:     message.VideoID,
		DisplayName: message.AuthorDetails.DisplayName,
		ChannelID:   message.AuthorDetails.ChannelID,
		Message:     message.ID,
		Timestamp:   message.Timestamp.Format(constants.TimeFormatISO8601),
	}

	h.sendSSEMessage(w, "message", messageData, message.VideoID)
}

// sendSSEMessage SSEメッセージ送信
func (h *SSEHandler) sendSSEMessage(w http.ResponseWriter, eventType string, data interface{}, videoID string) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to marshal SSE message", videoID, "", err, nil)
		return
	}

	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))
	if _, err := w.Write([]byte(message)); err != nil {
		h.logger.LogError("ERROR", "Failed to write SSE message", videoID, "", err, nil)
	}
}

// StreamUserList GET /api/sse/{videoId}/users を処理します
func (h *SSEHandler) StreamUserList(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("sse-users-%d", time.Now().Unix())
	videoID := chi.URLParam(r, "videoId")

	if videoID == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video ID is required")
		return
	}

	h.logger.LogAPI("INFO", "SSE user list stream request received", videoID, correlationID, map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	// SSEヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 監視セッションを取得
	_, exists := h.chatMonitoringUC.GetMonitoringSession(videoID)
	if !exists {
		h.logger.LogError("ERROR", "No active monitoring session found", videoID, correlationID, nil, nil)
		h.sendSSEMessage(w, "error", map[string]string{
			"message": "No active monitoring session for this video",
		}, videoID)
		return
	}

	// 初期接続メッセージを送信
	h.sendSSEMessage(w, "connected", map[string]interface{}{
		"message": "Connected to user list stream",
		"videoId": videoID,
	}, videoID)

	// 現在のユーザーリストを送信
	h.sendCurrentUserList(w, videoID, correlationID)

	// 初期レスポンスをフラッシュ
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	h.logger.LogAPI("INFO", "SSE user list stream established", videoID, correlationID, nil)

	// クライアント切断時にキャンセルされるコンテキストを作成
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// 定期的なユーザーリスト更新を送信
	updateTicker := time.NewTicker(constants.SSEUserListUpdateInterval)
	defer updateTicker.Stop()
	// ユーザーリスト用ハートビート（更新間隔を長くしたため）
	heartbeatTicker := time.NewTicker(constants.SSEUserListHeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.LogAPI("INFO", "SSE user list client disconnected", videoID, correlationID, nil)
			return

		case <-updateTicker.C:
			// 更新されたユーザーリストを送信
			h.sendCurrentUserList(w, videoID, correlationID)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-heartbeatTicker.C:
			// ハートビート（軽量イベント）
			h.sendSSEMessage(w, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Format(constants.TimeFormatISO8601),
				"type":      "user_list",
			}, videoID)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-time.After(constants.SSEConnectionTimeout):
			// タイムアウト - リソースリークを防ぐため接続を閉じる
			h.logger.LogAPI("INFO", "SSE user list connection timeout", videoID, correlationID, nil)
			h.sendSSEMessage(w, "timeout", map[string]string{
				"message": "Connection timeout",
			}, videoID)
			return
		}
	}
}

// 古いsendCurrentUserList関数は削除 - sendCurrentUserListToGinを使用

// sendCurrentUserList 現在のユーザーリスト送信
func (h *SSEHandler) sendCurrentUserList(w http.ResponseWriter, videoID, correlationID string) {
	users, err := h.chatMonitoringUC.GetUserList(context.Background(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get user list for SSE", videoID, correlationID, err, nil)
		h.sendSSEMessage(w, "error", map[string]string{
			"message": "Failed to retrieve user list",
		}, videoID)
		return
	}

	userListData := map[string]interface{}{
		"videoId":   videoID,
		"users":     users,
		"count":     len(users),
		"timestamp": time.Now().Format(constants.TimeFormatISO8601),
	}

	h.sendSSEMessage(w, "user_list", userListData, videoID)
}

// writeJSONError JSON形式のエラーレスポンスを書き込みます
func (h *SSEHandler) writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	}); err != nil {
		h.logger.LogError("ERROR", "Failed to encode JSON error response", "", "", err, nil)
	}
}
