package port

import "context"

// ChatPoster はライブチャットへのメッセージの投稿を抽象化する。
// 実装は bot アカウントのログイン許可を使う。
type ChatPoster interface {
	PostMessage(ctx context.Context, liveChatID, text string) error
}
