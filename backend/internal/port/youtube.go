package port

import (
	"context"
	"time"
)

// ChatMessage は YouTube Live Chat のメッセージの最小情報です。
type ChatMessage struct {
	ID          string    // メッセージID（重複チェック用）
	ChannelID   string
	DisplayName string
	PublishedAt time.Time
}

// YouTubePort は YouTube API 呼び出しを抽象化します。
type YouTubePort interface {
    // 指定 videoID の activeLiveChatId を取得します。
    GetActiveLiveChatID(ctx context.Context, videoID string) (string, error)
    // liveChatId のメッセージを取得します。ページング対応のため、pageToken を受け取り、nextPageToken を返します。
    // 配信終了検知は isEnded で返します。
    ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) (items []ChatMessage, nextPageToken string, isEnded bool, err error)
}
