package usecase

import (
	"context"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type StatusOutput struct {
	Status       domain.Status
	Count        int
	VideoID      string
	StartedAt    time.Time
	EndedAt      time.Time
	LastPulledAt time.Time
}

type Status struct {
	Users port.UserRepo
	State port.StateRepo
}

func (uc *Status) Execute(ctx context.Context) (StatusOutput, error) {
	// 現在の状態を取得
	state, err := uc.State.Get(ctx)
	if err != nil {
		return StatusOutput{}, err
	}

	// ユーザー数を取得
	count := uc.Users.Count()

	return StatusOutput{
		Status:       state.Status,
		Count:        count,
		VideoID:      state.VideoID,
		StartedAt:    state.StartedAt,
		EndedAt:      state.EndedAt,
		LastPulledAt: state.LastPulledAt,
	}, nil
}
