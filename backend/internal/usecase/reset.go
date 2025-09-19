package usecase

import (
	"context"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type ResetOutput struct {
	State domain.LiveState
}

type Reset struct {
	Users port.UserRepo
	State port.StateRepo
}

// Execute: Users クリア、State=WAITING
func (uc *Reset) Execute(ctx context.Context) (ResetOutput, error) {
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

	return ResetOutput{State: newState}, nil
}
