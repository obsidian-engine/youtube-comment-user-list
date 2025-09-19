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

	var messages []port.ChatMessage
	pageToken := ""
	totalPages := 0

	// ページング処理を実装 - すべてのコメントを取得
	for {
		// ライブチャットメッセージを取得 - 正しいAPIの使い方
		call := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"})

		// 最大件数を設定（YouTube APIの最大値は2000）
		call = call.MaxResults(2000)

		// ページトークンがある場合は設定
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

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
				return messages, true, nil
			}

			return nil, false, err
		}

		totalPages++

		// レスポンスからメッセージを変換
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
					ID:          item.Id, // メッセージIDを追加
					ChannelID:   item.AuthorDetails.ChannelId,
					DisplayName: item.AuthorDetails.DisplayName,
					PublishedAt: publishedAt,
				})
			}
		}

		log.Printf("[YOUTUBE_API] Page %d: Retrieved %d messages", totalPages, len(response.Items))

		// 次のページトークンを取得
		if response.NextPageToken != "" {
			pageToken = response.NextPageToken
			log.Printf("[YOUTUBE_API] Next page token found, continuing to fetch more messages...")
		} else {
			// 次のページがない場合は終了
			log.Printf("[YOUTUBE_API] No more pages, total pages fetched: %d", totalPages)
			break
		}
	}

	log.Printf("[YOUTUBE_API] Successfully retrieved total %d messages from %d pages", len(messages), totalPages)

	// デバッグ用：最初と最後のメッセージを出力
	if len(messages) > 0 {
		log.Printf("[YOUTUBE_API] First message: ID=%s, ChannelID=%s, DisplayName=%s, PublishedAt=%s",
			messages[0].ID, messages[0].ChannelID, messages[0].DisplayName, messages[0].PublishedAt.Format(time.RFC3339))
		if len(messages) > 1 {
			lastIdx := len(messages) - 1
			log.Printf("[YOUTUBE_API] Last message: ID=%s, ChannelID=%s, DisplayName=%s, PublishedAt=%s",
				messages[lastIdx].ID, messages[lastIdx].ChannelID, messages[lastIdx].DisplayName, messages[lastIdx].PublishedAt.Format(time.RFC3339))
		}
	}

	return messages, false, nil
}