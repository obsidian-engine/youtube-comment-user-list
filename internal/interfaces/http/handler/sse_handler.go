package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// SSEHandler Server-Sent Events for real-time updatesを処理します
type SSEHandler struct {
	chatMonitoringUC *usecase.ChatMonitoringUseCase
	logger           service.Logger
}

// NewSSEHandler 新しいSSEを作成します handler
func NewSSEHandler(
	chatMonitoringUC *usecase.ChatMonitoringUseCase,
	logger service.Logger,
) *SSEHandler {
	return &SSEHandler{
		chatMonitoringUC: chatMonitoringUC,
		logger:           logger,
	}
}

// SSEMessage a Server-Sent Event messageを表します
type SSEMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
	VideoID   string      `json:"video_id"`
}

// StreamMessages handles GET /api/sse/{videoId}
func (h *SSEHandler) StreamMessages(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("sse-%d", time.Now().Unix())

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract video ID from URL path
	videoID := h.extractVideoIDFromPath(r.URL.Path)
	if videoID == "" {
		http.Error(w, "video ID is required", http.StatusBadRequest)
		return
	}

	h.logger.LogAPI("INFO", "SSE stream request received", videoID, correlationID, map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Get the monitoring session
	session, exists := h.chatMonitoringUC.GetMonitoringSession(videoID)
	if !exists {
		h.logger.LogError("ERROR", "No active monitoring session found", videoID, correlationID, nil, nil)
		h.sendSSEMessage(w, "error", map[string]string{
			"message": "No active monitoring session for this video",
		}, videoID)
		return
	}

	// Send initial connection message
	h.sendSSEMessage(w, "connected", map[string]interface{}{
		"message":     "Connected to live chat stream",
		"videoId":     videoID,
		"subscribers": session.Subscribers,
	}, videoID)

	// Flush initial response
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	h.logger.LogAPI("INFO", "SSE stream established", videoID, correlationID, nil)

	// Create a context that will be cancelled when the client disconnects
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send periodic heartbeat and listen for messages
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Channel to receive messages from the monitoring session
	messagesChan := session.MessagesChan

	for {
		select {
		case <-ctx.Done():
			h.logger.LogAPI("INFO", "SSE client disconnected", videoID, correlationID, nil)
			return

		case <-heartbeatTicker.C:
			// Send heartbeat
			h.sendSSEMessage(w, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
			}, videoID)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case message, ok := <-messagesChan:
			if !ok {
				// Channel closed, monitoring stopped
				h.sendSSEMessage(w, "monitoring_stopped", map[string]string{
					"message": "Monitoring session ended",
				}, videoID)

				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}

				h.logger.LogAPI("INFO", "Monitoring session ended, closing SSE stream", videoID, correlationID, nil)
				return
			}

			// Send chat message to client
			h.sendChatMessage(w, message, correlationID)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-time.After(5 * time.Minute):
			// Timeout - close connection to prevent resource leaks
			h.logger.LogAPI("INFO", "SSE connection timeout", videoID, correlationID, nil)
			h.sendSSEMessage(w, "timeout", map[string]string{
				"message": "Connection timeout",
			}, videoID)
			return
		}
	}
}

// sendChatMessage sends a chat message via SSE
func (h *SSEHandler) sendChatMessage(w http.ResponseWriter, message entity.ChatMessage, correlationID string) {
	chatData := map[string]interface{}{
		"id":          message.ID,
		"displayName": message.AuthorDetails.DisplayName,
		"channelId":   message.AuthorDetails.ChannelID,
		"isChatOwner": message.AuthorDetails.IsChatOwner,
		"isModerator": message.AuthorDetails.IsModerator,
		"isMember":    message.AuthorDetails.IsMember,
		"timestamp":   message.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	h.sendSSEMessage(w, "chat_message", chatData, message.VideoID)

	h.logger.LogAPI("DEBUG", "Chat message sent via SSE", message.VideoID, correlationID, map[string]interface{}{
		"messageId":   message.ID,
		"displayName": message.AuthorDetails.DisplayName,
	})
}

// sendSSEMessage sends a formatted SSE message
func (h *SSEHandler) sendSSEMessage(w http.ResponseWriter, eventType string, data interface{}, videoID string) {
	message := SSEMessage{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		VideoID:   videoID,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		// Fallback to simple message if JSON marshaling fails
		if _, writeErr := fmt.Fprintf(w, `event: error
data: {"message": "JSON marshal error"}

`); writeErr != nil {
			// エラーをログに記録しますが処理を続行します
			h.logger.LogError("ERROR", "SSE write error during fallback", "", "", writeErr, nil)
		}
		return
	}

	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonData)); err != nil {
		h.logger.LogError("ERROR", "SSE write error", "", "", err, nil)
		return
	}

	// データをクライアントにフラッシュします
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// StreamUserList handles GET /api/sse/{videoId}/users
func (h *SSEHandler) StreamUserList(w http.ResponseWriter, r *http.Request) {
	correlationID := fmt.Sprintf("sse-users-%d", time.Now().Unix())

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract video ID from URL path
	videoID := h.extractVideoIDFromPath(r.URL.Path)
	if videoID == "" {
		http.Error(w, "video ID is required", http.StatusBadRequest)
		return
	}

	h.logger.LogAPI("INFO", "SSE user list stream request received", videoID, correlationID, nil)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send initial user list
	h.sendCurrentUserList(w, videoID, correlationID)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Update user list periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.LogAPI("INFO", "SSE user list client disconnected", videoID, correlationID, nil)
			return

		case <-ticker.C:
			h.sendCurrentUserList(w, videoID, correlationID)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-time.After(5 * time.Minute):
			// Timeout
			h.logger.LogAPI("INFO", "SSE user list connection timeout", videoID, correlationID, nil)
			return
		}
	}
}

// sendCurrentUserList sends the current user list via SSE
func (h *SSEHandler) sendCurrentUserList(w http.ResponseWriter, videoID, correlationID string) {
	users, err := h.chatMonitoringUC.GetUserList(context.Background(), videoID)
	if err != nil {
		h.logger.LogError("ERROR", "Failed to get user list for SSE", videoID, correlationID, err, nil)
		h.sendSSEMessage(w, "error", map[string]string{
			"message": "Failed to get user list",
		}, videoID)
		return
	}

	userListData := map[string]interface{}{
		"users": users,
		"count": len(users),
	}

	h.sendSSEMessage(w, "user_list", userListData, videoID)
}

// Helper method to extract video ID from path
func (h *SSEHandler) extractVideoIDFromPath(path string) string {
	// Expected paths: /api/sse/{videoId}, /api/sse/{videoId}/users
	parts := splitPath(path)
	if len(parts) >= 3 && parts[1] == "api" && parts[2] == "sse" {
		if len(parts) >= 4 {
			return parts[3]
		}
	}
	return ""
}
