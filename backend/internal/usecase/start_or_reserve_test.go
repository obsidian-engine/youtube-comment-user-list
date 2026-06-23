package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

func TestStartOrReserve_DispatchesToReserve_WhenNotStarted(t *testing.T) {
	ctx := context.Background()
	state := memory.NewStateRepo()
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	scheduled := now.Add(6 * time.Hour)
	yt := &fakeYTForReserve{
		details: port.VideoLiveDetails{
			IsLiveContent:      true,
			ScheduledStartTime: scheduled,
		},
	}
	clock := fixedClock{t: now}
	uc := &usecase.StartOrReserve{
		YT:          yt,
		Clock:       clock,
		SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: memory.NewUserRepo(), State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
		Reserve:     &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
	}

	out, err := uc.Execute(ctx, usecase.StartOrReserveInput{VideoID: "vid-reserve"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Dispatched != "reserve" {
		t.Errorf("Dispatched = %q, want reserve", out.Dispatched)
	}
	if out.State.Status != domain.StatusReserved {
		t.Errorf("Status = %v, want RESERVED", out.State.Status)
	}
}

func TestStartOrReserve_DispatchesToSwitch_WhenStarted(t *testing.T) {
	ctx := context.Background()
	state := memory.NewStateRepo()
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	yt := &fakeYTForReserve{
		details: port.VideoLiveDetails{
			LiveChatID:         "chat123",
			IsLiveContent:      true,
			ActualStartTime:    now.Add(-10 * time.Minute),
			ScheduledStartTime: now.Add(-15 * time.Minute),
		},
	}
	clock := fixedClock{t: now}
	uc := &usecase.StartOrReserve{
		YT:          yt,
		Clock:       clock,
		SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: memory.NewUserRepo(), State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
		Reserve:     &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
	}

	out, err := uc.Execute(ctx, usecase.StartOrReserveInput{VideoID: "vid-live"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out.Dispatched != "switch" {
		t.Errorf("Dispatched = %q, want switch", out.Dispatched)
	}
	if out.State.Status != domain.StatusActive {
		t.Errorf("Status = %v, want ACTIVE", out.State.Status)
	}
}

func TestIsLiveNotStarted(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		details port.VideoLiveDetails
		want    bool
	}{
		{
			name: "actualStartTime zero",
			details: port.VideoLiveDetails{
				IsLiveContent: true,
			},
			want: true,
		},
		{
			name: "scheduledStartTime in future",
			details: port.VideoLiveDetails{
				IsLiveContent:      true,
				ActualStartTime:    now.Add(-1 * time.Minute),
				ScheduledStartTime: now.Add(1 * time.Hour),
			},
			want: true,
		},
		{
			name: "started",
			details: port.VideoLiveDetails{
				IsLiveContent:      true,
				ActualStartTime:    now.Add(-10 * time.Minute),
				ScheduledStartTime: now.Add(-15 * time.Minute),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usecase.IsLiveNotStarted(tt.details, now)
			if got != tt.want {
				t.Errorf("IsLiveNotStarted() = %v, want %v", got, tt.want)
			}
		})
	}
}
