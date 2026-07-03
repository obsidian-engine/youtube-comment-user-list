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

// NewReservedState は videoId を予約状態にした新しい LiveState を返します。
// AutonomousMonitoring は true (サーバー側 pull を予約経由で自動有効化)。
func NewReservedState(videoID, liveChatID string, scheduledStart, now time.Time) LiveState {
	return LiveState{
		Status:               StatusReserved,
		VideoID:              videoID,
		LiveChatID:           liveChatID,
		AutonomousMonitoring: true,
		ReservedAt:           now,
		ScheduledStartTime:   scheduledStart,
	}
}

// NewWaitingState は videoId 未設定の待機状態を返します (Reset / CancelReserve 共通)。
func NewWaitingState() LiveState {
	return LiveState{Status: StatusWaiting}
}

// Activate は前状態から ACTIVE 遷移した LiveState を返します。
// 同 videoId 再切替の場合は前 StartedAt を維持、新規は now を採用します。
// AutonomousMonitoring は Reserve 経由の予約フラグを引き継ぎます。
func (s LiveState) Activate(videoID, liveChatID string, now time.Time) LiveState {
	startedAt := now
	if s.VideoID == videoID && !s.StartedAt.IsZero() {
		startedAt = s.StartedAt
	}
	return LiveState{
		Status:               StatusActive,
		VideoID:              videoID,
		LiveChatID:           liveChatID,
		StartedAt:            startedAt,
		AutonomousMonitoring: s.AutonomousMonitoring,
	}
}

// MarkEnded は前状態を「終了配信を再表示するための WAITING」として返します。
// 直近 StartedAt を維持しつつ EndedAt を now でセット、AutonomousMonitoring は明示的に false に落とします。
func (s LiveState) MarkEnded(videoID, liveChatID string, now time.Time) LiveState {
	startedAt := s.StartedAt
	if startedAt.IsZero() {
		startedAt = now
	}
	return LiveState{
		Status:               StatusWaiting,
		VideoID:              videoID,
		LiveChatID:           liveChatID,
		StartedAt:            startedAt,
		EndedAt:              now,
		AutonomousMonitoring: false,
	}
}

// CanReserve は Reserve/CancelReserve が受理可能かを判定します。
// ACTIVE 中の遷移は state 破壊につながるため *APIError (Conflict) を返します。
func (s LiveState) CanReserve() error {
	if s.Status == StatusActive {
		return &APIError{Code: ErrCodeConflict, Message: "stream is currently active, reset first"}
	}
	return nil
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
