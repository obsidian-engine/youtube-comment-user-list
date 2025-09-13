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

const (
	validationSuccessTTL = 120 * time.Second // 成功結果は2分キャッシュ
	validationFailureTTL = 60 * time.Second  // 失敗結果も1分キャッシュして過剰叩きを防止
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

// getVideoInfoCached キャッシュ＋TTL＋quotaExceeded バックオフを用いて動画情報を取得
// ttl: 再フェッチしない最小期間
// backoffOnQuota: クォータ超過時に設定するバックオフ期間
func (vs *VideoService) getVideoInfoCached(ctx context.Context, videoID string, ttl, backoffOnQuota time.Duration) (*entity.VideoInfo, error) {
	vs.mu.RLock()
	cv, ok := vs.cache[videoID]
	if ok {
		// quotaExceeded バックオフ中
		if time.Now().Before(cv.quotaExceededUntil) {
			info := cv.info
			vs.mu.RUnlock()
			if info != nil {
				return info, nil
			}
			return nil, cv.err
		}
		age := time.Since(cv.fetchedAt)
		if cv.info != nil && cv.err == nil && age < ttl {
			info := cv.info
			vs.mu.RUnlock()
			return info, nil
		}
		// 失敗結果も一定時間は再試行しない
		if cv.info == nil && cv.err != nil && age < validationFailureTTL {
			errCopy := cv.err
			vs.mu.RUnlock()
			return nil, errCopy
		}
	}
	vs.mu.RUnlock()

	// ここから更新取得
	info, err := vs.youtubeClient.FetchVideoInfo(ctx, videoID)

	vs.mu.Lock()
	defer vs.mu.Unlock()
	entry := &cachedVideo{info: info, err: err, fetchedAt: time.Now()}
	if err != nil {
		// quotaExceeded 判別（表記ゆれ吸収）
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "quotaexceeded") || strings.Contains(low, "quota exceeded") {
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

// ValidateLiveStreamAndGetInfo ライブ配信を検証し、同時に VideoInfo を返します（1回の取得で完了）
func (vs *VideoService) ValidateLiveStreamAndGetInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	correlationID := fmt.Sprintf("validate-%s", videoID)
	vs.logger.LogStructured("INFO", "video", "validate_live_stream", "Starting live stream validation", videoID, correlationID, map[string]interface{}{
		"operation": "validate_live_stream",
	})

	// 成功結果は長め(2分)・失敗は内部で 60秒 キャッシュ
	videoInfo, err := vs.getVideoInfoCached(ctx, videoID, validationSuccessTTL, 15*time.Minute)
	if err != nil {
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "quotaexceeded") || strings.Contains(low, "quota exceeded") {
			msg := "YouTube API のクォータを超過しています。しばらく（15分程度）待ってから再試行してください。"
			vs.logger.LogError("ERROR", "Quota exceeded when validating live stream", videoID, correlationID, err, map[string]interface{}{
				"operation":    "validate_live_stream",
				"user_message": msg,
			})
			return nil, fmt.Errorf(msg+": %w", err)
		}
		vs.logger.LogError("ERROR", "Failed to get video info for validation", videoID, correlationID, err, map[string]interface{}{
			"operation": "validate_live_stream",
		})
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	if !vs.isLiveStreamSupported(videoInfo.LiveBroadcastContent) {
		vs.logger.LogStructured("ERROR", "video", "validate_live_stream", "Not a live stream", videoID, correlationID, map[string]interface{}{
			"operation":     "validate_live_stream",
			"broadcastType": videoInfo.LiveBroadcastContent,
		})
		return nil, fmt.Errorf("video is not a live stream (type: %s)", videoInfo.LiveBroadcastContent)
	}

	if videoInfo.LiveStreamingDetails.ActiveLiveChatID == "" {
		vs.logger.LogStructured("ERROR", "video", "validate_live_stream", "No active live chat", videoID, correlationID, map[string]interface{}{
			"operation": "validate_live_stream",
		})
		return nil, fmt.Errorf("live chat is not available for this stream")
	}

	vs.logger.LogStructured("INFO", "video", "validate_live_stream", "Live stream validation successful", videoID, correlationID, map[string]interface{}{
		"operation":     "validate_live_stream",
		"liveChatID":    videoInfo.LiveStreamingDetails.ActiveLiveChatID,
		"broadcastType": videoInfo.LiveBroadcastContent,
	})

	return videoInfo, nil
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
