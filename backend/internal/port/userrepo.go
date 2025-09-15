package port

// UserRepo はユーザー一覧（重複排除）を管理します。
import (
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

type UserRepo interface {
	// Upsert は channelID をキーに displayName を登録/更新します。
	Upsert(channelID string, displayName string) error
	// ListDisplayNames は displayName の配列（安定順）を返します。
	ListDisplayNames() []string
	// Count は登録ユーザー数を返します。
	Count() int
	// Clear は全ユーザーを削除します。
	// UpsertWithJoinTime は channelID をキーに displayName と初回参加時間を登録/更新します。
	// 既に存在するユーザーの場合、joinedAt は更新されません。
	UpsertWithJoinTime(channelID string, displayName string, joinedAt time.Time) error
	// ListUsersSortedByJoinTime は User構造体の配列を参加時間順（早い順）で返します。
	ListUsersSortedByJoinTime() []domain.User

	Clear()
}
