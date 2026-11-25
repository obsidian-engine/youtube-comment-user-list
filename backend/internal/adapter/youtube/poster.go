package youtube

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Poster は bot アカウントのログイン許可 (OAuth の refresh token) で
// liveChatMessages.insert を呼ぶ port.ChatPoster 実装。1 回の投稿で API 使用量を 20 unit 使う。
type Poster struct {
	tokens oauth2.TokenSource
}

func NewPoster(clientID, clientSecret, refreshToken string) *Poster {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeForceSslScope},
	}
	return &Poster{tokens: conf.TokenSource(context.Background(), &oauth2.Token{RefreshToken: refreshToken})}
}

func (p *Poster) PostMessage(ctx context.Context, liveChatID, text string) error {
	service, err := youtube.NewService(ctx, option.WithTokenSource(p.tokens))
	if err != nil {
		return fmt.Errorf("youtube: new service for posting: %w", err)
	}
	msg := &youtube.LiveChatMessage{
		Snippet: &youtube.LiveChatMessageSnippet{
			LiveChatId:         liveChatID,
			Type:               "textMessageEvent",
			TextMessageDetails: &youtube.LiveChatTextMessageDetails{MessageText: text},
		},
	}
	if _, err := service.LiveChatMessages.Insert([]string{"snippet"}, msg).Context(ctx).Do(); err != nil {
		return fmt.Errorf("youtube: insert live chat message: %w", err)
	}
	return nil
}
