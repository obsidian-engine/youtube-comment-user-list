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

func TestReserve_DifferentVideoID_ClearsPreviousUsersAndComments(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	// 旧配信 (WAITING or ENDED 等 ACTIVE 以外) の state を残した状態から開始する
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusWaiting, VideoID: "prev-vid"})

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch-prev-1", "prev user", time.Now())
	_ = users.UpsertWithJoinTime("ch-prev-2", "prev user 2", time.Now())
	comments := memory.NewCommentRepo()
	_ = comments.Add(domain.Comment{ID: "msg-prev", ChannelID: "ch-prev-1", DisplayName: "prev user", Message: "prev"})

	yt := &fakeYTForReserve{
		details: port.VideoLiveDetails{
			IsLiveContent:      true,
			ScheduledStartTime: time.Date(2026, 7, 1, 18, 0, 0, 0, time.UTC),
		},
	}
	clock := fixedClock{t: time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)}

	uc := &usecase.Reserve{
		YT: yt, Users: users, Comments: comments,
		State: state, Clock: clock, Snap: &snapshot.NopCoordinator{},
	}

	if _, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "new-vid"}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if users.Count() != 0 {
		t.Errorf("users.Count = %d, want 0 (should be cleared for different videoId)", users.Count())
	}
	if comments.Count() != 0 {
		t.Errorf("comments.Count = %d, want 0 (should be cleared for different videoId)", comments.Count())
	}
}

func TestReserve_SameVideoID_KeepsUsersAndComments(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	// 同 videoId の再予約 (例: RESERVED 中に予約時刻を更新するケース)
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusReserved, VideoID: "same-vid"})

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch-1", "user", time.Now())
	comments := memory.NewCommentRepo()
	_ = comments.Add(domain.Comment{ID: "msg-1", ChannelID: "ch-1", DisplayName: "user", Message: "hello"})

	yt := &fakeYTForReserve{details: port.VideoLiveDetails{IsLiveContent: true}}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.Reserve{
		YT: yt, Users: users, Comments: comments,
		State: state, Clock: clock, Snap: &snapshot.NopCoordinator{},
	}

	if _, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "same-vid"}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if users.Count() != 1 {
		t.Errorf("users.Count = %d, want 1 (should keep for same videoId)", users.Count())
	}
	if comments.Count() != 1 {
		t.Errorf("comments.Count = %d, want 1 (should keep for same videoId)", comments.Count())
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
