package http

import (
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// TypeSafeStatusResponse は型安全なステータス応答構造体
// interface{}型を具体的な*time.Time型に置き換え
type TypeSafeStatusResponse struct {
	Status       string     `json:"status"`
	Count        int        `json:"count"`
	VideoID      string     `json:"videoId"`
	LiveChatID   string     `json:"liveChatId"`
	StartedAt    *time.Time `json:"startedAt"`    // interface{} → *time.Time
	EndedAt      *time.Time `json:"endedAt"`      // interface{} → *time.Time
	LastPulledAt *time.Time `json:"lastPulledAt"` // interface{} → *time.Time
}

// TypeSafeSwitchVideoResponse は型安全な動画切り替え応答構造体
type TypeSafeSwitchVideoResponse struct {
	Status     string     `json:"status"`
	VideoID    string     `json:"videoId"`
	LiveChatID string     `json:"liveChatId"`
	StartedAt  *time.Time `json:"startedAt"` // interface{} → *time.Time
}

// TypeSafePullResponse はプル応答構造体（既に型安全）
type TypeSafePullResponse struct {
	AddedCount            int   `json:"addedCount"`
	AutoReset             bool  `json:"autoReset"`
	PollingIntervalMillis int64 `json:"pollingIntervalMillis"`
}

// TypeSafeResetResponse はリセット応答構造体（既に型安全）
type TypeSafeResetResponse struct {
	Status string `json:"status"`
}

// ファクトリー関数: 既存のStatusResponseから型安全版を作成
func NewTypeSafeStatusResponseFromLegacy(legacy StatusResponse) TypeSafeStatusResponse {
	var startedAt, endedAt, lastPulledAt *time.Time

	// interface{}から*time.Timeへの安全な変換
	if legacy.StartedAt != nil {
		if t, ok := legacy.StartedAt.(time.Time); ok {
			startedAt = &t
		}
	}

	if legacy.EndedAt != nil {
		if t, ok := legacy.EndedAt.(time.Time); ok {
			endedAt = &t
		}
	}

	if legacy.LastPulledAt != nil {
		if t, ok := legacy.LastPulledAt.(time.Time); ok {
			lastPulledAt = &t
		}
	}

	return TypeSafeStatusResponse{
		Status:       legacy.Status,
		Count:        legacy.Count,
		VideoID:      legacy.VideoID,
		LiveChatID:   legacy.LiveChatID,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		LastPulledAt: lastPulledAt,
	}
}

// ファクトリー関数: 既存のSwitchVideoResponseから型安全版を作成
func NewTypeSafeSwitchVideoResponseFromLegacy(legacy SwitchVideoResponse) TypeSafeSwitchVideoResponse {
	var startedAt *time.Time

	if legacy.StartedAt != nil {
		if t, ok := legacy.StartedAt.(time.Time); ok {
			startedAt = &t
		}
	}

	return TypeSafeSwitchVideoResponse{
		Status:     legacy.Status,
		VideoID:    legacy.VideoID,
		LiveChatID: legacy.LiveChatID,
		StartedAt:  startedAt,
	}
}

// HistorySummaryResponse は /history/snapshots の 1 件分のサマリーレスポンスです。
type HistorySummaryResponse struct {
	VideoID      string `json:"videoId"`
	SavedAt      string `json:"savedAt"` // ISO8601
	UserCount    int    `json:"userCount"`
	CommentCount int    `json:"commentCount"`
	VideoTitle   string `json:"videoTitle,omitempty"`
	ChannelTitle string `json:"channelTitle,omitempty"`
}

// HistoryListResponse は /history/snapshots のレスポンスです。
type HistoryListResponse struct {
	Items []HistorySummaryResponse `json:"items"`
	Logs  []LogDetail              `json:"logs,omitempty"`
}

// newHistorySummaryResponse は port.SnapshotSummary から HistorySummaryResponse を生成します。
func newHistorySummaryResponse(s port.SnapshotSummary) HistorySummaryResponse {
	return HistorySummaryResponse{
		VideoID:      s.VideoID,
		SavedAt:      s.SavedAt.UTC().Format(time.RFC3339),
		UserCount:    s.UserCount,
		CommentCount: s.CommentCount,
		VideoTitle:   s.VideoTitle,
		ChannelTitle: s.ChannelTitle,
	}
}

// HistorySnapshotResponse は /history/snapshots/{videoID} のレスポンスです。
// port.Snapshot の JSON shape に合わせています。
type HistorySnapshotResponse struct {
	VideoID      string            `json:"videoId"`
	SavedAt      string            `json:"savedAt"` // ISO8601
	VideoTitle   string            `json:"videoTitle,omitempty"`
	ChannelTitle string            `json:"channelTitle,omitempty"`
	Users        []domain.User     `json:"users"`
	Comments     []domain.Comment  `json:"comments"`
	State        *domain.LiveState `json:"state,omitempty"`
	Logs         []LogDetail       `json:"logs,omitempty"`
}

// newHistorySnapshotResponse は port.Snapshot から HistorySnapshotResponse を生成します。
func newHistorySnapshotResponse(snap *port.Snapshot) HistorySnapshotResponse {
	users := snap.Users
	if users == nil {
		users = []domain.User{}
	}
	comments := snap.Comments
	if comments == nil {
		comments = []domain.Comment{}
	}
	return HistorySnapshotResponse{
		VideoID:      snap.VideoID,
		SavedAt:      snap.SavedAt.UTC().Format(time.RFC3339),
		VideoTitle:   snap.VideoTitle,
		ChannelTitle: snap.ChannelTitle,
		Users:        users,
		Comments:     comments,
		State:        snap.State,
	}
}
