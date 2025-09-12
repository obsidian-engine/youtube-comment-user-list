// Package entity ドメインエンティティを定義します
package entity

import "time"

// ChatMessage YouTubeからのライブチャットメッセージを表します
type ChatMessage struct {
	ID            string
	AuthorDetails AuthorDetails
	Timestamp     time.Time
	VideoID       string
}

// AuthorDetails チャットメッセージの作成者に関する情報を含みます
type AuthorDetails struct {
	DisplayName string
	ChannelID   string
	IsChatOwner bool
	IsModerator bool
	IsMember    bool
}

// PollResult ポーリング操作の結果を表します
type PollResult struct {
	Messages          []ChatMessage
	NextPageToken     string
	PollingIntervalMs int
	Success           bool
	Error             error
}

// LiveStreamingDetails YouTubeライブ配信の情報を含みます
type LiveStreamingDetails struct {
	ActiveLiveChatID   string
	ActualStartTime    string
	ActualEndTime      string
	ConcurrentViewers  string
	ScheduledStartTime string
}

// VideoInfo YouTube動画の情報を表します
type VideoInfo struct {
	ID                   string
	Title                string
	ChannelTitle         string
	LiveBroadcastContent string
	LiveStreamingDetails LiveStreamingDetails
}
