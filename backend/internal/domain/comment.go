package domain

import "time"

// Comment はYouTube Live Chatのコメント情報です。
type Comment struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channelId"`
	DisplayName string    `json:"displayName"`
	Handle      string    `json:"handle"`
	Message     string    `json:"message"`
	PublishedAt time.Time `json:"publishedAt"`
}
