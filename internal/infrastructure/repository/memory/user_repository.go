package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// UserRepository implements the UserRepository interface using in-memory storage
type UserRepository struct {
	mu        sync.RWMutex
	userLists map[string]*entity.UserList // videoID -> UserList
}

// NewUserRepository creates a new in-memory user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		userLists: make(map[string]*entity.UserList),
	}
}

// GetUserList retrieves the user list for a video
func (r *UserRepository) GetUserList(ctx context.Context, videoID string) (*entity.UserList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userList, exists := r.userLists[videoID]
	if !exists {
		// Create a new user list with default max users if it doesn't exist
		defaultMaxUsers := constants.DefaultMaxUsers
		userList = entity.NewUserList(defaultMaxUsers)
		r.userLists[videoID] = userList
	}

	return userList, nil
}

// UpdateUserList updates the user list for a video
func (r *UserRepository) UpdateUserList(ctx context.Context, videoID string, userList *entity.UserList) error {
	if userList == nil {
		return fmt.Errorf("userList cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.userLists[videoID] = userList
	return nil
}
