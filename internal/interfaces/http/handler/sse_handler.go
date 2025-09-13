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
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// SSEHandler リアルタイム更新用のServer-Sent Eventsを処理します
type SSEHandler struct {
    chatMonitoringUC *usecase.ChatMonitoringUseCase
    logger           repository.Logger
    idleTouch        func()
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

// SetIdleTouch SSE送出時に呼び出すアイドル更新関数を設定
func (h *SSEHandler) SetIdleTouch(fn func()) { h.idleTouch = fn }

// チャット本文ストリーム機能は未使用のため削除しました

// sendSSEMessage SSEメッセージ送信
func (h *SSEHandler) sendSSEMessage(w http.ResponseWriter, eventType string, data interface{}, videoID string) {
	jsonData, err := json.Marshal(data)
    if err != nil {
        h.logger.LogError(constants.LogLevelError, "Failed to marshal SSE message", videoID, "", err, nil)
        return
    }

	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))
    if _, err := w.Write([]byte(message)); err != nil {
        h.logger.LogError(constants.LogLevelError, "Failed to write SSE message", videoID, "", err, nil)
    }
    if h.idleTouch != nil { h.idleTouch() }
}

// StreamUserList GET /api/sse/{videoId}/users を処理します
func (h *SSEHandler) StreamUserList(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("sse-users-%d", time.Now().Unix())
	videoID := chi.URLParam(r, "videoId")

	if videoID == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video ID is required")
		return
	}

    h.logger.LogAPI(constants.LogLevelInfo, "SSE user list stream request received", videoID, correlationID, map[string]interface{}{
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
        // セッション未存在は異常ではなく状態情報のため INFO で記録
        h.logger.LogAPI(constants.LogLevelInfo, "No active monitoring session found", videoID, correlationID, nil)
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

    h.logger.LogAPI(constants.LogLevelInfo, "SSE user list stream established", videoID, correlationID, nil)

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
            h.logger.LogAPI(constants.LogLevelInfo, "SSE user list client disconnected", videoID, correlationID, nil)
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
            h.logger.LogAPI(constants.LogLevelInfo, "SSE user list connection timeout", videoID, correlationID, nil)
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
        h.logger.LogError(constants.LogLevelError, "Failed to get user list for SSE", videoID, correlationID, err, nil)
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
        h.logger.LogError(constants.LogLevelError, "Failed to encode JSON error response", "", "", err, nil)
    }
}
