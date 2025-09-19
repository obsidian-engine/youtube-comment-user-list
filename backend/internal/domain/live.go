package domain

import "time"

// Status は配信状態を表します。
type Status string

const (
	StatusWaiting Status = "WAITING"
	StatusActive  Status = "ACTIVE"
)

// LiveState は現在の配信に関する状態を保持します。
type LiveState struct {
	Status        Status
	VideoID       string
	LiveChatID    string
	NextPageToken string
	StartedAt     time.Time
	EndedAt       time.Time
	LastPulledAt  time.Time
}

// User represents a user with join time information
type User struct {
	ChannelID         string    `json:"channelId"`
	DisplayName       string    `json:"displayName"`
	JoinedAt          time.Time `json:"joinedAt"`
	CommentCount      int       `json:"commentCount"`
	FirstCommentedAt  time.Time `json:"firstCommentedAt"`
}
