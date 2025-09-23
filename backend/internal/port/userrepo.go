package port

// UserRepo はユーザー一覧（重複排除）を管理します。
import (
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

type UserRepo interface {
	// UpsertWithJoinTime は channelID をキーに displayName と初回参加時間を登録/更新します。
	// 既に存在するユーザーの場合、joinedAt は更新されません。
	UpsertWithJoinTime(channelID string, displayName string, joinedAt time.Time) error
	// UpsertWithMessage は channelID をキーに displayName と初回参加時間を登録/更新します。
	// messageID による重複チェックを行い、同じメッセージIDの場合は処理をスキップします。
	UpsertWithMessage(channelID string, displayName string, joinedAt time.Time, messageID string) error
	// UpsertWithMessageUpdated は UpsertWithMessage と同様だが、実際に更新されたかどうかを返します。
	// 重複メッセージの場合は false を返し、新規または更新の場合は true を返します。
	UpsertWithMessageUpdated(channelID string, displayName string, joinedAt time.Time, messageID string) (bool, error)
	// ListUsersSortedByJoinTime は User構造体の配列を参加時間順（早い順）で返します。
	ListUsersSortedByJoinTime() []domain.User
	// Count は登録ユーザー数を返します。
	Count() int
	// Clear は全ユーザーを削除します。
	Clear()
}
