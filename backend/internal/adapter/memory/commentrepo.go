package memory

import (
	"sort"
	"strings"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// CommentRepo はコメントをメモリ内に保存するリポジトリです。
type CommentRepo struct {
	mu       sync.RWMutex
	comments map[string]domain.Comment // ID -> Comment
}

// NewCommentRepo は新しいCommentRepoを作成します。
func NewCommentRepo() *CommentRepo {
	return &CommentRepo{
		comments: make(map[string]domain.Comment),
	}
}

// Add はコメントを追加します（重複IDは無視）
func (r *CommentRepo) Add(comment domain.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 重複チェック
	if _, exists := r.comments[comment.ID]; exists {
		return nil
	}

	r.comments[comment.ID] = comment
	return nil
}

// SearchByKeywords はキーワードでコメントを検索します（OR検索）
// 結果は時系列順（古い順）
func (r *CommentRepo) SearchByKeywords(keywords []string) []domain.Comment {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(keywords) == 0 {
		return []domain.Comment{}
	}

	var results []domain.Comment
	for _, comment := range r.comments {
		for _, keyword := range keywords {
			if strings.Contains(comment.Message, keyword) {
				results = append(results, comment)
				break // OR検索なので1つでもマッチしたら追加
			}
		}
	}

	// 時系列順（古い順）でソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].PublishedAt.Before(results[j].PublishedAt)
	})

	return results
}

// Clear は全コメントを削除します
func (r *CommentRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.comments = make(map[string]domain.Comment)
}

// Count は保存されているコメント数を返します
func (r *CommentRepo) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.comments)
}
