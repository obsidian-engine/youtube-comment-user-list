// Package usecase アプリケーションのビジネスロジックとユースケースを実装します
package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/service"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// ChatMonitoringUseCase ライブチャット監視ワークフローを処理します
type ChatMonitoringUseCase struct {
	pollingService *service.PollingService
	userService    *service.UserService
	videoService   *service.VideoService
	logger         repository.Logger

	// 単一の監視セッション
	currentSession *MonitoringSession
	// 最後に監視されたVideoID（セッション停止後でも保持）
	lastVideoID string
	mu          sync.RWMutex
}

// MonitoringSession 動画のアクティブな監視セッションを表します
type MonitoringSession struct {
	VideoID      string
	Cancel       context.CancelFunc
	MessagesChan chan entity.ChatMessage
	UserList     *entity.UserList
}

// NewChatMonitoringUseCase 新しいChatMonitoringUseCaseを作成します
func NewChatMonitoringUseCase(
	pollingService *service.PollingService,
	userService *service.UserService,
	videoService *service.VideoService,
	logger repository.Logger,
) *ChatMonitoringUseCase {
	return &ChatMonitoringUseCase{
		pollingService: pollingService,
		userService:    userService,
		videoService:   videoService,
		logger:         logger,
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
	err = uc.videoService.ValidateLiveStream(ctx, videoID)
	if err != nil {
		uc.logger.LogError("ERROR", "Live stream validation failed", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("live stream validation failed: %w", err)
	}

	// 動画情報を取得してliveChatIDを取得
	videoInfo, err := uc.videoService.GetVideoInfo(ctx, videoID)
	if err != nil {
		uc.logger.LogError("ERROR", "Failed to get video info", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	// 既に監視中の場合は停止
	if uc.currentSession != nil {
		uc.logger.LogStructured("INFO", "monitoring", "stopping_previous", "Stopping previous monitoring session", uc.currentSession.VideoID, correlationID, nil)
		uc.currentSession.Cancel()
		uc.currentSession = nil
	}

	// 新しい監視セッションを作成（バックグラウンドタスク用の独立したコンテキスト）
	sessionCtx, cancel := context.WithCancel(context.Background())
	messagesChan := make(chan entity.ChatMessage, constants.ChatMessageChannelBuffer)

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
		UserList:     userList,
	}

	uc.currentSession = session
	// 最後に監視したVideoIDを保存
	uc.lastVideoID = videoID

	uc.logger.LogStructured("INFO", "monitoring", "session_started", "Started new monitoring session", videoID, correlationID, map[string]interface{}{
		"maxUsers": maxUsers,
	})

	// バックグラウンドでポーリングを開始
	go uc.runPolling(sessionCtx, videoInfo.LiveStreamingDetails.ActiveLiveChatID, videoID, messagesChan, correlationID)

	// メッセージ処理を開始
	go uc.processMessages(sessionCtx, videoID, messagesChan, correlationID)

	return session, nil
}

// StopMonitoring 動画の監視を停止します
func (uc *ChatMonitoringUseCase) StopMonitoring() error {
	correlationID := "stop-monitoring"

	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.currentSession == nil {
		return fmt.Errorf("no active monitoring session")
	}

	videoID := uc.currentSession.VideoID
	uc.currentSession.Cancel()
	close(uc.currentSession.MessagesChan)
	uc.currentSession = nil

	uc.logger.LogStructured("INFO", "monitoring", "session_stopped", "Stopped monitoring session", videoID, correlationID, nil)
	return nil
}

// GetMonitoringSession 動画の現在の監視セッションを返します
func (uc *ChatMonitoringUseCase) GetMonitoringSession(videoID string) (*MonitoringSession, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	uc.mu.RLock()
	defer uc.mu.RUnlock()

	if uc.currentSession != nil && uc.currentSession.VideoID == videoID {
		return uc.currentSession, true
	}
	return nil, false
}

// GetActiveVideoID 現在監視中のvideoIDを取得します
func (uc *ChatMonitoringUseCase) GetActiveVideoID() (string, bool, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	if uc.currentSession != nil {
		// アクティブなセッションが存在
		return uc.currentSession.VideoID, true, true
	}

	if uc.lastVideoID != "" {
		// セッションは停止しているが最後のvideoIDが存在
		return uc.lastVideoID, false, true
	}

	// どちらも存在しない
	return "", false, false
}

// GetUserList 指定された動画のユーザーリストを返します
func (uc *ChatMonitoringUseCase) GetUserList(ctx context.Context, videoID string) ([]*entity.User, error) {
	return uc.userService.GetUserListSnapshot(ctx, videoID)
}

// runPolling 動画のポーリングループを処理します
func (uc *ChatMonitoringUseCase) runPolling(ctx context.Context, liveChatID, videoID string, messagesChan chan<- entity.ChatMessage, correlationID string) {
	if err := uc.pollingService.StartPolling(ctx, liveChatID, videoID, messagesChan); err != nil {
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
	streamStatus, err := uc.videoService.GetLiveStreamStatus(ctx, videoID)
	if err != nil {
		return nil, err
	}

	// ステータスマップを作成
	videoStatus := make(map[string]interface{})
	videoStatus["broadcastContent"] = streamStatus

	// 監視セッション情報を追加
	uc.mu.RLock()
	session := uc.currentSession
	isMonitoring := session != nil && session.VideoID == videoID
	uc.mu.RUnlock()

	videoStatus["isMonitoring"] = isMonitoring
	if isMonitoring {
		videoStatus["subscribers"] = 1 // 単一セッションのため固定値
		videoStatus["userCount"] = session.UserList.Count()
		videoStatus["userListFull"] = session.UserList.IsFull()
	}

	return videoStatus, nil
}
