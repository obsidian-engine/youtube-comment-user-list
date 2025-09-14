package usecase

import (
    "context"
    "time"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type StatusOutput struct {
    Status    domain.Status
    Count     int
    VideoID   string
    StartedAt time.Time
    EndedAt   time.Time
}

type Status struct {
    Users port.UserRepo
    State port.StateRepo
}

func (uc *Status) Execute(ctx context.Context) (StatusOutput, error) {
    return StatusOutput{}, ErrNotImplemented
}
