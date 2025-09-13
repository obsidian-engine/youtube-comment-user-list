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

// SSEHandler Server-Sent Events用のハンドラー
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

// sendSSEMessage SSEメッセージを送信します
func (h *SSEHandler) sendSSEMessage(w http.ResponseWriter, eventType string, data interface{}, videoID string) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to marshal SSE data", videoID, "", err, map[string]interface{}{"event": eventType})
		return
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", eventType); err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to write SSE event line", videoID, "", err, map[string]interface{}{"event": eventType})
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to write SSE data line", videoID, "", err, map[string]interface{}{"event": eventType})
		return
	}
	if h.idleTouch != nil {
		h.idleTouch()
	}
}

// StreamUserList GET /api/sse/{videoId}/users を処理します
func (h *SSEHandler) StreamUserList(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r, "sse")
	videoID := chi.URLParam(r, "videoId")
	if videoID == "" {
		h.writeJSONError(w, http.StatusBadRequest, "video ID is required")
		return
	}

	h.logger.LogAPI(constants.LogLevelInfo, "SSE user list stream request", videoID, correlationID, map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	// SSEヘッダー
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Nginx等のバッファ無効化
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 監視セッション確認
	_, exists := h.chatMonitoringUC.GetMonitoringSession(videoID)
	if !exists {
		h.logger.LogAPI(constants.LogLevelInfo, "No active monitoring session found", videoID, correlationID, nil)
		h.sendSSEMessage(w, "error", map[string]string{"message": "No active monitoring session for this video"}, videoID)
		return
	}

	// 初期送信
	h.sendSSEMessage(w, "connected", map[string]interface{}{"message": "Connected to user list stream", "videoId": videoID}, videoID)
	h.sendCurrentUserList(w, videoID, correlationID)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	h.logger.LogAPI(constants.LogLevelInfo, "SSE user list stream established", videoID, correlationID, nil)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	updateTicker := time.NewTicker(constants.SSEUserListUpdateInterval)
	defer updateTicker.Stop()
	heartbeatTicker := time.NewTicker(constants.SSEUserListHeartbeatInterval)
	defer heartbeatTicker.Stop()
	connTimer := time.NewTimer(constants.SSEConnectionTimeout)
	defer func() {
		if !connTimer.Stop() {
			select {
			case <-connTimer.C:
			default:
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			h.logger.LogAPI(constants.LogLevelInfo, "SSE user list client disconnected", videoID, correlationID, nil)
			return
		case <-updateTicker.C:
			h.sendCurrentUserList(w, videoID, correlationID)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			if !connTimer.Stop() {
				select {
				case <-connTimer.C:
				default:
				}
			}
			connTimer.Reset(constants.SSEConnectionTimeout)
		case <-heartbeatTicker.C:
			h.sendSSEMessage(w, "heartbeat", map[string]interface{}{"timestamp": time.Now().Format(constants.TimeFormatISO8601), "type": "user_list"}, videoID)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			if !connTimer.Stop() {
				select {
				case <-connTimer.C:
				default:
				}
			}
			connTimer.Reset(constants.SSEConnectionTimeout)
		case <-connTimer.C:
			h.logger.LogAPI(constants.LogLevelInfo, "SSE user list connection timeout", videoID, correlationID, nil)
			h.sendSSEMessage(w, "timeout", map[string]string{"message": "Connection timeout"}, videoID)
			return
		}
	}
}

// sendCurrentUserList 現在のユーザーリスト送信
func (h *SSEHandler) sendCurrentUserList(w http.ResponseWriter, videoID, correlationID string) {
	users, err := h.chatMonitoringUC.GetUserList(context.Background(), videoID)
	if err != nil {
		h.logger.LogError(constants.LogLevelError, "Failed to get user list for SSE", videoID, correlationID, err, nil)
		h.sendSSEMessage(w, "error", map[string]string{"message": "Failed to retrieve user list"}, videoID)
		return
	}
	data := map[string]interface{}{
		"videoId":   videoID,
		"users":     users,
		"count":     len(users),
		"timestamp": time.Now().Format(constants.TimeFormatISO8601),
	}
	h.sendSSEMessage(w, "user_list", data, videoID)
}

// writeJSONError JSON形式のエラーレスポンス
func (h *SSEHandler) writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": message})
}
