package usecase

import (
	"context"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type PullOutput struct {
	AddedCount int
	AutoReset  bool
}

type Pull struct {
	YT    port.YouTubePort
	Users port.UserRepo
	State port.StateRepo
	Clock port.Clock
}

// Execute: コメント取得・ユーザー追加、終了検知→WAITING へ（autoReset）。
func (uc *Pull) Execute(ctx context.Context) (PullOutput, error) {
	// 現在の状態を取得
	state, err := uc.State.Get(ctx)
	if err != nil {
		return PullOutput{}, err
	}

	// WAITING状態の場合は何もしない
	if state.Status != domain.StatusActive {
		return PullOutput{AddedCount: 0, AutoReset: false}, nil
	}

	// YouTube APIからメッセージを取得
	items, isEnded, err := uc.YT.ListLiveChatMessages(ctx, state.LiveChatID)
	if err != nil {
		return PullOutput{}, err
	}

	// 配信終了検知
	if isEnded {
		// ユーザークリア
		uc.Users.Clear()

		// WAITINGに戻す（現在時刻を終了時刻として設定）
		state.Status = domain.StatusWaiting
		state.EndedAt = uc.Clock.Now()
		if err := uc.State.Set(ctx, state); err != nil {
			return PullOutput{}, err
		}

		return PullOutput{AddedCount: 0, AutoReset: true}, nil
	}

	// ユーザー追加
	addedCount := 0
	now := uc.Clock.Now()
	for _, msg := range items {
		if err := uc.Users.UpsertWithJoinTime(msg.ChannelID, msg.DisplayName, now); err != nil {
			return PullOutput{}, err
		}
		addedCount++
	}

	return PullOutput{AddedCount: addedCount, AutoReset: false}, nil
}
