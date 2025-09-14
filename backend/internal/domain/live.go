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
    Status     Status
    VideoID    string
    LiveChatID string
    StartedAt  time.Time
    EndedAt    time.Time
}
