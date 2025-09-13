package entity

import (
	"sync"
	"time"
)

// User はチャットユーザーを表します
type User struct {
	ChannelID    string    `json:"channel_id"`
	DisplayName  string    `json:"display_name"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	MessageCount int       `json:"message_count"`
	IsChatOwner  bool      `json:"is_chat_owner"`
	IsModerator  bool      `json:"is_moderator"`
	IsMember     bool      `json:"is_member"`
}

// NewUserFromChatMessage 初回ユーザー生成
func NewUserFromChatMessage(message ChatMessage) *User {
	return &User{
		ChannelID:    message.AuthorDetails.ChannelID,
		DisplayName:  message.AuthorDetails.DisplayName,
		FirstSeen:    message.Timestamp,
		LastSeen:     message.Timestamp,
		MessageCount: 1,
		IsChatOwner:  message.AuthorDetails.IsChatOwner,
		IsModerator:  message.AuthorDetails.IsModerator,
		IsMember:     message.AuthorDetails.IsMember,
	}
}

// UpdateFromMessage 既存ユーザー情報を新しいメッセージで更新
func (u *User) UpdateFromMessage(msg ChatMessage) {
	u.LastSeen = msg.Timestamp
	u.MessageCount++
	// 表示名が変わるケース（絵文字追加等）に備えて上書き
	if msg.AuthorDetails.DisplayName != "" {
		u.DisplayName = msg.AuthorDetails.DisplayName
	}
	// 権限ロールは一度でも true になったら残す
	u.IsChatOwner = u.IsChatOwner || msg.AuthorDetails.IsChatOwner
	u.IsModerator = u.IsModerator || msg.AuthorDetails.IsModerator
	u.IsMember = u.IsMember || msg.AuthorDetails.IsMember
}

// UserList は並行安全性を持つユーザーのコレクションを管理します
type UserList struct {
	Users    map[string]*User // channelID -> User
	MaxUsers int
	mu       sync.RWMutex
}

// NewUserList 指定された最大サイズで新しいUserListを作成します
func NewUserList(maxUsers int) *UserList {
	return &UserList{
		Users:    make(map[string]*User),
		MaxUsers: maxUsers,
	}
}

// UpsertFromMessage メッセージからユーザーを新規作成または更新。戻り値: 新規追加されたか
func (ul *UserList) UpsertFromMessage(msg ChatMessage) bool {
	ul.mu.Lock()
	defer ul.mu.Unlock()
	if existing, ok := ul.Users[msg.AuthorDetails.ChannelID]; ok {
		existing.UpdateFromMessage(msg)
		return false
	}
	if len(ul.Users) >= ul.MaxUsers {
		return false
	}
	ul.Users[msg.AuthorDetails.ChannelID] = NewUserFromChatMessage(msg)
	return true
}

// HasUser 指定したチャンネルIDのユーザーが存在するか
func (ul *UserList) HasUser(channelID string) bool {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	_, exists := ul.Users[channelID]
	return exists
}

// GetUsers 全ユーザーのスナップショット
func (ul *UserList) GetUsers() []*User {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	users := make([]*User, 0, len(ul.Users))
	for _, user := range ul.Users {
		users = append(users, user)
	}
	return users
}

// Count 現在のユーザー数
func (ul *UserList) Count() int {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	return len(ul.Users)
}

// IsFull ユーザーリストが最大容量に達したか
func (ul *UserList) IsFull() bool {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	return len(ul.Users) >= ul.MaxUsers
}
