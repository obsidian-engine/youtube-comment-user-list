// Package constants すべてのアプリケーション定数とマジックナンバーを定義します
package constants

import "time"

// Repository Configuration
// リポジトリ設定
const (
	// DefaultMaxChatMessages デフォルトの最大チャットメッセージ数
	DefaultMaxChatMessages = 10000

	// DefaultMaxUsers デフォルトの最大ユーザー数
	DefaultMaxUsers = 1000

	// DefaultMaxLogEntries デフォルトの最大ログエントリ数
	DefaultMaxLogEntries = 1000

	// ChatMessageChannelBuffer チャットメッセージチャンネルのバッファサイズ
	ChatMessageChannelBuffer = 100

	// DefaultLogDisplayLimit デフォルトのログ表示制限数
	DefaultLogDisplayLimit = 100
)

// HTTP Server Configuration
// HTTPサーバー設定
const (
	// ShutdownTimeout アプリケーションシャットダウンのタイムアウト
	ShutdownTimeout = 10 * time.Second

	// IdleTimeout アイドル状態でのサーバー自動停止タイムアウト
	IdleTimeout = 30 * time.Minute
)

// YouTube API Configuration
// YouTube API設定
const (
	// YouTubeVideoIDLength YouTube動画IDの標準長
	YouTubeVideoIDLength = 11

	// YouTubeChatMaxResults YouTube Live Chat APIの最大結果数
	YouTubeChatMaxResults = 2000

	// YouTubeHTTPClientTimeout YouTubeクライアントのHTTPタイムアウト
	YouTubeHTTPClientTimeout = 45 * time.Second

	// YouTubeAPIMaxRetries YouTube API リトライの最大回数
	YouTubeAPIMaxRetries = 3

	// YouTubeAPIInitialRetryDelay YouTube API リトライの初期遅延
	YouTubeAPIInitialRetryDelay = 1 * time.Second

	// YouTubeAPIMaxRetryDelay YouTube API リトライの最大遅延
	YouTubeAPIMaxRetryDelay = 30 * time.Second

	// YouTubeAPIRetryMultiplier YouTube API リトライの指数バックオフ乗数
	YouTubeAPIRetryMultiplier = 2

	// HTTPStatusOK 成功ステータスコード
	HTTPStatusOK = 200
)

// Polling Service Configuration
// ポーリングサービス設定
const (
	// DefaultPollingIntervalMs デフォルトポーリング間隔（ミリ秒）
	DefaultPollingIntervalMs = 10000
	// MinPollingIntervalMs 最低保証ポーリング間隔（ミリ秒） - APIがより短い値を返してもこれ未満にはしない
	MinPollingIntervalMs = 10000
)

// Server-Sent Events Configuration
// Server-Sent Events設定
const (
	// SSEHeartbeatInterval SSEのハートビート送信間隔
	SSEHeartbeatInterval = 30 * time.Second

	// SSEConnectionTimeout SSEコネクションのタイムアウト
	SSEConnectionTimeout = 5 * time.Minute

	// SSEUserListUpdateInterval SSEユーザーリスト更新間隔
	SSEUserListUpdateInterval = 60 * time.Second

	// SSEUserListHeartbeatInterval ユーザーリストSSE専用ハートビート間隔（更新間隔が長いので別途）
	SSEUserListHeartbeatInterval = 30 * time.Second
)

// Time Format Constants
// 時刻フォーマット定数
const (
	// TimeFormatISO8601 ISO8601タイムスタンプフォーマット
	TimeFormatISO8601 = "2006-01-02T15:04:05Z07:00"

	// TimeFormatLog ログ用タイムスタンプフォーマット
	TimeFormatLog = "2006-01-02 15:04:05.000"
)

// Validation Constants
// バリデーション定数
const (
	// MinValidLimit 制限値の最小有効値
	MinValidLimit = 0
)
