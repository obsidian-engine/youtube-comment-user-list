package service

import (
	"context"
	"fmt"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// PollingService ライブチャットポーリングのビジネスロジックを処理します
type PollingService struct {
	youtubeClient YouTubeClient
	chatRepo      ChatRepository
	logger        Logger
	eventPub      EventPublisher
}

// NewPollingService 新しいPollingServiceを作成します
func NewPollingService(
	youtubeClient YouTubeClient,
	chatRepo ChatRepository,
	logger Logger,
	eventPub EventPublisher,
) *PollingService {
	return &PollingService{
		youtubeClient: youtubeClient,
		chatRepo:      chatRepo,
		logger:        logger,
		eventPub:      eventPub,
	}
}

// StartPolling 動画のライブチャットポーリングプロセスを開始します
func (ps *PollingService) StartPolling(ctx context.Context, videoID string, messagesChan chan<- entity.ChatMessage) error {
	correlationID := fmt.Sprintf("poll-%s-%d", videoID, time.Now().Unix())

	ps.logger.LogPoller("INFO", "Starting polling", videoID, correlationID, map[string]interface{}{
		"event": "polling_started",
	})

	// 動画情報を取得してライブ配信を検証
	videoInfo, err := ps.youtubeClient.FetchVideoInfo(ctx, videoID)
	if err != nil {
		ps.logger.LogError("ERROR", "Failed to fetch video info", videoID, correlationID, err, nil)
		return fmt.Errorf("failed to fetch video info: %w", err)
	}

	if !ps.isLiveStreamActive(videoInfo) {
		err := fmt.Errorf("video is not in active live streaming state: %s", videoInfo.LiveBroadcastContent)
		ps.logger.LogError("ERROR", "Invalid live streaming state", videoID, correlationID, err, map[string]interface{}{
			"liveBroadcastContent": videoInfo.LiveBroadcastContent,
		})
		return err
	}

	liveChatID := videoInfo.LiveStreamingDetails.ActiveLiveChatID
	if liveChatID == "" {
		err := fmt.Errorf("activeLiveChatId is empty")
		ps.logger.LogError("ERROR", "No active live chat ID", videoID, correlationID, err, nil)
		return err
	}

	ps.logger.LogPoller("INFO", "Live chat validation successful", videoID, correlationID, map[string]interface{}{
		"liveChatId": liveChatID,
		"title":      videoInfo.Title,
		"channel":    videoInfo.ChannelTitle,
	})

	// ポーリングループを開始
	return ps.pollMessages(ctx, videoID, liveChatID, messagesChan, correlationID)
}

// isLiveStreamActive 動画がアクティブなライブ配信状態かを確認します
func (ps *PollingService) isLiveStreamActive(videoInfo *entity.VideoInfo) bool {
	switch videoInfo.LiveBroadcastContent {
	case "active", "live":
		return true
	default:
		return false
	}
}

// pollMessages ライブチャットメッセージの継続的なポーリングを処理します
func (ps *PollingService) pollMessages(ctx context.Context, videoID, liveChatID string, messagesChan chan<- entity.ChatMessage, correlationID string) error {
	var pageToken string
	consecutiveErrors := constants.InitialCounterValue
	maxConsecutiveErrors := constants.MaxConsecutiveErrors
	baseWaitTime := constants.PollingBaseWaitTime
	maxWaitTime := constants.PollingMaxWaitTime

	for {
		select {
		case <-ctx.Done():
			ps.logger.LogPoller("INFO", "Polling cancelled", videoID, correlationID, nil)
			return ctx.Err()
		default:
			// 新しいメッセージをポーリング
			pollResult, err := ps.youtubeClient.FetchLiveChat(ctx, liveChatID, pageToken)
			if err != nil {
				consecutiveErrors++
				waitTime := ps.calculateBackoffTime(consecutiveErrors, baseWaitTime, maxWaitTime)

				ps.logger.LogError("ERROR", "Failed to fetch live chat", videoID, correlationID, err, map[string]interface{}{
					"consecutiveErrors": consecutiveErrors,
					"waitTime":          waitTime.String(),
				})

				if consecutiveErrors >= maxConsecutiveErrors {
					ps.logger.LogError("FATAL", "Max consecutive errors reached", videoID, correlationID, err, map[string]interface{}{
						"maxErrors": maxConsecutiveErrors,
					})
					return fmt.Errorf("max consecutive errors reached: %w", err)
				}

				time.Sleep(waitTime)
				continue
			}

			// 成功時にエラーカウントをリセット
			consecutiveErrors = constants.InitialCounterValue

			// メッセージを処理
			if len(pollResult.Messages) > 0 {
				ps.logger.LogPoller("INFO", "Messages received", videoID, correlationID, map[string]interface{}{
					"messageCount": len(pollResult.Messages),
				})

				// メッセージをリポジトリに保存
				if err := ps.chatRepo.SaveChatMessages(ctx, pollResult.Messages); err != nil {
					ps.logger.LogError("ERROR", "Failed to save messages", videoID, correlationID, err, nil)
				}

				// メッセージをチャネルに送信してイベントを発行
				for _, message := range pollResult.Messages {
					select {
					case messagesChan <- message:
						// メッセージ処理のイベントを発行
						if err := ps.eventPub.PublishChatMessage(ctx, message); err != nil {
							ps.logger.LogError("ERROR", "Failed to publish message event", videoID, correlationID, err, nil)
						}
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

			// 次のリクエスト用にページトークンを更新
			pageToken = pollResult.NextPageToken

			// 指定されたポーリング間隔を待機
			waitTime := time.Duration(pollResult.PollingIntervalMs) * time.Millisecond
			if waitTime < time.Second {
				waitTime = time.Second // 最小1秒
			}

			ps.logger.LogPoller("DEBUG", "Polling interval wait", videoID, correlationID, map[string]interface{}{
				"waitTime":      waitTime.String(),
				"nextPageToken": pageToken,
				"messageCount":  len(pollResult.Messages),
			})

			select {
			case <-time.After(waitTime):
				// ポーリングを継続
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// calculateBackoffTime エラー処理のための指数バックオフ時間を計算します
func (ps *PollingService) calculateBackoffTime(errorCount int, baseTime, maxTime time.Duration) time.Duration {
	backoffTime := baseTime
	for i := constants.BackoffInitialStep; i < errorCount; i++ {
		backoffTime *= constants.ExponentialBackoffMultiplier
		if backoffTime > maxTime {
			return maxTime
		}
	}
	return backoffTime
}

// GetPollingStatus 動画のポーリングの現在のステータスを返します
func (ps *PollingService) GetPollingStatus(ctx context.Context, videoID string) (map[string]interface{}, error) {
	// 通常はレジストリの状態をチェックしますが、現在は基本情報を返します
	return map[string]interface{}{
		"videoId": videoID,
		"status":  "active", // これは実際のポーリング状態によって決定されます
	}, nil
}
