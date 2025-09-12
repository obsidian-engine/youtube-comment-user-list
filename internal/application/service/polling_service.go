// Package service アプリケーション層のサービスを定義します
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// PollingService YouTubeライブチャットのポーリングを管理するサービスです
type PollingService struct {
	youtubeClient repository.YouTubeClient
	chatRepo      repository.ChatRepository
	logger        repository.Logger
	eventPub      repository.EventPublisher
}

// NewPollingService 新しいPollingServiceを作成します
func NewPollingService(
	youtubeClient repository.YouTubeClient,
	chatRepo repository.ChatRepository,
	logger repository.Logger,
	eventPub repository.EventPublisher,
) *PollingService {
	return &PollingService{
		youtubeClient: youtubeClient,
		chatRepo:      chatRepo,
		logger:        logger,
		eventPub:      eventPub,
	}
}

// StartPolling ライブチャットのポーリングを開始します
func (ps *PollingService) StartPolling(ctx context.Context, liveChatID string, videoID string, messagesChan chan<- entity.ChatMessage) error {
	correlationID := fmt.Sprintf("poll-%s", videoID)

	ps.logger.LogPoller("INFO", "Starting chat polling", videoID, correlationID, map[string]interface{}{
		"liveChatID": liveChatID,
		"operation":  "start_polling",
	})

	var nextPageToken string
	consecutiveErrors := 0
	maxErrors := 5

	for {
		select {
		case <-ctx.Done():
			ps.logger.LogPoller("INFO", "Polling stopped by context", videoID, correlationID, map[string]interface{}{
				"operation": "stop_polling",
				"reason":    ctx.Err().Error(),
			})
			return ctx.Err()

		default:
			// ライブ配信の状態をチェック
			if !ps.isLiveStreamActive(ctx, videoID, correlationID) {
				ps.logger.LogPoller("INFO", "Live stream is no longer active", videoID, correlationID, map[string]interface{}{
					"operation": "check_stream_status",
				})
				return fmt.Errorf("live stream is no longer active")
			}

			// チャットメッセージをポーリング
			messages, newPageToken, pollingInterval, err := ps.pollMessages(ctx, liveChatID, nextPageToken, videoID, correlationID)
			if err != nil {
				consecutiveErrors++
				ps.logger.LogError("ERROR", fmt.Sprintf("Polling error (attempt %d/%d)", consecutiveErrors, maxErrors), videoID, correlationID, err, map[string]interface{}{
					"operation":         "poll_messages",
					"consecutiveErrors": consecutiveErrors,
				})

				if consecutiveErrors >= maxErrors {
					ps.logger.LogPoller("ERROR", "Max consecutive errors reached, stopping polling", videoID, correlationID, map[string]interface{}{
						"operation":         "stop_polling",
						"consecutiveErrors": consecutiveErrors,
						"maxErrors":         maxErrors,
					})
					return fmt.Errorf("max consecutive polling errors reached: %w", err)
				}

				// エラー時のバックオフ
				backoffTime := ps.calculateBackoffTime(consecutiveErrors)
				ps.logger.LogPoller("WARN", fmt.Sprintf("Backing off for %v before retry", backoffTime), videoID, correlationID, map[string]interface{}{
					"operation":    "backoff",
					"backoffTime":  backoffTime.String(),
					"errorAttempt": consecutiveErrors,
				})

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoffTime):
					continue
				}
			}

			// エラーがない場合、エラーカウンタをリセット
			consecutiveErrors = 0
			nextPageToken = newPageToken

			// メッセージを処理
			for _, message := range messages {
				// メッセージをチャンネルに送信
				select {
				case messagesChan <- message:
				case <-ctx.Done():
					return ctx.Err()
				}

				// イベントを発行
				if err := ps.eventPub.PublishChatMessage(ctx, message); err != nil {
					ps.logger.LogError("WARN", "Failed to publish chat message event", videoID, correlationID, err, map[string]interface{}{
						"operation": "publish_event",
						"messageID": message.ID,
					})
				}
			}

			// チャットリポジトリにメッセージを保存
			if len(messages) > 0 {
				if err := ps.chatRepo.SaveChatMessages(ctx, messages); err != nil {
					ps.logger.LogError("WARN", "Failed to save chat messages", videoID, correlationID, err, map[string]interface{}{
						"operation":    "save_messages",
						"messageCount": len(messages),
					})
				}
			}

			ps.logger.LogPoller("DEBUG", "Polling cycle completed", videoID, correlationID, map[string]interface{}{
				"operation":       "poll_cycle",
				"messageCount":    len(messages),
				"pollingInterval": pollingInterval,
			})

			// 次のポーリングまで待機
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(pollingInterval) * time.Millisecond):
			}
		}
	}
}

// isLiveStreamActive ライブ配信がアクティブかどうかをチェックします
func (ps *PollingService) isLiveStreamActive(ctx context.Context, videoID, correlationID string) bool {
	// 簡単な実装として、YouTube APIで動画情報を取得してライブ状態をチェック
	// 実際の実装では、より効率的な方法を検討する必要があります
	return true // 一旦trueを返す
}

// pollMessages チャットメッセージをポーリングします
func (ps *PollingService) pollMessages(ctx context.Context, liveChatID, pageToken, videoID, correlationID string) ([]entity.ChatMessage, string, int, error) {
	ps.logger.LogPoller("DEBUG", "Polling messages", videoID, correlationID, map[string]interface{}{
		"operation":  "poll_messages",
		"liveChatID": liveChatID,
		"pageToken":  pageToken,
	})

	result, err := ps.youtubeClient.FetchLiveChat(ctx, liveChatID, pageToken)
	if err != nil {
		return nil, "", constants.DefaultPollingIntervalMs, fmt.Errorf("failed to fetch live chat: %w", err)
	}

	if result.Error != nil {
		return nil, "", constants.DefaultPollingIntervalMs, result.Error
	}

	ps.logger.LogPoller("DEBUG", "Messages fetched successfully", videoID, correlationID, map[string]interface{}{
		"operation":         "poll_messages",
		"messageCount":      len(result.Messages),
		"nextPageToken":     result.NextPageToken,
		"pollingIntervalMs": result.PollingIntervalMs,
	})

	return result.Messages, result.NextPageToken, result.PollingIntervalMs, nil
}

// calculateBackoffTime エラー数に基づいてバックオフ時間を計算します
func (ps *PollingService) calculateBackoffTime(consecutiveErrors int) time.Duration {
	// 指数バックオフ: 2^consecutiveErrors 秒（最大60秒）
	backoffSeconds := math.Pow(2, float64(consecutiveErrors))
	if backoffSeconds > 60 {
		backoffSeconds = 60
	}
	return time.Duration(backoffSeconds) * time.Second
}

// GetPollingStatus ポーリングの現在の状態を取得します
func (ps *PollingService) GetPollingStatus(ctx context.Context, videoID string) (bool, error) {
	// 実際の実装では、ポーリング状態を追跡する必要があります
	// 今は簡単な実装として false を返します
	return false, nil
}
