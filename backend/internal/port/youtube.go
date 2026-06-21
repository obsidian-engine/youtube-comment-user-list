package port

import (
	"context"
	"time"
)

// ChatMessage は YouTube Live Chat のメッセージの最小情報です。
type ChatMessage struct {
	ID          string // メッセージID（重複チェック用）
	ChannelID   string
	DisplayName string
	Message     string // コメント本文
	PublishedAt time.Time
}

// VideoMeta は videos.list から取得した動画メタデータです。
type VideoMeta struct {
	LiveChatID   string
	Title        string
	ChannelTitle string
}

// VideoLiveDetails は予約監視に必要な liveStreamingDetails の情報です。
type VideoLiveDetails struct {
	LiveChatID         string // activeLiveChatId (チャット未開なら空)
	Title              string
	ChannelTitle       string
	ScheduledStartTime time.Time // liveStreamingDetails.scheduledStartTime (未指定なら zero)
	IsLiveContent      bool      // liveStreamingDetails != nil なら true
}

// YouTubePort は YouTube API 呼び出しを抽象化します。
type YouTubePort interface {
	// 指定 videoID の activeLiveChatId と動画メタデータを取得します。
	GetActiveLiveChatID(ctx context.Context, videoID string) (VideoMeta, error)
	// liveChatId のメッセージを取得します。ページング対応のため、pageToken を受け取り、nextPageToken を返します。
	// 配信終了検知は isEnded で返します。
	ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) (items []ChatMessage, nextPageToken string, pollingIntervalMillis int64, skippedCount int, isEnded bool, err error)
	// チャンネルIDからチャンネル名（snippet.title）を一括取得します。
	GetChannelDisplayNames(ctx context.Context, channelIDs []string) (map[string]string, error)
	// チャンネルIDからハンドル（@username, snippet.customUrl）を一括取得します。
	// ハンドルが存在しないチャンネルはマップに含まれません。
	GetChannelHandles(ctx context.Context, channelIDs []string) (map[string]string, error)
	// 指定 videoID の liveStreamingDetails を取得します。activeLiveChatId 空でもエラーにせず返します (予約用)。
	GetVideoLiveDetails(ctx context.Context, videoID string) (VideoLiveDetails, error)
}
