// Package memory メモリベースのリポジトリ実装を提供します
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// ChatRepository implements the ChatRepository interface using in-memory storage
type ChatRepository struct {
	mu       sync.RWMutex
	messages map[string][]entity.ChatMessage // videoID -> []ChatMessage
	maxSize  int                             // maximum messages to store per video
}

// NewChatRepository creates a new in-memory chat repository
func NewChatRepository(maxMessagesPerVideo int) *ChatRepository {
	return &ChatRepository{
		messages: make(map[string][]entity.ChatMessage),
		maxSize:  maxMessagesPerVideo,
	}
}

// SaveChatMessages persists chat messages to storage
func (r *ChatRepository) SaveChatMessages(ctx context.Context, messages []entity.ChatMessage) error {
	if len(messages) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Group messages by video ID
	messagesByVideo := make(map[string][]entity.ChatMessage)
	for _, message := range messages {
		messagesByVideo[message.VideoID] = append(messagesByVideo[message.VideoID], message)
	}

	// Store messages for each video
	for videoID, videoMessages := range messagesByVideo {
		existingMessages, exists := r.messages[videoID]
		if !exists {
			existingMessages = make([]entity.ChatMessage, 0)
		}

		// Append new messages
		allMessages := append(existingMessages, videoMessages...)

		// Sort by timestamp to maintain order
		sort.Slice(allMessages, func(i, j int) bool {
			return allMessages[i].Timestamp.Before(allMessages[j].Timestamp)
		})

		// Trim to max size if necessary (keep most recent messages)
		if len(allMessages) > r.maxSize {
			allMessages = allMessages[len(allMessages)-r.maxSize:]
		}

		r.messages[videoID] = allMessages
	}

	return nil
}

// GetChatHistory retrieves chat message history for a video
func (r *ChatRepository) GetChatHistory(ctx context.Context, videoID string, limit int) ([]entity.ChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages, exists := r.messages[videoID]
	if !exists {
		return []entity.ChatMessage{}, nil
	}

	// Apply limit if specified
	if limit > 0 && len(messages) > limit {
		// Return the most recent messages
		return messages[len(messages)-limit:], nil
	}

	// Return a copy to avoid external modification
	result := make([]entity.ChatMessage, len(messages))
	copy(result, messages)
	return result, nil
}

// GetRecentMessages retrieves recent messages within a time window
func (r *ChatRepository) GetRecentMessages(ctx context.Context, videoID string, since time.Time) ([]entity.ChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages, exists := r.messages[videoID]
	if !exists {
		return []entity.ChatMessage{}, nil
	}

	var recentMessages []entity.ChatMessage
	for _, message := range messages {
		if message.Timestamp.After(since) {
			recentMessages = append(recentMessages, message)
		}
	}

	return recentMessages, nil
}

// GetMessageCount returns the total number of messages for a video
func (r *ChatRepository) GetMessageCount(ctx context.Context, videoID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages, exists := r.messages[videoID]
	if !exists {
		return 0, nil
	}

	return len(messages), nil
}

// DeleteChatHistory removes all chat history for a video
func (r *ChatRepository) DeleteChatHistory(ctx context.Context, videoID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.messages, videoID)
	return nil
}

// GetAllVideoIDs returns all video IDs that have chat history
func (r *ChatRepository) GetAllVideoIDs(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	videoIDs := make([]string, 0, len(r.messages))
	for videoID := range r.messages {
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, nil
}

// GetChatStats returns statistics about stored chat messages
func (r *ChatRepository) GetChatStats(ctx context.Context) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]interface{}{
		"totalVideoIds": len(r.messages),
		"totalMessages": 0,
		"maxSize":       r.maxSize,
		"videoStats":    make(map[string]interface{}),
	}

	totalMessages := 0
	videoStats := make(map[string]interface{})

	for videoID, messages := range r.messages {
		messageCount := len(messages)
		totalMessages += messageCount

		var oldestTime, newestTime time.Time
		if messageCount > 0 {
			oldestTime = messages[0].Timestamp
			newestTime = messages[messageCount-1].Timestamp
		}

		videoStats[videoID] = map[string]interface{}{
			"messageCount": messageCount,
			"oldestTime":   oldestTime.Format("2006-01-02 15:04:05"),
			"newestTime":   newestTime.Format("2006-01-02 15:04:05"),
		}
	}

	stats["totalMessages"] = totalMessages
	stats["videoStats"] = videoStats

	return stats, nil
}
