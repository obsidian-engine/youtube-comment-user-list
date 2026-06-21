package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

// fakeYTForReserve は GetVideoLiveDetails の戻り値を制御できる fake。
type fakeYTForReserve struct {
	details port.VideoLiveDetails
	err     error
}

func (f *fakeYTForReserve) GetActiveLiveChatID(_ context.Context, _ string) (port.VideoMeta, error) {
	return port.VideoMeta{}, nil
}
func (f *fakeYTForReserve) ListLiveChatMessages(_ context.Context, _ string, _ string) ([]port.ChatMessage, string, int64, int, bool, error) {
	return nil, "", 0, 0, false, nil
}
func (f *fakeYTForReserve) GetChannelDisplayNames(_ context.Context, _ []string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeYTForReserve) GetChannelHandles(_ context.Context, _ []string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeYTForReserve) GetVideoLiveDetails(_ context.Context, _ string) (port.VideoLiveDetails, error) {
	return f.details, f.err
}

func TestReserve_WaitingAndIsLive_SetsReserved(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	scheduled := time.Date(2026, 7, 1, 18, 0, 0, 0, time.UTC)
	yt := &fakeYTForReserve{
		details: port.VideoLiveDetails{
			LiveChatID:         "",
			IsLiveContent:      true,
			ScheduledStartTime: scheduled,
		},
	}
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}

	uc := &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	out, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "vid-reserved"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if out.State.Status != domain.StatusReserved {
		t.Errorf("Status = %v, want %v", out.State.Status, domain.StatusReserved)
	}
	if out.State.VideoID != "vid-reserved" {
		t.Errorf("VideoID = %v, want vid-reserved", out.State.VideoID)
	}
	if !out.State.AutonomousMonitoring {
		t.Error("AutonomousMonitoring = false, want true")
	}
	if !out.State.ScheduledStartTime.Equal(scheduled) {
		t.Errorf("ScheduledStartTime = %v, want %v", out.State.ScheduledStartTime, scheduled)
	}
	if !out.State.ReservedAt.Equal(now) {
		t.Errorf("ReservedAt = %v, want %v", out.State.ReservedAt, now)
	}

	// state に保存されているか確認
	persisted, _ := state.Get(ctx)
	if persisted.Status != domain.StatusReserved {
		t.Errorf("persisted Status = %v, want %v", persisted.Status, domain.StatusReserved)
	}
}

func TestReserve_ActiveStatus_ReturnsConflict(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "current"})

	yt := &fakeYTForReserve{details: port.VideoLiveDetails{IsLiveContent: true}}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "vid-new"})
	if err == nil {
		t.Fatal("Execute should return error on ACTIVE status")
	}
	var apiErr *domain.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *domain.APIError", err)
	}
	if apiErr.Code != domain.ErrCodeConflict {
		t.Errorf("Code = %v, want %v", apiErr.Code, domain.ErrCodeConflict)
	}
}

func TestReserve_EmptyVideoID_ReturnsInvalidArgument(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	yt := &fakeYTForReserve{}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: ""})
	if err == nil {
		t.Fatal("Execute should return error on empty videoId")
	}
	var apiErr *domain.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *domain.APIError", err)
	}
	if apiErr.Code != domain.ErrCodeInvalidArgument {
		t.Errorf("Code = %v, want %v", apiErr.Code, domain.ErrCodeInvalidArgument)
	}
}

func TestReserve_NotLiveContent_ReturnsInvalidArgument(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	// IsLiveContent = false: 通常動画に対する予約
	yt := &fakeYTForReserve{details: port.VideoLiveDetails{IsLiveContent: false}}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "vid-normal"})
	if err == nil {
		t.Fatal("Execute should return error on non-live video")
	}
	var apiErr *domain.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *domain.APIError", err)
	}
	if apiErr.Code != domain.ErrCodeInvalidArgument {
		t.Errorf("Code = %v, want %v", apiErr.Code, domain.ErrCodeInvalidArgument)
	}
}

func TestReserve_YTAPIError_ReturnsWrappedError(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	yt := &fakeYTForReserve{err: errors.New("youtube api quota exceeded")}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.Reserve{YT: yt, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "vid-x"})
	if err == nil {
		t.Fatal("Execute should return error on YT API failure")
	}
	// *domain.APIError ではなく wrapped error として返る
	var apiErr *domain.APIError
	if errors.As(err, &apiErr) {
		t.Errorf("should not be *domain.APIError, got code=%v", apiErr.Code)
	}
}
