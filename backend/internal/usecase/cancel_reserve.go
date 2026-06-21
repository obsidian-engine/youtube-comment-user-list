package usecase

import (
	"context"
	"fmt"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type CancelReserveOutput struct {
	State domain.LiveState
}

type CancelReserve struct {
	State port.StateRepo
	Snap  snapshot.Coordinator
}

// Execute: RESERVED/WAITING を WAITING に正規化 (冪等)。
// ACTIVE 中は 409 Conflict で拒否 — Reserve とシンメトリックに現セッションを守る。
// users/comments は触らない (Reserve では users 未変更のため)。
func (uc *CancelReserve) Execute(ctx context.Context) (CancelReserveOutput, error) {
	cur, err := uc.State.Get(ctx)
	if err != nil {
		return CancelReserveOutput{}, fmt.Errorf("state_get: %w", err)
	}
	// ACTIVE 中のキャンセルは state 破壊につながるため拒否する
	if cur.Status == domain.StatusActive {
		return CancelReserveOutput{}, &domain.APIError{Code: domain.ErrCodeConflict, Message: "stream is currently active, reset first"}
	}

	newState := domain.LiveState{Status: domain.StatusWaiting}
	if err := uc.State.Set(ctx, newState); err != nil {
		return CancelReserveOutput{}, fmt.Errorf("state_set: %w", err)
	}

	uc.Snap.MarkDirty()
	if err := uc.Snap.Flush(ctx); err != nil {
		logging.Log(ctx, "warn", "SNAPSHOT", "cancel_reserve: snapshot flush failed: %v", err)
	}

	return CancelReserveOutput{State: newState}, nil
}
