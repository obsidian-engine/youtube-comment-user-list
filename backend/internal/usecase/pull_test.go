package usecase_test

import (
    "context"
    "testing"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

type fakeYTForPull struct{
    items []port.ChatMessage
    ended bool
}

func (f *fakeYTForPull) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
    return "live:abc", nil
}
func (f *fakeYTForPull) ListLiveChatMessages(ctx context.Context, liveChatID string) ([]port.ChatMessage, bool, error) {
    return f.items, f.ended, nil
}

func TestPull_AddsUsers_NormalFlow(t *testing.T) {
    ctx := context.Background()
    users := memory.NewUserRepo()
    state := memory.NewStateRepo()
    _ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
    yt := &fakeYTForPull{items: []port.ChatMessage{{ChannelID: "ch1", DisplayName: "Alice"}}, ended: false}

    uc := &usecase.Pull{YT: yt, Users: users, State: state}
    _, err := uc.Execute(ctx)
    if err == nil {
        t.Fatalf("expected not implemented error (Red phase), got nil")
    }
}

func TestPull_Ended_AutoReset(t *testing.T) {
    ctx := context.Background()
    users := memory.NewUserRepo()
    state := memory.NewStateRepo()
    _ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
    yt := &fakeYTForPull{items: nil, ended: true}

    uc := &usecase.Pull{YT: yt, Users: users, State: state}
    _, err := uc.Execute(ctx)
    if err == nil {
        t.Fatalf("expected not implemented error (Red phase), got nil")
    }
}

