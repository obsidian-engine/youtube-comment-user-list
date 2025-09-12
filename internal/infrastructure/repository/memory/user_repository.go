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

// DeleteUserList removes the user list for a video (useful for cleanup)
func (r *UserRepository) DeleteUserList(ctx context.Context, videoID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.userLists, videoID)
	return nil
}

// GetAllVideoIDs returns all video IDs that have user lists
func (r *UserRepository) GetAllVideoIDs(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	videoIDs := make([]string, 0, len(r.userLists))
	for videoID := range r.userLists {
		videoIDs = append(videoIDs, videoID)
	}

	return videoIDs, nil
}

// GetUserListStats returns statistics about user lists
func (r *UserRepository) GetUserListStats(ctx context.Context) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]interface{}{
		"totalVideoIds": len(r.userLists),
		"totalUsers":    0,
		"videoStats":    make(map[string]interface{}),
	}

	totalUsers := 0
	videoStats := make(map[string]interface{})

	for videoID, userList := range r.userLists {
		userCount := userList.Count()
		totalUsers += userCount

		videoStats[videoID] = map[string]interface{}{
			"userCount": userCount,
			"maxUsers":  userList.MaxUsers,
			"isFull":    userList.IsFull(),
		}
	}

	stats["totalUsers"] = totalUsers
	stats["videoStats"] = videoStats

	return stats, nil
}
