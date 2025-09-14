package usecase

import (
    "context"
    "time"

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
    // 未実装: ここでは署名のみ（スケルトン）
    _ = time.Now // 参照保持で未使用警告回避（Clock を使う実装想定）
    return SwitchVideoOutput{}, ErrNotImplemented
}
