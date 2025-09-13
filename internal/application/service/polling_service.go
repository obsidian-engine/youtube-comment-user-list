// Package service アプリケーション層のサービスを定義します
package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
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

	videoService *VideoService
	// autoEndEnabled: videoID -> enabled
	autoEndEnabled map[string]bool
	autoMu         sync.RWMutex
}

// NewPollingService 新しいPollingServiceを作成します
func NewPollingService(
	youtubeClient repository.YouTubeClient,
	chatRepo repository.ChatRepository,
	logger repository.Logger,
	eventPub repository.EventPublisher,
	videoService *VideoService,
) *PollingService {
	return &PollingService{
		youtubeClient:  youtubeClient,
		chatRepo:       chatRepo,
		logger:         logger,
		eventPub:       eventPub,
		videoService:   videoService,
		autoEndEnabled: make(map[string]bool),
	}
}

// StartPolling ライブチャットのポーリングを開始します
// 実装を簡素化し、1つのループでメッセージ取得、送信、保存、バックオフを行います。
func (ps *PollingService) StartPolling(ctx context.Context, liveChatID string, videoID string, messagesChan chan<- entity.ChatMessage) error {
	correlationID := fmt.Sprintf("poll-%s", videoID)
	ps.logger.LogPoller("INFO", "Starting chat polling", videoID, correlationID, map[string]interface{}{
		"liveChatID": liveChatID,
		"operation":  "start_polling",
	})

	var nextPageToken string
	consecutiveErrors := 0
	maxErrors := 5
	pollingIntervalMs := constants.DefaultPollingIntervalMs

	for {
		// キャンセル優先
		select {
		case <-ctx.Done():
			ps.logger.LogPoller("INFO", "Polling stopped by context", videoID, correlationID, map[string]interface{}{
				"operation": "stop_polling",
				"reason":    ctx.Err(),
			})
			// 上位でキャンセルは正常終了と扱う
			return ctx.Err()
		default:
		}

		// ライブ配信の状態をチェック（自動終了検知が有効な場合）
		if !ps.isLiveStreamActive(ctx, videoID, correlationID) {
			ps.logger.LogPoller("INFO", "Live stream is no longer active (auto-end)", videoID, correlationID, map[string]interface{}{
				"operation": "check_stream_status",
			})
			// メッセージチャンネルをクローズして下流へ通知
			// 注意: StartPolling の所有下でのみ close する
			defer func() {
				// recoverで二重closeを回避
				defer func() { _ = recover() }()
				close(messagesChan)
			}()
			return fmt.Errorf("live stream is no longer active")
		}

		// メッセージ取得を試みる
		messages, newPageToken, pollingInterval, err := ps.pollMessages(ctx, liveChatID, nextPageToken, videoID, correlationID)
		if err != nil {
			// コンテキスト由来のエラーは正常終了扱い
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				ps.logger.LogPoller("INFO", "Polling stopped (context closed)", videoID, correlationID, map[string]interface{}{
					"operation": "stop_polling",
					"reason":    err.Error(),
				})
				return nil
			}

			consecutiveErrors++
			ps.logger.LogError("ERROR", fmt.Sprintf("Polling error (attempt %d/%d): %v", consecutiveErrors, maxErrors, err), videoID, correlationID, err, map[string]interface{}{
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

			// バックオフしてリトライ
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

		// 成功パス: エラーカウントをリセットし、次ページトークンと間隔を更新
		consecutiveErrors = 0
		nextPageToken = newPageToken
		if pollingInterval > 0 {
			// 下限強制（過剰ポーリングによるQuota超過防止）
			if pollingInterval < constants.MinPollingIntervalMs {
				ps.logger.LogPoller("DEBUG", "Adjust polling interval to minimum threshold", videoID, correlationID, map[string]interface{}{
					"operation":             "adjust_polling_interval",
					"requestedIntervalMs":   pollingInterval,
					"appliedIntervalMs":     constants.MinPollingIntervalMs,
					"minIntervalEnforcedMs": constants.MinPollingIntervalMs,
				})
				pollingIntervalMs = constants.MinPollingIntervalMs
			} else {
				pollingIntervalMs = pollingInterval
			}
		}

		// メッセージを順次処理
		for _, message := range messages {
			message.VideoID = videoID

			select {
			case messagesChan <- message:
			case <-ctx.Done():
				return ctx.Err()
			}

			if err := ps.eventPub.PublishChatMessage(ctx, message); err != nil {
				ps.logger.LogError("WARN", "Failed to publish chat message event", videoID, correlationID, err, map[string]interface{}{
					"operation": "publish_event",
					"messageID": message.ID,
				})
			}
		}

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
			"pollingInterval": pollingIntervalMs,
		})

		// 固定間隔で待機（APIが返す間隔を尊重）
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(pollingIntervalMs) * time.Millisecond):
		}
	}
}

// isLiveStreamActive ライブ配信がアクティブかをチェック（自動終了検知が有効なときのみ）
func (ps *PollingService) isLiveStreamActive(ctx context.Context, videoID, correlationID string) bool {
	if !ps.IsAutoEndEnabled(videoID) {
		return true
	}
	if ps.videoService == nil {
		// フォールバック: 無効扱いにしない
		return true
	}
	status, err := ps.videoService.GetLiveStreamStatus(ctx, videoID)
	if err != nil {
		ps.logger.LogError("WARN", "Failed to get live stream status (auto-end check)", videoID, correlationID, err, nil)
		return true // ステータス取得失敗時は継続
	}
	// アクティブは "live" のみと判定
	return status == "live"
}

// SetAutoEndEnabled 自動終了検知の有効/無効を設定
func (ps *PollingService) SetAutoEndEnabled(videoID string, enabled bool) {
	ps.autoMu.Lock()
	ps.autoEndEnabled[videoID] = enabled
	ps.autoMu.Unlock()
}

// IsAutoEndEnabled 自動終了検知の状態を取得
func (ps *PollingService) IsAutoEndEnabled(videoID string) bool {
	ps.autoMu.RLock()
	enabled, ok := ps.autoEndEnabled[videoID]
	ps.autoMu.RUnlock()
	if !ok {
		// 既定: 有効
		return true
	}
	return enabled
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
