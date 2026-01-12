package port

import "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"

// CommentRepo はコメントの永続化とクエリを提供します。
type CommentRepo interface {
	// Add はコメントを追加します（重複IDは無視）
	Add(comment domain.Comment) error

	// SearchByKeywords はキーワードでコメントを検索します（OR検索）
	// 結果は時系列順（古い順）
	SearchByKeywords(keywords []string) []domain.Comment

	// Clear は全コメントを削除します
	Clear()

	// Count は保存されているコメント数を返します
	Count() int
}
