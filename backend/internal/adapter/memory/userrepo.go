package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

type UserRepo struct {
	mu         sync.RWMutex
	byChan     map[string]string    // channelID -> displayName (backward compatibility)
	usersByID  map[string]domain.User // channelID -> User with join time
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		byChan:    make(map[string]string),
		usersByID: make(map[string]domain.User),
	}
}

func (r *UserRepo) Upsert(channelID string, displayName string) error {
	r.mu.Lock()
	r.byChan[channelID] = displayName
	r.mu.Unlock()
	return nil
}

func (r *UserRepo) ListDisplayNames() []string {
	r.mu.RLock()
	names := make([]string, 0, len(r.byChan))
	for _, n := range r.byChan {
		names = append(names, n)
	}
	r.mu.RUnlock()
	return names
}

func (r *UserRepo) Count() int {
	r.mu.RLock()
	c := len(r.byChan)
	r.mu.RUnlock()
	return c
}

func (r *UserRepo) Clear() {
	r.mu.Lock()
	r.byChan = make(map[string]string)
	r.usersByID = make(map[string]domain.User)
	r.mu.Unlock()
}

// UpsertWithJoinTime は channelID をキーに displayName と初回参加時間を登録/更新します。
// 既に存在するユーザーの場合、joinedAt は更新されません。
func (r *UserRepo) UpsertWithJoinTime(channelID string, displayName string, joinedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 後方互換性のためにbyChannも更新
	r.byChan[channelID] = displayName

	// 既存ユーザーの場合は参加時間を保持
	if existingUser, exists := r.usersByID[channelID]; exists {
		existingUser.DisplayName = displayName // 表示名は更新
		r.usersByID[channelID] = existingUser
	} else {
		// 新規ユーザーの場合
		r.usersByID[channelID] = domain.User{
			ChannelID:   channelID,
			DisplayName: displayName,
			JoinedAt:    joinedAt,
		}
	}

	return nil
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
