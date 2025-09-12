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
	// HTTPReadTimeout HTTPサーバーの読み込みタイムアウト
	HTTPReadTimeout = 15 * time.Second

	// HTTPWriteTimeout HTTPサーバーの書き込みタイムアウト
	HTTPWriteTimeout = 15 * time.Second

	// HTTPIdleTimeout HTTPサーバーのアイドルタイムアウト
	HTTPIdleTimeout = 60 * time.Second

	// ShutdownTimeout アプリケーションシャットダウンのタイムアウト
	ShutdownTimeout = 10 * time.Second
)

// YouTube API Configuration
// YouTube API設定
const (
	// YouTubeVideoIDLength YouTube動画IDの標準長
	YouTubeVideoIDLength = 11

	// YouTubeChatMaxResults YouTube Live Chat APIの最大結果数
	YouTubeChatMaxResults = 2000

	// YouTubeHTTPClientTimeout YouTubeクライアントのHTTPタイムアウト
	YouTubeHTTPClientTimeout = 30 * time.Second

	// HTTPStatusOK 成功ステータスコード
	HTTPStatusOK = 200
)

// Polling Service Configuration
// ポーリングサービス設定
const (
	// MaxConsecutiveErrors 最大連続エラー数
	MaxConsecutiveErrors = 5

	// PollingBaseWaitTime ポーリングのベース待機時間
	PollingBaseWaitTime = 5 * time.Second

	// PollingMaxWaitTime ポーリングの最大待機時間
	PollingMaxWaitTime = 60 * time.Second

	// ExponentialBackoffMultiplier 指数バックオフの乗数
	ExponentialBackoffMultiplier = 2
)

// Server-Sent Events Configuration
// Server-Sent Events設定
const (
	// SSEHeartbeatInterval SSEのハートビート送信間隔
	SSEHeartbeatInterval = 30 * time.Second

	// SSEConnectionTimeout SSEコネクションのタイムアウト
	SSEConnectionTimeout = 5 * time.Minute

	// SSEUserListUpdateInterval SSEユーザーリスト更新間隔
	SSEUserListUpdateInterval = 10 * time.Second

	// SSEReconnectDelay SSE再接続の遅延時間
	SSEReconnectDelay = 5 * time.Second
)

// URL Path Configuration
// URLパス設定
const (
	// MinPathPartsForAPI APIパスの最小セグメント数
	MinPathPartsForAPI = 3

	// MinPathPartsForVideoID 動画ID付きパスの最小セグメント数
	MinPathPartsForVideoID = 4
)

// Log Auto Refresh Configuration
// ログ自動更新設定
const (
	// LogAutoRefreshInterval ログの自動更新間隔（ミリ秒）
	LogAutoRefreshInterval = 30000 // 30 seconds
)

// HTML Form Configuration
// HTMLフォーム設定
const (
	// HTMLMaxUsersMin HTMLフォームでの最大ユーザー数の最小値
	HTMLMaxUsersMin = 1

	// HTMLMaxUsersMax HTMLフォームでの最大ユーザー数の最大値
	HTMLMaxUsersMax = 10000

	// HTMLLogLimitOptions HTMLログ制限のオプション値
	HTMLLogLimit50  = 50
	HTMLLogLimit100 = 100
	HTMLLogLimit200 = 200
	HTMLLogLimit500 = 500
)

// Time Format Constants
// 時刻フォーマット定数
const (
	// TimeFormatISO8601 ISO8601タイムスタンプフォーマット
	TimeFormatISO8601 = "2006-01-02T15:04:05Z07:00"

	// TimeFormatLog ログ用タイムスタンプフォーマット
	TimeFormatLog = "2006-01-02 15:04:05.000"

	// TimeFormatSimple シンプルなタイムスタンプフォーマット
	TimeFormatSimple = "2006-01-02 15:04:05"
)

// Validation Constants
// バリデーション定数
const (
	// MinValidMaxUsers 最大ユーザー数の最小有効値
	MinValidMaxUsers = 0

	// MinValidLimit 制限値の最小有効値
	MinValidLimit = 0

	// EmptySliceLength 空のスライスの長さ
	EmptySliceLength = 0

	// SingleItemIndex 単一アイテムのインデックス
	SingleItemIndex = 0

	// InitialCounterValue 初期カウンター値
	InitialCounterValue = 0

	// BackoffInitialStep バックオフの初期ステップ
	BackoffInitialStep = 1
)
