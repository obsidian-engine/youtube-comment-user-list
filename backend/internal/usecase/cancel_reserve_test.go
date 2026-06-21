package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

func TestCancelReserve_Reserved_SetsWaiting(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:               domain.StatusReserved,
		VideoID:              "vid-reserved",
		AutonomousMonitoring: true,
	})

	uc := &usecase.CancelReserve{State: state, Snap: &snapshot.NopCoordinator{}}

	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if out.State.Status != domain.StatusWaiting {
		t.Errorf("Status = %v, want %v", out.State.Status, domain.StatusWaiting)
	}
	// 予約 field がゼロになっているか確認
	if out.State.VideoID != "" {
		t.Errorf("VideoID = %v, want empty", out.State.VideoID)
	}
	if out.State.AutonomousMonitoring {
		t.Error("AutonomousMonitoring = true, want false")
	}

	persisted, _ := state.Get(ctx)
	if persisted.Status != domain.StatusWaiting {
		t.Errorf("persisted Status = %v, want %v", persisted.Status, domain.StatusWaiting)
	}
}

func TestCancelReserve_Waiting_IsIdempotent(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	// 既に WAITING — 冪等性確認
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusWaiting})

	uc := &usecase.CancelReserve{State: state, Snap: &snapshot.NopCoordinator{}}

	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed on WAITING state: %v", err)
	}
	if out.State.Status != domain.StatusWaiting {
		t.Errorf("Status = %v, want %v", out.State.Status, domain.StatusWaiting)
	}
}

// TestCancelReserve_Active_ReturnsConflict: ACTIVE 中のキャンセルは 409 Conflict。
// Reserve とシンメトリックに現セッションのデータ破壊を防ぐ。
// ACTIVE 停止は /reset エンドポイントが担う設計。
func TestCancelReserve_Active_ReturnsConflict(t *testing.T) {
	ctx := context.Background()

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "live-now"})

	uc := &usecase.CancelReserve{State: state, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx)
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
