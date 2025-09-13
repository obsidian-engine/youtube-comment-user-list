// Package service アプリケーション層のサービスを定義します
package service

import (
    "context"
    "fmt"
    "sort"
    "strings"
    "unicode"

    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
    "github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
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

	userList, err := us.userRepo.GetUserList(ctx, message.VideoID)
    if err != nil {
        us.logger.LogError(constants.LogLevelError, "Failed to get user list", message.VideoID, correlationID, err, nil)
        return fmt.Errorf("failed to get user list: %w", err)
    }

	existedBefore := userList.HasUser(message.AuthorDetails.ChannelID)
	added := userList.UpsertFromMessage(message)

	// 追加 or 更新 の判定
	if added {
		u := userList.Users[message.AuthorDetails.ChannelID]
        us.logger.LogUser(constants.LogLevelInfo, "New user added", message.VideoID, correlationID, map[string]interface{}{
			"channelId":    u.ChannelID,
			"displayName":  u.DisplayName,
			"userCount":    userList.Count(),
			"messageCount": u.MessageCount,
			"isFull":       userList.IsFull(),
		})

		if err := us.userRepo.UpdateUserList(ctx, message.VideoID, userList); err != nil {
            us.logger.LogError(constants.LogLevelError, "Failed to update user list (add)", message.VideoID, correlationID, err, nil)
            return fmt.Errorf("failed to update user list: %w", err)
        }

		if err := us.eventPub.PublishUserAdded(ctx, *u, message.VideoID); err != nil {
            us.logger.LogError(constants.LogLevelError, "Failed to publish user added event", message.VideoID, correlationID, err, nil)
        }
		return nil
	}

	// 既存更新 or リスト満杯
	if existedBefore {
		u := userList.Users[message.AuthorDetails.ChannelID]
		// 更新ログ（DEBUG）
        us.logger.LogUser(constants.LogLevelDebug, "User updated", message.VideoID, correlationID, map[string]interface{}{
			"channelId":    u.ChannelID,
			"displayName":  u.DisplayName,
			"messageCount": u.MessageCount,
			"lastSeen":     u.LastSeen,
		})
		// 保存（メモリなので参照更新だが一貫性のため）
		if err := us.userRepo.UpdateUserList(ctx, message.VideoID, userList); err != nil {
            us.logger.LogError(constants.LogLevelError, "Failed to persist updated user list", message.VideoID, correlationID, err, nil)
            return fmt.Errorf("failed to update user list: %w", err)
        }
    } else {
        // 追加できず（満杯）
        us.logger.LogUser(constants.LogLevelWarning, "User list full - user skipped", message.VideoID, correlationID, map[string]interface{}{
            "channelId":   message.AuthorDetails.ChannelID,
            "displayName": message.AuthorDetails.DisplayName,
            "userCount":   userList.Count(),
            "maxUsers":    userList.MaxUsers,
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

    us.logger.LogUser(constants.LogLevelInfo, "Creating new user list", videoID, correlationID, map[string]interface{}{
        "operation": "create_user_list",
        "maxUsers":  maxUsers,
    })

	userList := entity.NewUserList(maxUsers)

    if err := us.userRepo.UpdateUserList(ctx, videoID, userList); err != nil {
        us.logger.LogError(constants.LogLevelError, "Failed to create user list", videoID, correlationID, err, nil)
        return nil, fmt.Errorf("failed to create user list: %w", err)
    }

    us.logger.LogUser(constants.LogLevelInfo, "User list created successfully", videoID, correlationID, map[string]interface{}{
        "operation": "create_user_list",
        "maxUsers":  maxUsers,
    })

	return userList, nil
}

// GetUserListSnapshot ユーザーリストのスナップショットを取得します
func (us *UserService) GetUserListSnapshot(ctx context.Context, videoID string) ([]*entity.User, error) {
	// 互換: 既定の並び順は参加順（first_seen）
	return us.GetUserListSnapshotWithOrder(ctx, videoID, "first_seen")
}

// GetUserListSnapshotWithOrder フィルタ適用後に指定した順序でスナップショットを取得します
// 既定で以下を除外します: チャットオーナー, モデレーター, ボット(Nightbot等)
// order: first_seen | kana | message_count
func (us *UserService) GetUserListSnapshotWithOrder(ctx context.Context, videoID string, order string) ([]*entity.User, error) {
	userList, err := us.GetUserList(ctx, videoID)
	if err != nil {
		return nil, err
	}

	raw := userList.GetUsers()
	// 既定フィルタ適用
	filtered := make([]*entity.User, 0, len(raw))
	for _, u := range raw {
		if shouldExcludeUser(u) {
			continue
		}
		filtered = append(filtered, u)
	}

	// 並び替え
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "kana":
		sort.SliceStable(filtered, func(i, j int) bool {
			return lessJapanese(filtered[i].DisplayName, filtered[j].DisplayName)
		})
	case "message_count":
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].MessageCount == filtered[j].MessageCount {
				return lessJapanese(filtered[i].DisplayName, filtered[j].DisplayName)
			}
			return filtered[i].MessageCount > filtered[j].MessageCount
		})
	default: // first_seen
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].FirstSeen.Equal(filtered[j].FirstSeen) {
				return lessJapanese(filtered[i].DisplayName, filtered[j].DisplayName)
			}
			return filtered[i].FirstSeen.Before(filtered[j].FirstSeen)
		})
	}

	return filtered, nil
}

// lessJapanese は日本語の簡易的な「あいうえお」順になるよう正規化して比較します
// 仕様:
// - カタカナをひらがなに変換
// - 英数字は小文字化
// - 前後の空白を削除
// この簡易実装は一般的なケースで「あ→い→う…」の順序感を満たすことを目的とし、
// 厳密な辞書順や漢字の読みには対応しません。
func lessJapanese(a, b string) bool {
	na := normalizeJapanese(a)
	nb := normalizeJapanese(b)
	if na == nb {
		return a < b
	}
	return na < nb
}

func normalizeJapanese(s string) string {
	s = strings.TrimSpace(s)
	var buf []rune
	for _, r := range s {
		// アルファベットは小文字化
		r = unicode.ToLower(r)
		// カタカナ → ひらがな (U+30A1–U+30F3 → U+3041–U+3093)
		if r >= 0x30A1 && r <= 0x30F3 {
			r = r - 0x60
		}
		// 全角カタカナの長音記号(ー)などはそのまま
		buf = append(buf, r)
	}
	return string(buf)
}

// shouldExcludeUser 既定の除外条件を適用します
// - チャットオーナー
// - モデレーター
// - 既知/推定ボット
func shouldExcludeUser(u *entity.User) bool {
	if u == nil {
		return true
	}
	if u.IsChatOwner || u.IsModerator {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(u.DisplayName))
	if name == "" {
		return false
	}
	if isBotDisplayName(name) {
		return true
	}
	return false
}

func isBotDisplayName(name string) bool {
	// 代表例: Nightbot, StreamElements, Streamlabs 等
	if strings.Contains(name, "bot") {
		return true
	}
	known := []string{"nightbot", "streamelements", "streamlabs"}
	for _, k := range known {
		if strings.Contains(name, k) {
			return true
		}
	}
	return false
}
