package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// UserRepository インメモリストレージを使用してUserRepositoryインターフェースを実装します
type UserRepository struct {
	mu        sync.RWMutex
	userLists map[string]*entity.UserList // 動画ID -> ユーザーリスト
}

// NewUserRepository 新しいインメモリユーザーリポジトリを作成します
func NewUserRepository() *UserRepository {
	return &UserRepository{
		userLists: make(map[string]*entity.UserList),
	}
}

// GetUserList 動画のユーザーリストを取得します
func (r *UserRepository) GetUserList(ctx context.Context, videoID string) (*entity.UserList, error) {
	_ = ctx
	r.mu.RLock()
	userList, exists := r.userLists[videoID]
	r.mu.RUnlock()

	if !exists {
		// 書き込みロックで再確認（ダブルチェック）
		r.mu.Lock()
		defer r.mu.Unlock()
		if userList, exists = r.userLists[videoID]; !exists {
			defaultMaxUsers := constants.DefaultMaxUsers
			userList = entity.NewUserList(defaultMaxUsers)
			r.userLists[videoID] = userList
		}
	}

	return userList, nil
}

// UpdateUserList 動画のユーザーリストを更新します
func (r *UserRepository) UpdateUserList(ctx context.Context, videoID string, userList *entity.UserList) error {
	_ = ctx
	if userList == nil {
		return fmt.Errorf("userList cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.userLists[videoID] = userList
	return nil
}
