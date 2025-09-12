package service

import (
	"context"
	"fmt"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// UserService ユーザー管理のビジネスロジックを処理します
type UserService struct {
	userRepo UserRepository
	logger   Logger
	eventPub EventPublisher
}

// NewUserService 新しいUserServiceを作成します
func NewUserService(
	userRepo UserRepository,
	logger Logger,
	eventPub EventPublisher,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
		eventPub: eventPub,
	}
}

// ProcessChatMessage チャットメッセージを処理し、必要に応じてユーザーリストを更新します
func (us *UserService) ProcessChatMessage(ctx context.Context, message entity.ChatMessage) error {
	correlationID := fmt.Sprintf("user-%s-%s", message.VideoID, message.ID)

	// 動画の現在のユーザーリストを取得
	userList, err := us.userRepo.GetUserList(ctx, message.VideoID)
	if err != nil {
		us.logger.LogError("ERROR", "Failed to get user list", message.VideoID, correlationID, err, nil)
		return fmt.Errorf("failed to get user list: %w", err)
	}

	// チャットメッセージからユーザーを作成
	user := entity.User{
		ChannelID:   message.AuthorDetails.ChannelID,
		DisplayName: message.AuthorDetails.DisplayName,
	}

	// ユーザーをリストに追加を試行
	wasAdded := userList.AddUser(user.ChannelID, user.DisplayName)

	if wasAdded {
		us.logger.LogUser("INFO", "New user added", message.VideoID, correlationID, map[string]interface{}{
			"channelId":   user.ChannelID,
			"displayName": user.DisplayName,
			"userCount":   userList.Count(),
			"isFull":      userList.IsFull(),
		})

		// リポジトリでユーザーリストを更新
		if err := us.userRepo.UpdateUserList(ctx, message.VideoID, userList); err != nil {
			us.logger.LogError("ERROR", "Failed to update user list", message.VideoID, correlationID, err, nil)
			return fmt.Errorf("failed to update user list: %w", err)
		}

		// ユーザー追加イベントを発行
		if err := us.eventPub.PublishUserAdded(ctx, user, message.VideoID); err != nil {
			us.logger.LogError("ERROR", "Failed to publish user added event", message.VideoID, correlationID, err, nil)
			// コア機能にとって重要ではないため、ここではエラーを返しません
		}
	} else {
		us.logger.LogUser("DEBUG", "User already exists or list is full", message.VideoID, correlationID, map[string]interface{}{
			"channelId":   user.ChannelID,
			"displayName": user.DisplayName,
			"userCount":   userList.Count(),
			"isFull":      userList.IsFull(),
		})
	}

	return nil
}

// GetUserList 動画のユーザーリストを取得します
func (us *UserService) GetUserList(ctx context.Context, videoID string) (*entity.UserList, error) {
	correlationID := fmt.Sprintf("get-users-%s", videoID)

	userList, err := us.userRepo.GetUserList(ctx, videoID)
	if err != nil {
		us.logger.LogError("ERROR", "Failed to get user list", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("failed to get user list: %w", err)
	}

	us.logger.LogUser("DEBUG", "User list retrieved", videoID, correlationID, map[string]interface{}{
		"userCount": userList.Count(),
		"isFull":    userList.IsFull(),
	})

	return userList, nil
}

// CreateUserList 指定された最大ユーザー数で動画用の新しいユーザーリストを作成します
func (us *UserService) CreateUserList(ctx context.Context, videoID string, maxUsers int) (*entity.UserList, error) {
	correlationID := fmt.Sprintf("create-users-%s", videoID)

	// 最大ユーザー数を検証
	if maxUsers <= 0 {
		err := fmt.Errorf("maxUsers must be positive, got: %d", maxUsers)
		us.logger.LogError("ERROR", "Invalid maxUsers parameter", videoID, correlationID, err, map[string]interface{}{
			"maxUsers": maxUsers,
		})
		return nil, err
	}

	userList := entity.NewUserList(maxUsers)

	// リポジトリに保存
	if err := us.userRepo.UpdateUserList(ctx, videoID, userList); err != nil {
		us.logger.LogError("ERROR", "Failed to create user list", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("failed to create user list: %w", err)
	}

	us.logger.LogUser("INFO", "User list created", videoID, correlationID, map[string]interface{}{
		"maxUsers": maxUsers,
	})

	return userList, nil
}

// GetUserListSnapshot 表示目的でユーザーのスナップショットを返します
func (us *UserService) GetUserListSnapshot(ctx context.Context, videoID string) ([]*entity.User, error) {
	userList, err := us.GetUserList(ctx, videoID)
	if err != nil {
		return nil, err
	}

	return userList.GetUsers(), nil
}

// ValidateUser ビジネスルールに対するユーザーデータを検証します
func (us *UserService) ValidateUser(user *entity.User) error {
	if user.ChannelID == "" {
		return fmt.Errorf("channelID cannot be empty")
	}

	if user.DisplayName == "" {
		return fmt.Errorf("displayName cannot be empty")
	}

	// ここに追加のビジネス検証ルールを追加してください
	return nil
}
