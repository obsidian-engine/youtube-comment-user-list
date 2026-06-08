package memory

import (
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

func TestCommentRepo_SearchByKeywords(t *testing.T) {
	t.Run("hit 0 で空 slice を返す (nil ではない)", func(t *testing.T) {
		repo := NewCommentRepo()
		if err := repo.Add(domain.Comment{ID: "1", Message: "hello world", PublishedAt: time.Now()}); err != nil {
			t.Fatalf("Add error: %v", err)
		}

		results := repo.SearchByKeywords([]string{"テスト🍔"})
		if results == nil {
			t.Fatal("expected empty slice, got nil (frontend が null.length で crash する原因)")
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("空 keyword 配列でも空 slice を返す", func(t *testing.T) {
		repo := NewCommentRepo()
		results := repo.SearchByKeywords([]string{})
		if results == nil {
			t.Fatal("expected empty slice, got nil")
		}
	})

	t.Run("Unicode 絵文字 keyword で正しく hit する", func(t *testing.T) {
		repo := NewCommentRepo()
		if err := repo.Add(domain.Comment{ID: "1", Message: "テスト🍔うまい", PublishedAt: time.Now()}); err != nil {
			t.Fatalf("Add error: %v", err)
		}
		if err := repo.Add(domain.Comment{ID: "2", Message: "別コメント", PublishedAt: time.Now()}); err != nil {
			t.Fatalf("Add error: %v", err)
		}

		results := repo.SearchByKeywords([]string{"🍔"})
		if len(results) != 1 {
			t.Fatalf("expected 1 hit, got %d", len(results))
		}
		if results[0].ID != "1" {
			t.Errorf("expected ID=1, got %s", results[0].ID)
		}
	})
}
