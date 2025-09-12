package entity

import (
	"sync"
	"time"
)

// User はチャットユーザーを表します
type User struct {
	ChannelID   string    `json:"channel_id"`
	DisplayName string    `json:"display_name"`
	FirstSeen   time.Time `json:"first_seen"`
}

// NewUserFromChatMessage チャットメッセージからユーザーを作成します
func NewUserFromChatMessage(message ChatMessage) User {
	return User{
		ChannelID:   message.AuthorDetails.ChannelID,
		DisplayName: message.AuthorDetails.DisplayName,
		FirstSeen:   message.Timestamp,
	}
}

// UserList は並行安全性を持つユーザーのコレクションを管理します
type UserList struct {
	Users    map[string]*User // channelID -> User
	MaxUsers int
	// 内部同期用
	mu sync.RWMutex
}

// NewUserList 指定された最大サイズで新しいUserListを作成します
func NewUserList(maxUsers int) *UserList {
	return &UserList{
		Users:    make(map[string]*User),
		MaxUsers: maxUsers,
	}
}

// AddUser ユーザーがまだ存在せず制限内の場合、新しいユーザーをリストに追加します
func (ul *UserList) AddUser(channelID, displayName string) bool {
	ul.mu.Lock()
	defer ul.mu.Unlock()
	if _, exists := ul.Users[channelID]; exists {
		return false // 既に存在
	}
	if len(ul.Users) >= ul.MaxUsers {
		return false // 制限超過
	}
	ul.Users[channelID] = &User{
		ChannelID:   channelID,
		DisplayName: displayName,
		FirstSeen:   time.Now(),
	}
	return true
}

// HasUser 指定したチャンネルIDのユーザーが存在するかを返します
func (ul *UserList) HasUser(channelID string) bool {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	_, exists := ul.Users[channelID]
	return exists
}

// GetUsers 全ユーザーのスナップショットを返します
func (ul *UserList) GetUsers() []*User {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	users := make([]*User, 0, len(ul.Users))
	for _, user := range ul.Users {
		users = append(users, user)
	}
	return users
}

// Count 現在のユーザー数を返します
func (ul *UserList) Count() int {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	return len(ul.Users)
}

// IsFull ユーザーリストが最大容量に達した場合trueを返します
func (ul *UserList) IsFull() bool {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	return len(ul.Users) >= ul.MaxUsers
}
