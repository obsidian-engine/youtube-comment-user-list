// Package service アプリケーション層のサービスを定義します
package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// VideoService YouTube動画情報を管理するサービスです
type VideoService struct {
	youtubeClient repository.YouTubeClient
	logger        repository.Logger

	mu    sync.RWMutex
	cache map[string]*cachedVideo
}

type cachedVideo struct {
	info               *entity.VideoInfo
	err                error
	fetchedAt          time.Time
	quotaExceededUntil time.Time
}

// NewVideoService 新しいVideoServiceを作成します
func NewVideoService(youtubeClient repository.YouTubeClient, logger repository.Logger) *VideoService {
	return &VideoService{
		youtubeClient: youtubeClient,
		logger:        logger,
		cache:         make(map[string]*cachedVideo),
	}
}

// GetVideoInfo 動画情報を取得します（常に最新取得 / キャッシュ無視）
func (vs *VideoService) GetVideoInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	correlationID := fmt.Sprintf("video-%s", videoID)

	vs.logger.LogAPI("INFO", "Fetching video info", videoID, correlationID, map[string]interface{}{
		"operation": "get_video_info",
	})

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

// getVideoInfoCached キャッシュ＋TTL＋quotaExceeded バックオフを用いて���画情報を取得
// ttl: 再フェッチしない最小期間
// backoffOnQuota: クォータ超過時に設定するバックオフ期間
func (vs *VideoService) getVideoInfoCached(ctx context.Context, videoID string, ttl, backoffOnQuota time.Duration) (*entity.VideoInfo, error) {
	vs.mu.RLock()
	cv, ok := vs.cache[videoID]
	if ok {
		// quotaExceeded バックオフ中なら即座に前回成功結果（あれば）を返し、なければエラー
		if time.Now().Before(cv.quotaExceededUntil) {
			info := cv.info
			vs.mu.RUnlock()
			if info != nil {
				return info, nil
			}
			return nil, cv.err
		}
		// TTL 内で前回取得が成功していれば再利用
		if cv.info != nil && cv.err == nil && time.Since(cv.fetchedAt) < ttl {
			info := cv.info
			vs.mu.RUnlock()
			return info, nil
		}
	}
	vs.mu.RUnlock()

	// ここから更新取得
	info, err := vs.youtubeClient.FetchVideoInfo(ctx, videoID)

	vs.mu.Lock()
	defer vs.mu.Unlock()
	entry := &cachedVideo{info: info, err: err, fetchedAt: time.Now()}
	if err != nil {
		// quotaExceeded 判別
		if strings.Contains(strings.ToLower(err.Error()), "quotaexceeded") {
			entry.quotaExceededUntil = time.Now().Add(backoffOnQuota)
			// 既存成功データがあれば温存
			if ok && cv != nil && cv.info != nil && cv.err == nil {
				vs.cache[videoID] = &cachedVideo{info: cv.info, err: nil, fetchedAt: cv.fetchedAt, quotaExceededUntil: entry.quotaExceededUntil}
				return cv.info, nil // フォールバックで前回値返す
			}
		}
		vs.cache[videoID] = entry
		return nil, err
	}
	// 成功時キャッシュ保存
	vs.cache[videoID] = entry
	return info, nil
}

// ExtractVideoIDFromURL URLからYouTube動画IDを抽出します
func (vs *VideoService) ExtractVideoIDFromURL(inputURL string) (string, error) {
	return entity.ExtractVideoIDFromURL(inputURL)
}

// ValidateLiveStream ライブ配信が有効かどうかを検証します（常に最新取得）
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
// キャッシュ: 60秒 TTL, quotaExceeded 時 15分バックオフ
func (vs *VideoService) GetLiveStreamStatus(ctx context.Context, videoID string) (string, error) {
	info, err := vs.getVideoInfoCached(ctx, videoID, 60*time.Second, 15*time.Minute)
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", fmt.Errorf("video info unavailable")
	}
	return info.LiveBroadcastContent, nil
}
