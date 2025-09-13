// Package usecase アプリケーションのビジネスロジックとユースケースを実装します
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
        uc.logger.LogError(constants.LogLevelError, "Failed to extract video ID", videoInput, correlationID, err, map[string]interface{}{
			"input": videoInput,
		})
		return nil, fmt.Errorf("failed to extract video ID: %w", err)
	}

	// Validation with retry/backoff (transient errors only)
	var videoInfo *entity.VideoInfo
	var lastErr error
	maxAttempts := 5
	backoff := 500 * time.Millisecond
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		videoInfo, lastErr = uc.videoService.ValidateLiveStreamAndGetInfo(ctx, videoID)
		if lastErr == nil {
			break
		}
		if !isRetryableValidationError(lastErr) || attempt == maxAttempts || ctx.Err() != nil {
            uc.logger.LogError(constants.LogLevelError, "Live stream validation failed", videoID, correlationID, lastErr, map[string]interface{}{
				"attempt":     attempt,
				"maxAttempts": maxAttempts,
			})
			return nil, fmt.Errorf("live stream validation failed: %w", lastErr)
		}
        uc.logger.LogStructured(constants.LogLevelWarning, "validation", "retry", "Retrying live stream validation", videoID, correlationID, map[string]interface{}{
			"attempt": attempt,
			"backoff": backoff.String(),
		})
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > 8*time.Second {
			backoff = 8 * time.Second
		}
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	// 既存セッションが同一 VideoID なら再利用（多重起動防止）
	if uc.currentSession != nil {
		if uc.currentSession.VideoID == videoID {
			uc.logger.LogStructured(constants.LogLevelInfo, "monitoring", "reuse_existing", "Reusing existing monitoring session", videoID, correlationID, nil)
			return uc.currentSession, nil
		}
		// 別動画であれば既存を停止して差し替え
		uc.logger.LogStructured(constants.LogLevelInfo, "monitoring", "stopping_previous", "Stopping previous monitoring session", uc.currentSession.VideoID, correlationID, nil)
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

    uc.logger.LogStructured(constants.LogLevelInfo, "monitoring", "session_started", "Started new monitoring session", videoID, correlationID, map[string]interface{}{
        "maxUsers": maxUsers,
    })

	// 既定で自動終了検知を有効化
	uc.pollingService.SetAutoEndEnabled(videoID, true)

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
	// Cancel the context to stop background operations. Do not close the channel here;
	// the sender or context-driven shutdown will ensure goroutines exit safely.
	uc.currentSession.Cancel()
	uc.currentSession = nil

	uc.logger.LogStructured(constants.LogLevelInfo, "monitoring", "session_stopped", "Stopped monitoring session", videoID, correlationID, nil)
	return nil
}

// ResumeMonitoring 停止後に最後の動画で監視を再開します（または指定最大ユーザー数で再開）
func (uc *ChatMonitoringUseCase) ResumeMonitoring(ctx context.Context, maxUsers int) (*MonitoringSession, error) {
	uc.mu.RLock()
	active := uc.currentSession != nil
	last := uc.lastVideoID
	uc.mu.RUnlock()

	if active {
		// 既に監視中ならそのまま成功扱い
		return uc.currentSession, nil
	}
	if last == "" {
		return nil, fmt.Errorf("no previous video to resume")
	}
	if maxUsers <= 0 {
		maxUsers = constants.DefaultMaxUsers
	}
	// StartMonitoring は videoInput として ID も受け付ける
	return uc.StartMonitoring(ctx, last, maxUsers)
}

// GetMonitoringSession 動画の現在の監視セッションを返します
func (uc *ChatMonitoringUseCase) GetMonitoringSession(videoID string) (*MonitoringSession, bool) {
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

// GetUserListOrdered 指定順序でユーザーリストを返します
func (uc *ChatMonitoringUseCase) GetUserListOrdered(ctx context.Context, videoID string, order string) ([]*entity.User, error) {
	return uc.userService.GetUserListSnapshotWithOrder(ctx, videoID, order)
}

// SetAutoEndEnabled 自動終了検知のON/OFFを切り替える（アクティブセッション対象）
func (uc *ChatMonitoringUseCase) SetAutoEndEnabled(enabled bool) (string, error) {
	uc.mu.RLock()
	session := uc.currentSession
	uc.mu.RUnlock()
	if session == nil {
		return "", fmt.Errorf("no active monitoring session")
	}
	uc.pollingService.SetAutoEndEnabled(session.VideoID, enabled)
	return session.VideoID, nil
}

// IsAutoEndEnabled 現在の自動終了検知状態を返す
func (uc *ChatMonitoringUseCase) IsAutoEndEnabled() (string, bool, error) {
	uc.mu.RLock()
	session := uc.currentSession
	uc.mu.RUnlock()
	if session == nil {
		return "", false, fmt.Errorf("no active monitoring session")
	}
	return session.VideoID, uc.pollingService.IsAutoEndEnabled(session.VideoID), nil
}

// runPolling 動画のポーリングループを処理します
func (uc *ChatMonitoringUseCase) runPolling(ctx context.Context, liveChatID, videoID string, messagesChan chan<- entity.ChatMessage, correlationID string) {
	if err := uc.pollingService.StartPolling(ctx, liveChatID, videoID, messagesChan); err != nil {
		// コンテキストキャンセル/タイムアウトは正常停止として扱う
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
            uc.logger.LogStructured(constants.LogLevelInfo, "monitoring", "polling_stopped", "Polling stopped normally", videoID, correlationID, map[string]interface{}{
                "reason": err.Error(),
            })
            return
        }
        if strings.Contains(strings.ToLower(err.Error()), "no longer active") {
            uc.logger.LogStructured(constants.LogLevelInfo, "monitoring", "polling_auto_end", "Polling stopped (stream ended)", videoID, correlationID, nil)
            return
        }
        uc.logger.LogError(constants.LogLevelError, "Polling ended with error", videoID, correlationID, err, nil)
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
                uc.logger.LogError(constants.LogLevelError, "Failed to process chat message", videoID, correlationID, err, map[string]interface{}{
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

// isRetryableValidationError 判定: 一時的なエラーなら true
func isRetryableValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// 非リトライキーワード
	nonRetry := []string{"quotaexceeded", "quota exceeded", "dailylimitexceeded", "daily limit", "invalid", "forbidden", "not live", "video is not a live"}
	for _, k := range nonRetry {
		if strings.Contains(msg, k) {
			return false
		}
	}
	// ネットワーク/内部/timeout系はリトライ許容
	retryHints := []string{"timeout", "internal", "backend", "temporarily", "try again"}
	for _, k := range retryHints {
		if strings.Contains(msg, k) {
			return true
		}
	}
	// デフォルトは非リトライとする（過剰リトライ回避）
	return false
}
