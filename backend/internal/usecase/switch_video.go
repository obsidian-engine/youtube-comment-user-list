package usecase

import (
    "context"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type SwitchVideoInput struct {
    VideoID string
}

type SwitchVideoOutput struct {
    State domain.LiveState
}

type SwitchVideo struct {
    YT    port.YouTubePort
    Users port.UserRepo
    State port.StateRepo
    Clock port.Clock
}

// Execute: videoId 切替、ユーザー初期化、State=ACTIVE に遷移。
func (uc *SwitchVideo) Execute(ctx context.Context, in SwitchVideoInput) (SwitchVideoOutput, error) {
	// 1. YouTube APIでliveChatIDを取得
	liveChatID, err := uc.YT.GetActiveLiveChatID(ctx, in.VideoID)
	if err != nil {
		return SwitchVideoOutput{}, err
	}

	// 2. ユーザーをクリア
	uc.Users.Clear()

	// 3. StateをACTIVEに更新
	now := uc.Clock.Now()
	newState := domain.LiveState{
		Status:     domain.StatusActive,
		VideoID:    in.VideoID,
		LiveChatID: liveChatID,
		StartedAt:  now,
	}
	
	if err := uc.State.Set(ctx, newState); err != nil {
		return SwitchVideoOutput{}, err
	}

	return SwitchVideoOutput{State: newState}, nil
}
