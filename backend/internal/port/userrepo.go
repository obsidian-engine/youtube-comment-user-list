package port

// UserRepo はユーザー一覧（重複排除）を管理します。
type UserRepo interface {
    // Upsert は channelID をキーに displayName を登録/更新します。
    Upsert(channelID string, displayName string) error
    // ListDisplayNames は displayName の配列（安定順）を返します。
    ListDisplayNames() []string
    // Count は登録ユーザー数を返します。
    Count() int
    // Clear は全ユーザーを削除します。
    Clear()
}
