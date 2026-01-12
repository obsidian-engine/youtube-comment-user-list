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

// isLiveChatEnded はエラーがライブチャットの終了または無効化を示すかどうかを判定します
func isLiveChatEnded(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	endedKeywords := []string{
		"forbidden",
		"livechatdisabled",
		"livechatended",
		"livechatnotfound",
		"chatdisabled",
		"livechatnotactive",
	}

	for _, keyword := range endedKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	// "notfound" + "livechat" の組み合わせもチェック
	return strings.Contains(errMsg, "notfound") && strings.Contains(errMsg, "livechat")
}

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

func (a *API) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) (items []port.ChatMessage, nextPageToken string, pollingIntervalMillis int64, isEnded bool, err error) {
	log.Printf("[YOUTUBE_API] ListLiveChatMessages called with liveChatID: %s", liveChatID)

	if a.APIKey == "" {
		log.Printf("[YOUTUBE_API] Error: API key is empty")
		return nil, "", 0, false, errors.New("youtube api key is required")
	}

	if liveChatID == "" {
		log.Printf("[YOUTUBE_API] Error: liveChatID is empty")
		return nil, "", 0, false, errors.New("live chat ID is required")
	}

	// YouTube Data API v3を使用してライブチャットメッセージを取得
	service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to create YouTube service: %v", err)
		return nil, "", 0, false, err
	}

	// Live Chat APIは1回の呼び出しで増分取得を行う設計
	// 無限ループを避けるため、1ページのみ取得
	call := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"})

	// ページトークンを設定（初回は空）
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	// Live Chat API仕様: デフォルト200、最大2000
	// 最大値に設定してより多くのコメントを一度に取得
	call = call.MaxResults(2000)

	response, err := call.Do()
	if err != nil {
		log.Printf("[YOUTUBE_API] API call failed: %v", err)

		if isLiveChatEnded(err) {
			log.Printf("[YOUTUBE_API] Live chat ended or disabled")
			return nil, "", 0, true, nil
		}

		return nil, "", 0, false, err
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
				ID:          item.Id,
				ChannelID:   item.AuthorDetails.ChannelId,
				DisplayName: item.AuthorDetails.DisplayName,
				Message:     item.Snippet.DisplayMessage,
				PublishedAt: publishedAt,
			})
		}
	}

	log.Printf("[YOUTUBE_API] Successfully retrieved %d messages (pageToken=%s next=%s, pollingIntervalMillis=%d)", len(messages), pageToken, response.NextPageToken, response.PollingIntervalMillis)

	return messages, response.NextPageToken, int64(response.PollingIntervalMillis), false, nil
}
