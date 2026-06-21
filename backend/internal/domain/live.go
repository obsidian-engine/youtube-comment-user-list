package domain

import "time"

// Status は配信状態を表します。
type Status string

const (
	StatusWaiting  Status = "WAITING"
	StatusActive   Status = "ACTIVE"
	StatusReserved Status = "RESERVED"
)

// LiveState は現在の配信に関する状態を保持します。
type LiveState struct {
	Status               Status
	VideoID              string
	LiveChatID           string
	StartedAt            time.Time
	EndedAt              time.Time
	LastPulledAt         time.Time
	NextPageToken        string
	AutonomousMonitoring bool      // 予約経由 ACTIVE 中はサーバー側で pull する
	ReservedAt           time.Time // 予約を受け付けた時刻
	ScheduledStartTime   time.Time // YouTube が返す配信予定開始時刻
}

// User represents a user with join time information
type User struct {
	ChannelID         string    `json:"channelId"`
	DisplayName       string    `json:"displayName"`
	JoinedAt          time.Time `json:"joinedAt"`
	CommentCount      int       `json:"commentCount"`
	FirstCommentedAt  time.Time `json:"firstCommentedAt"`
	LatestCommentedAt time.Time `json:"latestCommentedAt"`
}
