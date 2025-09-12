// Package usecase アプリケーションのビジネスロジックとユースケースを実装します
package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// ChatMonitoringUseCase ライブチャット監視ワークフローを処理します
type ChatMonitoringUseCase struct {
	pollingService *service.PollingService
	userService    *service.UserService
	videoService   *service.VideoService
	logger         service.Logger

	// アクティブなポーリングセッション
	sessions map[string]*MonitoringSession
	mu       sync.RWMutex
}

// MonitoringSession 動画のアクティブな監視セッションを表します
type MonitoringSession struct {
	VideoID      string
	Cancel       context.CancelFunc
	MessagesChan chan entity.ChatMessage
	Subscribers  int
	UserList     *entity.UserList
}

// NewChatMonitoringUseCase 新しいChatMonitoringUseCaseを作成します
func NewChatMonitoringUseCase(
	pollingService *service.PollingService,
	userService *service.UserService,
	videoService *service.VideoService,
	logger service.Logger,
) *ChatMonitoringUseCase {
	return &ChatMonitoringUseCase{
		pollingService: pollingService,
		userService:    userService,
		videoService:   videoService,
		logger:         logger,
		sessions:       make(map[string]*MonitoringSession),
	}
}

// StartMonitoring 動画のライブチャット監視を開始します
func (uc *ChatMonitoringUseCase) StartMonitoring(ctx context.Context, videoInput string, maxUsers int) (*MonitoringSession, error) {
	correlationID := fmt.Sprintf("start-monitoring-%s", videoInput)

	// 動画IDを抽出して検証
	videoID, err := uc.videoService.ExtractVideoIDFromURL(videoInput)
	if err != nil {
		uc.logger.LogError("ERROR", "Failed to extract video ID", videoInput, correlationID, err, map[string]interface{}{
			"input": videoInput,
		})
		return nil, fmt.Errorf("failed to extract video ID: %w", err)
	}

	// ライブ配信を検証
	videoInfo, err := uc.videoService.ValidateLiveStream(ctx, videoID)
	if err != nil {
		uc.logger.LogError("ERROR", "Live stream validation failed", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("live stream validation failed: %w", err)
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	// この動画を既に監視しているかチェック
	if session, exists := uc.sessions[videoID]; exists {
		// 購読者数を増加
		session.Subscribers++
		uc.logger.LogStructured("INFO", "monitoring", "subscriber_added", "Added subscriber to existing session", videoID, correlationID, map[string]interface{}{
			"subscribers": session.Subscribers,
		})
		return session, nil
	}

	// 新しい監視セッションを作成
	sessionCtx, cancel := context.WithCancel(ctx)
	messagesChan := make(chan entity.ChatMessage, 100)

	// この動画用のユーザーリストを作成
	userList, err := uc.userService.CreateUserList(ctx, videoID, maxUsers)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create user list: %w", err)
	}

	session := &MonitoringSession{
		VideoID:      videoID,
		Cancel:       cancel,
		MessagesChan: messagesChan,
		Subscribers:  1,
		UserList:     userList,
	}

	uc.sessions[videoID] = session

	uc.logger.LogStructured("INFO", "monitoring", "session_started", "Started new monitoring session", videoID, correlationID, map[string]interface{}{
		"title":        videoInfo.Title,
		"channelTitle": videoInfo.ChannelTitle,
		"maxUsers":     maxUsers,
	})

	// バックグラウンドでポーリングを開始
	go uc.runPolling(sessionCtx, videoID, messagesChan, correlationID)

	// メッセージ処理を開始
	go uc.processMessages(sessionCtx, videoID, messagesChan, correlationID)

	return session, nil
}

// StopMonitoring 動画の監視を停止します
func (uc *ChatMonitoringUseCase) StopMonitoring(videoID string) error {
	correlationID := fmt.Sprintf("stop-monitoring-%s", videoID)

	uc.mu.Lock()
	defer uc.mu.Unlock()

	session, exists := uc.sessions[videoID]
	if !exists {
		return fmt.Errorf("no active monitoring session for video: %s", videoID)
	}

	session.Subscribers--

	if session.Subscribers <= 0 {
		// セッションを停止
		session.Cancel()
		close(session.MessagesChan)
		delete(uc.sessions, videoID)

		uc.logger.LogStructured("INFO", "monitoring", "session_stopped", "Stopped monitoring session", videoID, correlationID, nil)
	} else {
		uc.logger.LogStructured("INFO", "monitoring", "subscriber_removed", "Removed subscriber from session", videoID, correlationID, map[string]interface{}{
			"remainingSubscribers": session.Subscribers,
		})
	}

	return nil
}

// GetMonitoringSession 動画の現在の監視セッションを返します
func (uc *ChatMonitoringUseCase) GetMonitoringSession(videoID string) (*MonitoringSession, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	session, exists := uc.sessions[videoID]
	return session, exists
}

// GetUserList 監視中の動画のユーザーリストを返します
func (uc *ChatMonitoringUseCase) GetUserList(ctx context.Context, videoID string) ([]*entity.User, error) {
	return uc.userService.GetUserListSnapshot(ctx, videoID)
}

// GetActiveVideos 現在監視中の動画リストを返します
func (uc *ChatMonitoringUseCase) GetActiveVideos() []string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	videos := make([]string, 0, len(uc.sessions))
	for videoID := range uc.sessions {
		videos = append(videos, videoID)
	}
	return videos
}

// runPolling 動画のポーリングループを処理します
func (uc *ChatMonitoringUseCase) runPolling(ctx context.Context, videoID string, messagesChan chan<- entity.ChatMessage, correlationID string) {
	if err := uc.pollingService.StartPolling(ctx, videoID, messagesChan); err != nil {
		uc.logger.LogError("ERROR", "Polling ended with error", videoID, correlationID, err, nil)
	}
}

// processMessages 受信したチャットメッセージを処理します
func (uc *ChatMonitoringUseCase) processMessages(ctx context.Context, videoID string, messagesChan <-chan entity.ChatMessage, correlationID string) {
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-messagesChan:
			if !ok {
				// チャンネルがクローズされた
				return
			}

			// ユーザーサービス経由でメッセージを処理
			if err := uc.userService.ProcessChatMessage(ctx, message); err != nil {
				uc.logger.LogError("ERROR", "Failed to process chat message", videoID, correlationID, err, map[string]interface{}{
					"messageId":   message.ID,
					"channelId":   message.AuthorDetails.ChannelID,
					"displayName": message.AuthorDetails.DisplayName,
				})
			}
		}
	}
}

// GetVideoStatus 動画のステータス情報を返します
func (uc *ChatMonitoringUseCase) GetVideoStatus(ctx context.Context, videoID string) (map[string]interface{}, error) {
	// 基本的な動画ステータスを取得
	videoStatus, err := uc.videoService.GetLiveStreamStatus(ctx, videoID)
	if err != nil {
		return nil, err
	}

	// 監視セッション情報を追加
	uc.mu.RLock()
	session, isMonitoring := uc.sessions[videoID]
	uc.mu.RUnlock()

	videoStatus["isMonitoring"] = isMonitoring
	if isMonitoring {
		videoStatus["subscribers"] = session.Subscribers
		videoStatus["userCount"] = session.UserList.Count()
		videoStatus["userListFull"] = session.UserList.IsFull()
	}

	return videoStatus, nil
}
