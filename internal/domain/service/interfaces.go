// Package service ドメインサービスとインターフェースを定義します
package service

import (
	"context"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// YouTubeClient YouTube API操作のインターフェースを表します
type YouTubeClient interface {
	// FetchVideoInfo ライブ配信の詳細を含む動画情報を取得します
	FetchVideoInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error)

	// FetchLiveChat ライブチャットメッセージを取得します
	FetchLiveChat(ctx context.Context, liveChatID string, pageToken string) (*entity.PollResult, error)
}

// ChatRepository チャットデータの永続化のインターフェースを表します
type ChatRepository interface {
	// SaveChatMessages チャットメッセージをストレージに永続化します
	SaveChatMessages(ctx context.Context, messages []entity.ChatMessage) error
}

// UserRepository ユーザーデータ操作のインターフェースを表します
type UserRepository interface {
	// GetUserList 動画のユーザーリストを取得します
	GetUserList(ctx context.Context, videoID string) (*entity.UserList, error)

	// UpdateUserList 動画のユーザーリストを更新します
	UpdateUserList(ctx context.Context, videoID string, userList *entity.UserList) error
}

// Logger ログ出力操作のインターフェースを表します
type Logger interface {
	// LogStructured 構造化メッセージをログ出力します
	LogStructured(level, component, event, message, videoID, correlationID string, context map[string]interface{})

	// LogAPI API関連のイベントをログ出力します
	LogAPI(level, message, videoID, correlationID string, context map[string]interface{})

	// LogPoller ポーリング関連のイベントをログ出力します
	LogPoller(level, message, videoID, correlationID string, context map[string]interface{})

	// LogUser ユーザー関連のイベントをログ出力します
	LogUser(level, message, videoID, correlationID string, context map[string]interface{})

	// LogError エラーイベントをログ出力します
	LogError(level, message, videoID, correlationID string, err error, context map[string]interface{})
}

// EventPublisher イベント発行のインターフェースを表します
type EventPublisher interface {
	// PublishChatMessage 新しいチャットメッセージイベントを発行します
	PublishChatMessage(ctx context.Context, message entity.ChatMessage) error

	// PublishUserAdded ユーザー追加イベントを発行します
	PublishUserAdded(ctx context.Context, user entity.User, videoID string) error
}
