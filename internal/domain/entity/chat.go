// Package entity ドメインエンティティを定義します
package entity

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
)

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

// ExtractVideoIDFromURL URLからYouTube動画IDを抽出します
func ExtractVideoIDFromURL(inputURL string) (string, error) {
	// 直接の動画ID入力を処理
	if !strings.Contains(inputURL, "youtube.com") && !strings.Contains(inputURL, "youtu.be") {
		// 既に動画IDである可能性があるため、形式を検証
		if err := ValidateVideoID(inputURL); err == nil {
			return inputURL, nil
		}
	}

	// URLをパース
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	var videoID string

	// 異なるYouTube URL形式を処理
	switch {
	case strings.Contains(parsedURL.Host, "youtube.com"):
		// 標準YouTube URL: https://www.youtube.com/watch?v=VIDEO_ID
		if parsedURL.Path == "/watch" {
			videoID = parsedURL.Query().Get("v")
			break
		}

		if strings.HasPrefix(parsedURL.Path, "/embed/") {
			// 埋め込みURL: https://www.youtube.com/embed/VIDEO_ID
			videoID = strings.TrimPrefix(parsedURL.Path, "/embed/")
			break
		}
	case strings.Contains(parsedURL.Host, "youtu.be"):
		// 短縮URL: https://youtu.be/VIDEO_ID
		videoID = strings.TrimPrefix(parsedURL.Path, "/")
	}

	if videoID == "" {
		return "", fmt.Errorf("could not extract video ID from URL: %s", inputURL)
	}

	// 動画IDから追加のパラメータを削除
	if idx := strings.Index(videoID, "?"); idx != -1 {
		videoID = videoID[:idx]
	}
	if idx := strings.Index(videoID, "&"); idx != -1 {
		videoID = videoID[:idx]
	}

	// 抽出された動画IDを検証
	if err := ValidateVideoID(videoID); err != nil {
		return "", fmt.Errorf("extracted invalid video ID '%s': %w", videoID, err)
	}

	return videoID, nil
}

// ValidateVideoID YouTube動画IDの形式を検証します
func ValidateVideoID(videoID string) error {
	if videoID == "" {
		return fmt.Errorf("video ID cannot be empty")
	}

	// YouTube動画IDは通常11文字で、英数字、ハイフン、アンダースコアを含みます
	if len(videoID) != constants.YouTubeVideoIDLength {
		return fmt.Errorf("invalid video ID length: expected %d characters, got %d", constants.YouTubeVideoIDLength, len(videoID))
	}

	// 有効な文字をチェック（英数字、ハイフン、アンダースコア）
	for _, char := range videoID {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return fmt.Errorf("invalid character in video ID: %c", char)
		}
	}

	return nil
}
