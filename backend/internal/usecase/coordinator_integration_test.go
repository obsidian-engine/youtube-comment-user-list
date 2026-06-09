package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

// mockCoord は Coordinator の呼出し回数・引数を記録するテスト用 mock。
type mockCoord struct {
	markDirty int
	flush     int
	setVideo  [][2]string
	calls     []string // 呼出し順を記録
}

func (m *mockCoord) Restore(_ context.Context) error { return nil }
func (m *mockCoord) SetVideo(videoID, liveChatID string) {
	m.setVideo = append(m.setVideo, [2]string{videoID, liveChatID})
	m.calls = append(m.calls, "SetVideo")
}
func (m *mockCoord) MarkDirty() {
	m.markDirty++
	m.calls = append(m.calls, "MarkDirty")
}
func (m *mockCoord) Flush(_ context.Context) error {
	m.flush++
	m.calls = append(m.calls, "Flush")
	return nil
}
func (m *mockCoord) Start(_ context.Context) {}
func (m *mockCoord) Stop()                   {}
func (m *mockCoord) RestoredAt() (time.Time, time.Time, bool) {
	return time.Time{}, time.Time{}, false
}

// TestSwitchVideo_CoordinatorCallOrder: 旧 video Flush → SetVideo → MarkDirty → Flush の順序を検証
func TestSwitchVideo_CoordinatorCallOrder(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := &fakeYT{}
	clock := fixedClock{t: time.Unix(2000, 0)}
	coord := &mockCoord{}

	uc := &usecase.SwitchVideo{
		YT:    yt,
		Users: users,
		State: state,
		Clock: clock,
		Snap:  coord,
	}

	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "new-video"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 期待呼出し順: Flush(旧) → SetVideo → MarkDirty → Flush(新)
	want := []string{"Flush", "SetVideo", "MarkDirty", "Flush"}
	if len(coord.calls) != len(want) {
		t.Fatalf("calls = %v, want %v", coord.calls, want)
	}
	for i, call := range coord.calls {
		if call != want[i] {
			t.Errorf("calls[%d] = %q, want %q", i, call, want[i])
		}
	}

	// SetVideo の引数確認
	if len(coord.setVideo) != 1 {
		t.Fatalf("SetVideo called %d times, want 1", len(coord.setVideo))
	}
	if coord.setVideo[0][0] != "new-video" {
		t.Errorf("SetVideo videoID = %q, want %q", coord.setVideo[0][0], "new-video")
	}
	if coord.setVideo[0][1] != "live:abc" {
		t.Errorf("SetVideo liveChatID = %q, want %q", coord.setVideo[0][1], "live:abc")
	}

	// Flush は 2 回呼ばれる
	if coord.flush != 2 {
		t.Errorf("flush count = %d, want 2", coord.flush)
	}
}

// TestPull_MarkDirty_CalledOnlyWhenDiff: 差分あり時のみ MarkDirty が呼ばれることを検証
func TestPull_MarkDirty_CalledOnlyWhenDiff(t *testing.T) {
	t.Helper()

	tests := []struct {
		name              string
		items             []port.ChatMessage
		wantMarkDirtyCall int
	}{
		{
			name: "差分あり: MarkDirty 呼ばれる",
			items: []port.ChatMessage{
				{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
			},
			wantMarkDirtyCall: 1,
		},
		{
			name:              "差分なし (items 空): MarkDirty 呼ばれない",
			items:             []port.ChatMessage{},
			wantMarkDirtyCall: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			ctx := context.Background()
			users := memory.NewUserRepo()
			comments := memory.NewCommentRepo()
			state := memory.NewStateRepo()
			_ = state.Set(ctx, domain.LiveState{
				Status:     domain.StatusActive,
				VideoID:    "v",
				LiveChatID: "live:chat",
			})
			clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
			coord := &mockCoord{}

			yt := &fakeYTForPull{items: tt.items, ended: false}
			uc := &usecase.Pull{
				YT:       yt,
				Users:    users,
				Comments: comments,
				State:    state,
				Clock:    clock,
				Snap:     coord,
			}

			_, err := uc.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if coord.markDirty != tt.wantMarkDirtyCall {
				t.Errorf("markDirty calls = %d, want %d", coord.markDirty, tt.wantMarkDirtyCall)
			}
		})
	}
}
