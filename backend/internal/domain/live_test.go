package domain

import (
	"encoding/json"
	"strings"
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

	t.Run("LatestCommentedAtのJSONシリアライゼーションが正しく動作する", func(t *testing.T) {
		now := time.Now()
		user := User{
			ChannelID:         "UC123",
			DisplayName:       "TestUser",
			JoinedAt:          now,
			CommentCount:      5,
			FirstCommentedAt:  now.Add(-time.Hour),
			LatestCommentedAt: now.Add(-time.Minute),
		}

		// JSON にシリアライズ
		jsonData, err := json.Marshal(user)
		if err != nil {
			t.Fatalf("JSON Marshal failed: %v", err)
		}

		// latestCommentedAt フィールドが含まれることを確認
		jsonStr := string(jsonData)
		if !strings.Contains(jsonStr, "latestCommentedAt") {
			t.Error("JSON should contain latestCommentedAt field")
		}

		// JSON からデシリアライズ
		var deserializedUser User
		err = json.Unmarshal(jsonData, &deserializedUser)
		if err != nil {
			t.Fatalf("JSON Unmarshal failed: %v", err)
		}

		// デシリアライズしたデータが元のデータと一致することを確認
		if !deserializedUser.LatestCommentedAt.Equal(user.LatestCommentedAt) {
			t.Errorf("LatestCommentedAt not preserved: got %v, want %v", 
				deserializedUser.LatestCommentedAt, user.LatestCommentedAt)
		}
	})
}