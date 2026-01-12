package domain

import "time"

// Comment はYouTube Live Chatのコメント情報です。
type Comment struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channelId"`
	DisplayName string    `json:"displayName"`
	Message     string    `json:"message"`
	PublishedAt time.Time `json:"publishedAt"`
}
