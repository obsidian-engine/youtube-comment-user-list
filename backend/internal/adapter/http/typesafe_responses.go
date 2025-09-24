package http

import "time"

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