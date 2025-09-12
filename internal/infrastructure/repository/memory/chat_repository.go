// Package memory メモリベースのリポジトリ実装を提供します
package memory

import (
	"context"
	"sort"
	"sync"

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
