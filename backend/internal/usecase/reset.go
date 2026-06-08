package usecase

import (
	"context"
	"log"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type ResetOutput struct {
	State domain.LiveState
}

type Reset struct {
	Users port.UserRepo
	State port.StateRepo
	Snap  snapshot.Coordinator
}

// Execute: Users クリア、State=WAITING
func (uc *Reset) Execute(ctx context.Context) (ResetOutput, error) {
	// リセット前の状態を snapshot に保存
	if uc.Snap != nil {
		if err := uc.Snap.Flush(ctx); err != nil {
			log.Printf("[WARN] reset: snapshot flush (pre-reset) failed: %v", err)
		}
	}

	// ユーザーをクリア
	uc.Users.Clear()

	// StateをWAITINGに戻す
	newState := domain.LiveState{
		Status:        domain.StatusWaiting,
		NextPageToken: "",
	}

	if err := uc.State.Set(ctx, newState); err != nil {
		return ResetOutput{}, err
	}

	// video unset 状態にして current.json を更新
	if uc.Snap != nil {
		uc.Snap.SetVideo("", "")
		if err := uc.Snap.Flush(ctx); err != nil {
			log.Printf("[WARN] reset: snapshot flush (clear current) failed: %v", err)
		}
	}

	return ResetOutput{State: newState}, nil
}
