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
	Users    port.UserRepo
	Comments port.CommentRepo
	State    port.StateRepo
	Snap     snapshot.Coordinator // 必須 (GCS 不要な場合は NopCoordinator を渡す)
}

// Execute: Users クリア、State=WAITING
func (uc *Reset) Execute(ctx context.Context) (ResetOutput, error) {
	// リセット前の状態を snapshot に保存
	if err := uc.Snap.Flush(ctx); err != nil {
		log.Printf("[WARN] reset: snapshot flush (pre-reset) failed: %v", err)
	}

	// ユーザーとコメントをクリア
	uc.Users.Clear()
	if uc.Comments != nil {
		uc.Comments.Clear()
	}

	// StateをWAITINGに戻す
	newState := domain.LiveState{
		Status:        domain.StatusWaiting,
		NextPageToken: "",
	}

	if err := uc.State.Set(ctx, newState); err != nil {
		return ResetOutput{}, err
	}

	// video unset 状態にして current.json を更新
	uc.Snap.SetVideo("", "")
	if err := uc.Snap.Flush(ctx); err != nil {
		log.Printf("[WARN] reset: snapshot flush (clear current) failed: %v", err)
	}

	return ResetOutput{State: newState}, nil
}
