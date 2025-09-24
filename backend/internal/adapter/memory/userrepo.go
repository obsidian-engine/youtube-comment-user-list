package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

type UserRepo struct {
	mu            sync.RWMutex
	usersByID     map[string]domain.User // channelID -> User with join time
	processedMsgs map[string]bool        // messageID -> processed flag for deduplication
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		usersByID:     make(map[string]domain.User),
		processedMsgs: make(map[string]bool),
	}
}



func (r *UserRepo) Count() int {
	r.mu.RLock()
	c := len(r.usersByID)
	r.mu.RUnlock()
	return c
}

func (r *UserRepo) Clear() {
	r.mu.Lock()
	r.usersByID = make(map[string]domain.User)
	r.processedMsgs = make(map[string]bool)
	r.mu.Unlock()
}

// UpsertWithJoinTime は channelID をキーに displayName と初回参加時間を登録/更新します。
// 既に存在するユーザーの場合、joinedAt は更新されません。
func (r *UserRepo) UpsertWithJoinTime(channelID string, displayName string, joinedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 既存ユーザーの場合は参加時間を保持、発言数をインクリメント
	if existingUser, exists := r.usersByID[channelID]; exists {
		existingUser.DisplayName = displayName     // 表示名は更新
		existingUser.CommentCount++                // 発言数をインクリメント
		existingUser.LatestCommentedAt = joinedAt  // 最新コメント時間を更新
		r.usersByID[channelID] = existingUser
	} else {
		// 新規ユーザーの場合
		r.usersByID[channelID] = domain.User{
			ChannelID:         channelID,
			DisplayName:       displayName,
			JoinedAt:          joinedAt,
			CommentCount:      1,        // 初回コメントなので1
			FirstCommentedAt:  joinedAt, // 初回コメント時刻
			LatestCommentedAt: joinedAt, // 最新コメント時刻（初回なので同じ）
		}
	}

	return nil
}

// UpsertWithMessageUpdated adds a user with join time and message ID for deduplication
// Returns true if the user data was actually updated (not a duplicate message)
func (r *UserRepo) UpsertWithMessageUpdated(channelID string, displayName string, joinedAt time.Time, messageID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// メッセージIDの重複チェック
	if r.processedMsgs[messageID] {
		// 既に処理済みのメッセージの場合、何もしない
		return false, nil
	}

	// 既存ユーザーの場合は参加時間を保持、発言数をインクリメント
	if existingUser, exists := r.usersByID[channelID]; exists {
		existingUser.DisplayName = displayName     // 表示名は更新
		existingUser.CommentCount++                // 発言数をインクリメント
		existingUser.LatestCommentedAt = joinedAt  // 最新コメント時間を更新
		r.usersByID[channelID] = existingUser
	} else {
		// 新規ユーザーの場合
		r.usersByID[channelID] = domain.User{
			ChannelID:         channelID,
			DisplayName:       displayName,
			JoinedAt:          joinedAt,
			CommentCount:      1,        // 初回コメントなので1
			FirstCommentedAt:  joinedAt, // 初回コメント時刻
			LatestCommentedAt: joinedAt, // 最新コメント時刻（初回なので同じ）
		}
	}

	// メッセージIDを処理済みとして記録
	r.processedMsgs[messageID] = true

	return true, nil
}

// UpsertWithMessage adds a user with join time and message ID for deduplication (backward compatibility)
func (r *UserRepo) UpsertWithMessage(channelID string, displayName string, joinedAt time.Time, messageID string) error {
	_, err := r.UpsertWithMessageUpdated(channelID, displayName, joinedAt, messageID)
	return err
}

// ListUsersSortedByJoinTime は User構造体の配列を参加時間順（早い順）で返します。
func (r *UserRepo) ListUsersSortedByJoinTime() []domain.User {
	r.mu.RLock()
	users := make([]domain.User, 0, len(r.usersByID))
	for _, user := range r.usersByID {
		users = append(users, user)
	}
	r.mu.RUnlock()

	// 参加時間順でソート（早い順）
	sort.Slice(users, func(i, j int) bool {
		return users[i].JoinedAt.Before(users[j].JoinedAt)
	})

	return users
}
