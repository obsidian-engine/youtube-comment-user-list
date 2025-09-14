package usecase

import (
    "context"

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
}

// Execute: コメント取得・ユーザー追加、終了検知→WAITING へ（autoReset）。
func (uc *Pull) Execute(ctx context.Context) (PullOutput, error) {
    return PullOutput{}, ErrNotImplemented
}
