package youtube

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type API struct {
	APIKey string
}

func New(apiKey string) *API { return &API{APIKey: apiKey} }

func (a *API) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	log.Printf("[YOUTUBE_API] GetActiveLiveChatID called with videoID: %s", videoID)

	if a.APIKey == "" {
		log.Printf("[YOUTUBE_API] Error: API key is empty")
		return "", errors.New("youtube api key is required")
	}

	if videoID == "" {
		log.Printf("[YOUTUBE_API] Error: videoID is empty")
		return "", errors.New("video ID is required")
	}

	// YouTube Data API v3を使用して動画情報を取得
	service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to create YouTube service: %v", err)
		return "", err
	}

	// 動画情報を取得してライブチャットIDを取得
	call := service.Videos.List([]string{"liveStreamingDetails"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to get video details: %v", err)
		return "", err
	}

	if len(response.Items) == 0 {
		log.Printf("[YOUTUBE_API] Video not found: %s", videoID)
		return "", errors.New("video not found")
	}

	video := response.Items[0]
	if video.LiveStreamingDetails == nil {
		log.Printf("[YOUTUBE_API] Video is not a live stream: %s", videoID)
		return "", errors.New("video is not a live stream")
	}

	if video.LiveStreamingDetails.ActiveLiveChatId == "" {
		log.Printf("[YOUTUBE_API] Live chat is not active for video: %s", videoID)
		return "", errors.New("live chat is not active")
	}

	liveChatID := video.LiveStreamingDetails.ActiveLiveChatId
	log.Printf("[YOUTUBE_API] Found active liveChatID: %s", liveChatID)
	return liveChatID, nil
}

func (a *API) ListLiveChatMessages(ctx context.Context, liveChatID string) (items []port.ChatMessage, isEnded bool, err error) {
	log.Printf("[YOUTUBE_API] ListLiveChatMessages called with liveChatID: %s", liveChatID)

	if a.APIKey == "" {
		log.Printf("[YOUTUBE_API] Error: API key is empty")
		return nil, false, errors.New("youtube api key is required")
	}

	if liveChatID == "" {
		log.Printf("[YOUTUBE_API] Error: liveChatID is empty")
		return nil, false, errors.New("live chat ID is required")
	}

	// YouTube Data API v3を使用してライブチャットメッセージを取得
	service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to create YouTube service: %v", err)
		return nil, false, err
	}

	// ライブチャットメッセージを取得 - 正しいAPIの使い方
	call := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"})
	response, err := call.Do()
	if err != nil {
		log.Printf("[YOUTUBE_API] API call failed: %v", err)

		// より詳細な配信終了検知条件
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "forbidden") ||
			strings.Contains(errMsg, "livechatdisabled") ||
			strings.Contains(errMsg, "livechatended") ||
			strings.Contains(errMsg, "livechatnotfound") ||
			strings.Contains(errMsg, "chatdisabled") ||
			strings.Contains(errMsg, "livechatnotactive") ||
			(strings.Contains(errMsg, "notfound") && strings.Contains(errMsg, "livechat")) {
			log.Printf("[YOUTUBE_API] Live chat ended or disabled - Error: %s", errMsg)
			return nil, true, nil
		}

		return nil, false, err
	}

	// レスポンスからメッセージを変換
	var messages []port.ChatMessage
	for _, item := range response.Items {
		if item.AuthorDetails != nil && item.Snippet != nil {
			// publishedAtを解析
			publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			if err != nil {
				log.Printf("[YOUTUBE_API] Failed to parse publishedAt: %v", err)
				// エラーの場合は現在時刻を使用
				publishedAt = time.Now()
			}

			messages = append(messages, port.ChatMessage{
				ChannelID:   item.AuthorDetails.ChannelId,
				DisplayName: item.AuthorDetails.DisplayName,
				PublishedAt: publishedAt,
			})
		}
	}

	log.Printf("[YOUTUBE_API] Successfully retrieved %d messages", len(messages))
	for i, msg := range messages {
		log.Printf("[YOUTUBE_API] Message %d: ChannelID=%s, DisplayName=%s, PublishedAt=%s", 
			i+1, msg.ChannelID, msg.DisplayName, msg.PublishedAt.Format(time.RFC3339))
	}

	return messages, false, nil
}
