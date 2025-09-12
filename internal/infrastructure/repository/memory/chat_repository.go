// Package memory メモリベースのリポジトリ実装を提供します
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// ChatRepository インメモリストレージを使用してChatRepositoryインターフェースを実装します
type ChatRepository struct {
	mu       sync.RWMutex
	messages map[string][]entity.ChatMessage // videoID -> []ChatMessage
	maxSize  int                             // 動画ごとに保存する最大メッセージ数
}

// NewChatRepository 新しいインメモリチャットリポジトリを作成します
func NewChatRepository(maxMessagesPerVideo int) *ChatRepository {
	return &ChatRepository{
		messages: make(map[string][]entity.ChatMessage),
		maxSize:  maxMessagesPerVideo,
	}
}

// SaveChatMessages チャットメッセージをストレージに永続化します
func (r *ChatRepository) SaveChatMessages(ctx context.Context, messages []entity.ChatMessage) error {
	if len(messages) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 動画IDでメッセージをグループ化
	messagesByVideo := make(map[string][]entity.ChatMessage)
	for _, message := range messages {
		messagesByVideo[message.VideoID] = append(messagesByVideo[message.VideoID], message)
	}

	// 各動画のメッセージを保存
	for videoID, videoMessages := range messagesByVideo {
		existingMessages, exists := r.messages[videoID]
		if !exists {
			existingMessages = make([]entity.ChatMessage, 0)
		}

		// 新しいメッセージを追加
		allMessages := append(existingMessages, videoMessages...)

		// タイムスタンプでソートして順序を保持
		sort.Slice(allMessages, func(i, j int) bool {
			return allMessages[i].Timestamp.Before(allMessages[j].Timestamp)
		})

		// 必要に応じて最大サイズまでトリミング（最新のメッセージを保持）
		if len(allMessages) > r.maxSize {
			allMessages = allMessages[len(allMessages)-r.maxSize:]
		}

		r.messages[videoID] = allMessages
	}

	return nil
}
