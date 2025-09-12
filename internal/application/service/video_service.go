// Package service アプリケーション層のサービスを定義します
package service

import (
	"context"
	"fmt"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// VideoService YouTube動画情報を管理するサービスです
type VideoService struct {
	youtubeClient repository.YouTubeClient
	logger        repository.Logger
}

// NewVideoService 新しいVideoServiceを作成します
func NewVideoService(youtubeClient repository.YouTubeClient, logger repository.Logger) *VideoService {
	return &VideoService{
		youtubeClient: youtubeClient,
		logger:        logger,
	}
}

// GetVideoInfo 動画情報を取得します
func (vs *VideoService) GetVideoInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	correlationID := fmt.Sprintf("video-%s", videoID)

	vs.logger.LogAPI("INFO", "Fetching video info", videoID, correlationID, map[string]interface{}{
		"operation": "get_video_info",
	})

	// YouTube APIから動画情報を取得
	videoInfo, err := vs.youtubeClient.FetchVideoInfo(ctx, videoID)
	if err != nil {
		vs.logger.LogError("ERROR", "Failed to fetch video info", videoID, correlationID, err, map[string]interface{}{
			"operation": "get_video_info",
		})
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

	vs.logger.LogAPI("INFO", "Successfully fetched video info", videoID, correlationID, map[string]interface{}{
		"operation":     "get_video_info",
		"title":         videoInfo.Title,
		"channelTitle":  videoInfo.ChannelTitle,
		"broadcastType": videoInfo.LiveBroadcastContent,
	})

	return videoInfo, nil
}

// ExtractVideoIDFromURL URLからYouTube動画IDを抽出します
func (vs *VideoService) ExtractVideoIDFromURL(inputURL string) (string, error) {
	return entity.ExtractVideoIDFromURL(inputURL)
}

// ValidateLiveStream ライブ配信が有効かどうかを検証します
func (vs *VideoService) ValidateLiveStream(ctx context.Context, videoID string) error {
	correlationID := fmt.Sprintf("validate-%s", videoID)

	vs.logger.LogStructured("INFO", "video", "validate_live_stream", "Starting live stream validation", videoID, correlationID, map[string]interface{}{
		"operation": "validate_live_stream",
	})

	videoInfo, err := vs.GetVideoInfo(ctx, videoID)
	if err != nil {
		vs.logger.LogError("ERROR", "Failed to get video info for validation", videoID, correlationID, err, map[string]interface{}{
			"operation": "validate_live_stream",
		})
		return fmt.Errorf("failed to get video info: %w", err)
	}

	if !vs.isLiveStreamSupported(videoInfo.LiveBroadcastContent) {
		vs.logger.LogStructured("ERROR", "video", "validate_live_stream", "Not a live stream", videoID, correlationID, map[string]interface{}{
			"operation":     "validate_live_stream",
			"broadcastType": videoInfo.LiveBroadcastContent,
		})
		return fmt.Errorf("video is not a live stream (type: %s)", videoInfo.LiveBroadcastContent)
	}

	if videoInfo.LiveStreamingDetails.ActiveLiveChatID == "" {
		vs.logger.LogStructured("ERROR", "video", "validate_live_stream", "No active live chat", videoID, correlationID, map[string]interface{}{
			"operation": "validate_live_stream",
		})
		return fmt.Errorf("live chat is not available for this stream")
	}

	vs.logger.LogStructured("INFO", "video", "validate_live_stream", "Live stream validation successful", videoID, correlationID, map[string]interface{}{
		"operation":     "validate_live_stream",
		"liveChatID":    videoInfo.LiveStreamingDetails.ActiveLiveChatID,
		"broadcastType": videoInfo.LiveBroadcastContent,
	})

	return nil
}

// isLiveStreamSupported サポートされているライブ配信タイプかどうかをチェックします
func (vs *VideoService) isLiveStreamSupported(broadcastContent string) bool {
	switch broadcastContent {
	case "live", "upcoming":
		return true
	default:
		return false
	}
}

// GetLiveStreamStatus ライブ配信のステータスを取得します
func (vs *VideoService) GetLiveStreamStatus(ctx context.Context, videoID string) (string, error) {
	videoInfo, err := vs.GetVideoInfo(ctx, videoID)
	if err != nil {
		return "", err
	}
	return videoInfo.LiveBroadcastContent, nil
}
