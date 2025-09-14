package youtube

import (
	"context"
	"errors"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	// "google.golang.org/api/option"
	// "google.golang.org/api/youtube/v3"
)

type API struct {
	APIKey string
}

func New(apiKey string) *API { return &API{APIKey: apiKey} }

func (a *API) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	if a.APIKey == "" {
		return "", errors.New("youtube api key is required")
	}

	// YouTube Data API v3を使用してライブストリーム詳細を取得
	// 実際のAPIキーがない場合のモック実装
	// 本番では google.golang.org/api/youtube/v3 を使用
	if videoID == "" {
		return "", errors.New("video ID is required")
	}

	// モック実装：実際の本番環境では以下のようにAPI呼び出しを行う
	// service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	// if err != nil { return "", err }
	// call := service.Videos.List([]string{"liveStreamingDetails"}).Id(videoID)
	// response, err := call.Do()
	// if err != nil { return "", err }
	// return response.Items[0].LiveStreamingDetails.ActiveLiveChatId, nil

	// テスト用のモック値を返す
	return "live:" + videoID + ":chat", nil
}

func (a *API) ListLiveChatMessages(ctx context.Context, liveChatID string) (items []port.ChatMessage, isEnded bool, err error) {
	if a.APIKey == "" {
		return nil, false, errors.New("youtube api key is required")
	}

	if liveChatID == "" {
		return nil, false, errors.New("live chat ID is required")
	}

	// YouTube Data API v3を使用してライブチャットメッセージを取得
	// 実際のAPIキーがない場合のモック実装
	// 本番では google.golang.org/api/youtube/v3 を使用

	// モック実装：実際の本番環境では以下のようにAPI呼び出しを行う
	// service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	// if err != nil { return nil, false, err }
	// call := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"})
	// response, err := call.Do()
	// if err != nil {
	//     // 403エラーやliveChatDisabled等で配信終了を検知
	//     return nil, true, nil
	// }
	//
	// var messages []port.ChatMessage
	// for _, item := range response.Items {
	//     messages = append(messages, port.ChatMessage{
	//         ChannelID:   item.AuthorDetails.ChannelId,
	//         DisplayName: item.AuthorDetails.DisplayName,
	//     })
	// }
	// return messages, false, nil

	// テスト用のモック値を返す
	mockMessages := []port.ChatMessage{
		{ChannelID: "UCtest1", DisplayName: "TestUser1"},
		{ChannelID: "UCtest2", DisplayName: "TestUser2"},
	}

	return mockMessages, false, nil
}
