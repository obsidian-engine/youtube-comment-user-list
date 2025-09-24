package domain

import (
	"testing"
	"time"
)

func TestUser_LatestCommentedAt(t *testing.T) {
	t.Run("User構造体にLatestCommentedAtフィールドが存在する", func(t *testing.T) {
		now := time.Now()
		user := User{
			ChannelID:         "UC123",
			DisplayName:       "TestUser",
			JoinedAt:          now,
			CommentCount:      5,
			FirstCommentedAt:  now.Add(-time.Hour),
			LatestCommentedAt: now.Add(-time.Minute), // 最新コメント時間
		}

		if user.LatestCommentedAt.IsZero() {
			t.Error("LatestCommentedAtが初期化されていません")
		}

		if user.LatestCommentedAt.Before(user.FirstCommentedAt) {
			t.Error("LatestCommentedAtはFirstCommentedAtより新しい時間であるべきです")
		}
	})

	t.Run("LatestCommentedAtのJSONタグが正しく設定されている", func(t *testing.T) {
		// この段階では構造体にLatestCommentedAtフィールドがないため、テストは失敗するはず
		// 実装後にJSONシリアライゼーションをテストする
	})
}