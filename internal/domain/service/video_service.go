package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// VideoService 動画関連のビジネスロジックを処理します
type VideoService struct {
	youtubeClient YouTubeClient
	logger        Logger
}

// NewVideoService 新しいVideoServiceを作成します
func NewVideoService(
	youtubeClient YouTubeClient,
	logger Logger,
) *VideoService {
	return &VideoService{
		youtubeClient: youtubeClient,
		logger:        logger,
	}
}

// GetVideoInfo 包括的な動画情報を取得します
func (vs *VideoService) GetVideoInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	correlationID := fmt.Sprintf("video-%s", videoID)

	// 動画IDの形式を検証
	if err := vs.validateVideoID(videoID); err != nil {
		vs.logger.LogError("ERROR", "Invalid video ID", videoID, correlationID, err, map[string]interface{}{
			"videoId": videoID,
		})
		return nil, err
	}

	vs.logger.LogAPI("INFO", "Fetching video info", videoID, correlationID, map[string]interface{}{
		"event": "video_info_request",
	})

	// YouTube APIから動画情報を取得
	videoInfo, err := vs.youtubeClient.FetchVideoInfo(ctx, videoID)
	if err != nil {
		vs.logger.LogError("ERROR", "Failed to fetch video info", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

	vs.logger.LogAPI("INFO", "Video info retrieved successfully", videoID, correlationID, map[string]interface{}{
		"title":                videoInfo.Title,
		"channelTitle":         videoInfo.ChannelTitle,
		"liveBroadcastContent": videoInfo.LiveBroadcastContent,
		"activeLiveChatId":     videoInfo.LiveStreamingDetails.ActiveLiveChatID,
	})

	return videoInfo, nil
}

// ExtractVideoIDFromURL video ID from various YouTube URL formatsを抽出します
func (vs *VideoService) ExtractVideoIDFromURL(inputURL string) (string, error) {
	// 直接の動画ID入力を処理
	if !strings.Contains(inputURL, "youtube.com") && !strings.Contains(inputURL, "youtu.be") {
		// 既に動画IDである可能性があるため、形式を検証
		if err := vs.validateVideoID(inputURL); err == nil {
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
	if err := vs.validateVideoID(videoID); err != nil {
		return "", fmt.Errorf("extracted invalid video ID '%s': %w", videoID, err)
	}

	return videoID, nil
}

// ValidateLiveStream 動画がライブチャット監視に適しているかを確認します
func (vs *VideoService) ValidateLiveStream(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	correlationID := fmt.Sprintf("validate-%s", videoID)

	videoInfo, err := vs.GetVideoInfo(ctx, videoID)
	if err != nil {
		return nil, err
	}

	// ライブ配信状態を確認
	if !vs.isLiveStreamSupported(videoInfo.LiveBroadcastContent) {
		err := fmt.Errorf("video is not in supported live streaming state: %s", videoInfo.LiveBroadcastContent)
		vs.logger.LogError("ERROR", "Unsupported live streaming state", videoID, correlationID, err, map[string]interface{}{
			"liveBroadcastContent": videoInfo.LiveBroadcastContent,
		})
		return nil, err
	}

	// ライブチャットが利用可能かを確認
	if videoInfo.LiveStreamingDetails.ActiveLiveChatID == "" {
		err := fmt.Errorf("live chat is not available for this video")
		vs.logger.LogError("ERROR", "No active live chat", videoID, correlationID, err, map[string]interface{}{
			"liveBroadcastContent": videoInfo.LiveBroadcastContent,
		})
		return nil, err
	}

	vs.logger.LogAPI("INFO", "Live stream validation successful", videoID, correlationID, map[string]interface{}{
		"liveBroadcastContent": videoInfo.LiveBroadcastContent,
		"activeLiveChatId":     videoInfo.LiveStreamingDetails.ActiveLiveChatID,
	})

	return videoInfo, nil
}

// validateVideoID YouTube動画IDの形式を検証します
func (vs *VideoService) validateVideoID(videoID string) error {
	if videoID == "" {
		return fmt.Errorf("video ID cannot be empty")
	}

	// YouTube動画IDは通常11文字で、英数字、ハイフン、アンダースコアを含みます
	if len(videoID) != constants.YouTubeVideoIDLength {
		return fmt.Errorf("invalid video ID length: expected %d characters, got %d", constants.YouTubeVideoIDLength, len(videoID))
	}

	// 有効な文字をチェック（英数字、ハイフン、アンダースコア）
	for _, char := range videoID {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return fmt.Errorf("invalid character in video ID: %c", char)
		}
	}

	return nil
}

// isLiveStreamSupported ライブ配信コンテンツタイプがサポートされているかを確認します
func (vs *VideoService) isLiveStreamSupported(liveBroadcastContent string) bool {
	switch liveBroadcastContent {
	case "active", "live":
		return true
	case "upcoming":
		// 将来的にサポートされる可能性があるが、現在はサポートされていません
		return false
	case "none":
		return false
	default:
		return false
	}
}

// GetLiveStreamStatus 現在のライブ配信ステータス情報を返します
func (vs *VideoService) GetLiveStreamStatus(ctx context.Context, videoID string) (map[string]interface{}, error) {
	videoInfo, err := vs.GetVideoInfo(ctx, videoID)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"videoId":              videoID,
		"title":                videoInfo.Title,
		"channelTitle":         videoInfo.ChannelTitle,
		"liveBroadcastContent": videoInfo.LiveBroadcastContent,
		"isLiveStreamActive":   vs.isLiveStreamSupported(videoInfo.LiveBroadcastContent),
		"hasActiveLiveChat":    videoInfo.LiveStreamingDetails.ActiveLiveChatID != "",
		"activeLiveChatId":     videoInfo.LiveStreamingDetails.ActiveLiveChatID,
		"actualStartTime":      videoInfo.LiveStreamingDetails.ActualStartTime,
		"actualEndTime":        videoInfo.LiveStreamingDetails.ActualEndTime,
		"concurrentViewers":    videoInfo.LiveStreamingDetails.ConcurrentViewers,
		"scheduledStartTime":   videoInfo.LiveStreamingDetails.ScheduledStartTime,
	}

	return status, nil
}
