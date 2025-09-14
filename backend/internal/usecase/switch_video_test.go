package usecase_test

import (
    "context"
    "testing"
    "time"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

type fakeYT struct{}

func (f *fakeYT) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
    return "live:abc", nil
}
func (f *fakeYT) ListLiveChatMessages(ctx context.Context, liveChatID string) ([]port.ChatMessage, bool, error) {
    return nil, false, nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func TestSwitchVideo_UsersClearedAndStateActive(t *testing.T) {
    ctx := context.Background()

    users := memory.NewUserRepo()
    _ = users.Upsert("ch1", "to-be-cleared")
    state := memory.NewStateRepo()
    yt := &fakeYT{}

    uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: fixedClock{t: time.Unix(1000,0)}}

    _, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video123"})
    if err == nil {
        t.Fatalf("expected not implemented error (Red phase), got nil")
    }

    // 将来的に: Users がクリアされ、State が ACTIVE、VideoID/liveChatId/StartedAt が設定されることを検証
    // この時点では Red を維持するため、厳密検証は実装後に追加予定。
    _ = domain.StatusActive
}

