package youtube

import (
    "context"
    "errors"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type API struct {
    APIKey string
}

func New(apiKey string) *API { return &API{APIKey: apiKey} }

func (a *API) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
    return "", errors.New("youtube api: not implemented")
}

func (a *API) ListLiveChatMessages(ctx context.Context, liveChatID string) (items []port.ChatMessage, isEnded bool, err error) {
    return nil, false, errors.New("youtube api: not implemented")
}
