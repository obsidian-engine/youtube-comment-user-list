package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLiveState_NewReservedState(t *testing.T) {
	scheduled := time.Date(2026, 7, 1, 18, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)

	s := NewReservedState("vid-1", "chat-1", scheduled, now)

	if s.Status != StatusReserved {
		t.Errorf("Status = %v, want %v", s.Status, StatusReserved)
	}
	if s.VideoID != "vid-1" || s.LiveChatID != "chat-1" {
		t.Errorf("videoId/liveChatId not set: %+v", s)
	}
	if !s.AutonomousMonitoring {
		t.Error("AutonomousMonitoring should be true for reserved state")
	}
	if !s.ReservedAt.Equal(now) || !s.ScheduledStartTime.Equal(scheduled) {
		t.Errorf("time fields wrong: %+v", s)
	}
}

func TestLiveState_NewWaitingState(t *testing.T) {
	s := NewWaitingState()

	if s.Status != StatusWaiting {
		t.Errorf("Status = %v, want %v", s.Status, StatusWaiting)
	}
	if s.VideoID != "" || s.LiveChatID != "" {
		t.Errorf("videoId/liveChatId should be empty: %+v", s)
	}
	if s.AutonomousMonitoring {
		t.Error("AutonomousMonitoring should be false")
	}
}

func TestLiveState_Activate(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)

	t.Run("new_videoId_uses_now_as_started_at", func(t *testing.T) {
		prev := LiveState{Status: StatusReserved, VideoID: "prev", AutonomousMonitoring: true}
		got := prev.Activate("new-vid", "new-chat", now)

		if got.Status != StatusActive {
			t.Errorf("Status = %v, want ACTIVE", got.Status)
		}
		if got.VideoID != "new-vid" {
			t.Errorf("VideoID = %q, want new-vid", got.VideoID)
		}
		if !got.StartedAt.Equal(now) {
			t.Errorf("StartedAt = %v, want %v", got.StartedAt, now)
		}
		if !got.AutonomousMonitoring {
			t.Error("AutonomousMonitoring should carry over from prev (Reserve 経由)")
		}
	})

	t.Run("same_videoId_keeps_prev_started_at", func(t *testing.T) {
		prevStart := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
		prev := LiveState{Status: StatusActive, VideoID: "same", StartedAt: prevStart}
		got := prev.Activate("same", "chat", now)

		if !got.StartedAt.Equal(prevStart) {
			t.Errorf("StartedAt = %v, want prev %v (same videoId)", got.StartedAt, prevStart)
		}
	})

	t.Run("Activate_clears_NextPageToken", func(t *testing.T) {
		prev := LiveState{Status: StatusReserved, VideoID: "x", NextPageToken: "old-token"}
		got := prev.Activate("y", "chat", now)
		if got.NextPageToken != "" {
			t.Errorf("NextPageToken = %q, want empty", got.NextPageToken)
		}
	})
}

func TestLiveState_MarkEnded(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	prevStart := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	prev := LiveState{Status: StatusActive, VideoID: "vid", StartedAt: prevStart, AutonomousMonitoring: true}

	got := prev.MarkEnded("vid", "chat", now)

	if got.Status != StatusWaiting {
		t.Errorf("Status = %v, want WAITING", got.Status)
	}
	if !got.StartedAt.Equal(prevStart) {
		t.Errorf("StartedAt = %v, want prev %v", got.StartedAt, prevStart)
	}
	if !got.EndedAt.Equal(now) {
		t.Errorf("EndedAt = %v, want %v", got.EndedAt, now)
	}
	if got.AutonomousMonitoring {
		t.Error("AutonomousMonitoring should be false on MarkEnded")
	}
}

func TestLiveState_MarkEnded_ZeroStartedAtFallsBackToNow(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	prev := LiveState{Status: StatusWaiting, VideoID: "vid"} // StartedAt zero

	got := prev.MarkEnded("vid", "chat", now)

	if !got.StartedAt.Equal(now) {
		t.Errorf("StartedAt = %v, want now %v when prev is zero", got.StartedAt, now)
	}
}

func TestLiveState_CanReserve(t *testing.T) {
	tests := map[string]struct {
		status  Status
		wantErr bool
	}{
		"waiting_ok":      {StatusWaiting, false},
		"reserved_ok":     {StatusReserved, false},
		"active_conflict": {StatusActive, true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := LiveState{Status: tt.status}
			err := s.CanReserve()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUser_LatestCommentedAt(t *testing.T) {
	t.Run("User構造体にLatestCommentedAtフィールドが存在する", func(t *testing.T) {
		now := time.Now()
		user := User{
			ChannelID:         "UC123",
			DisplayName:       "TestUser",
			JoinedAt:          now,
			CommentCount:      5,
			FirstCommentedAt:  now.Add(-time.Hour),
			LatestCommentedAt: now.Add(-time.Minute), // 最新コメント時間
		}

		if user.LatestCommentedAt.IsZero() {
			t.Error("LatestCommentedAtが初期化されていません")
		}

		if user.LatestCommentedAt.Before(user.FirstCommentedAt) {
			t.Error("LatestCommentedAtはFirstCommentedAtより新しい時間であるべきです")
		}
	})

	t.Run("LatestCommentedAtのJSONシリアライゼーションが正しく動作する", func(t *testing.T) {
		now := time.Now()
		user := User{
			ChannelID:         "UC123",
			DisplayName:       "TestUser",
			JoinedAt:          now,
			CommentCount:      5,
			FirstCommentedAt:  now.Add(-time.Hour),
			LatestCommentedAt: now.Add(-time.Minute),
		}

		// JSON にシリアライズ
		jsonData, err := json.Marshal(user)
		if err != nil {
			t.Fatalf("JSON Marshal failed: %v", err)
		}

		// latestCommentedAt フィールドが含まれることを確認
		jsonStr := string(jsonData)
		if !strings.Contains(jsonStr, "latestCommentedAt") {
			t.Error("JSON should contain latestCommentedAt field")
		}

		// JSON からデシリアライズ
		var deserializedUser User
		err = json.Unmarshal(jsonData, &deserializedUser)
		if err != nil {
			t.Fatalf("JSON Unmarshal failed: %v", err)
		}

		// デシリアライズしたデータが元のデータと一致することを確認
		if !deserializedUser.LatestCommentedAt.Equal(user.LatestCommentedAt) {
			t.Errorf("LatestCommentedAt not preserved: got %v, want %v",
				deserializedUser.LatestCommentedAt, user.LatestCommentedAt)
		}
	})
}
