package entity

// User はチャットユーザーを表します
type User struct {
	ChannelID   string
	DisplayName string
}

// UserList は並行安全性を持つユーザーのコレクションを管理します
type UserList struct {
	Users    map[string]*User // channelID -> User
	MaxUsers int
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
	if _, exists := ul.Users[channelID]; exists {
		return false // ユーザーは既に存在します
	}

	if len(ul.Users) >= ul.MaxUsers {
		return false // 制限に達しました
	}

	ul.Users[channelID] = &User{
		ChannelID:   channelID,
		DisplayName: displayName,
	}
	return true
}

// GetUsers 全ユーザーのスナップショットを返します
func (ul *UserList) GetUsers() []*User {
	users := make([]*User, 0, len(ul.Users))
	for _, user := range ul.Users {
		users = append(users, user)
	}
	return users
}

// Count 現在のユーザー数を返します
func (ul *UserList) Count() int {
	return len(ul.Users)
}

// IsFull ユーザーリストが最大容量に達した場合trueを返します
func (ul *UserList) IsFull() bool {
	return len(ul.Users) >= ul.MaxUsers
}
