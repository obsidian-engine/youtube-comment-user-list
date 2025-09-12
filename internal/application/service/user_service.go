// Package service アプリケーション層のサービスを定義します
package service

import (
	"context"
	"fmt"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// UserService チャットユーザーを管理するサービスです
type UserService struct {
	userRepo repository.UserRepository
	logger   repository.Logger
	eventPub repository.EventPublisher
}

// NewUserService 新しいUserServiceを作成します
func NewUserService(
	userRepo repository.UserRepository,
	logger repository.Logger,
	eventPub repository.EventPublisher,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
		eventPub: eventPub,
	}
}

// ProcessChatMessage チャットメッセージを処理してユーザーを管理します
func (us *UserService) ProcessChatMessage(ctx context.Context, message entity.ChatMessage) error {
	correlationID := fmt.Sprintf("user-%s-%s", message.VideoID, message.ID)

	// 動画の現在のユーザーリストを取得
	userList, err := us.userRepo.GetUserList(ctx, message.VideoID)
	if err != nil {
		us.logger.LogError("ERROR", "Failed to get user list", message.VideoID, correlationID, err, nil)
		return fmt.Errorf("failed to get user list: %w", err)
	}

	// チャットメッセージからユーザーを作成
	user := entity.NewUserFromChatMessage(message)

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
	return us.userRepo.GetUserList(ctx, videoID)
}

// CreateUserList 動画用の新しいユーザーリストを作成します
func (us *UserService) CreateUserList(ctx context.Context, videoID string, maxUsers int) (*entity.UserList, error) {
	correlationID := fmt.Sprintf("create-userlist-%s", videoID)

	us.logger.LogUser("INFO", "Creating new user list", videoID, correlationID, map[string]interface{}{
		"operation": "create_user_list",
		"maxUsers":  maxUsers,
	})

	userList := entity.NewUserList(maxUsers)

	if err := us.userRepo.UpdateUserList(ctx, videoID, userList); err != nil {
		us.logger.LogError("ERROR", "Failed to create user list", videoID, correlationID, err, nil)
		return nil, fmt.Errorf("failed to create user list: %w", err)
	}

	us.logger.LogUser("INFO", "User list created successfully", videoID, correlationID, map[string]interface{}{
		"operation": "create_user_list",
		"maxUsers":  maxUsers,
	})

	return userList, nil
}

// GetUserListSnapshot ユーザーリストのスナップショットを取得します
func (us *UserService) GetUserListSnapshot(ctx context.Context, videoID string) ([]*entity.User, error) {
	userList, err := us.GetUserList(ctx, videoID)
	if err != nil {
		return nil, err
	}
	return userList.GetUsers(), nil
}
