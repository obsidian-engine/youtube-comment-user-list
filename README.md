# YouTube Live Chat Monitor

> Clean Architecture Go アプリケーション - YouTube ライブチャット参加者監視＆OBSオーバーレイ対応

## 🎯 概要

YouTube Live配信のチャット参加者をリアルタイムで収集・監視し、OBSオーバーレイ表示に対応したWebアプリケーションです。

### ✨ 主な機能

- 📺 **YouTube Live Chat 監視** - リアルタイムチャット参加者収集
- 👥 **ユーザーリスト管理** - 参加者の表示名・権限・統計情報
- 🔄 **Server-Sent Events (SSE)** - リアルタイム更新配信
- 🎮 **OBS対応** - オーバーレイ表示用Web画面
- 📊 **システムログ** - 構造化ログ表示・管理
- 💾 **インメモリ高速処理** - 軽量・高速動作

## 🖌 UI スタイル / Pico.css 採用理由

本プロジェクトでは `@picocss/pico` を **ベースリセット & 軽量な初期スタイル** として利用し、その上から `static/ui.css` でコンポーネントデザインを完全上書きしています。

### 採用理由
- 0 からの CSS Reset 実装工数を削減（フォーム/タイポ系の初期値が整う）
- CSS ファイル 1 枚 / 依存レスで Cloud Run コールドスタートに影響しない
- Utility framework ではないため BEM/独自クラス構成を阻害しない

### 読み込み順 (base.gohtml)
1. Google Fonts / Material Symbols
2. Pico (`pico.min.css`)
3. プロジェクト固有スタイル (`/static/ui.css`) ← ここで上書き

```html
<link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
<link rel="stylesheet" href="/static/ui.css">
```

### 上書きポリシー
- Pico の要素セレクタは極力利用せず、`.btn`, `.card`, `.appbar`, `.sidebar` など固有クラスを定義
- 競合しやすい `button`, `a`, `table` などは **目的化クラス (.btn / .toolbar / .table-scroll)** に限定
- ダーク/ライト切替は `:root` と `html[data-theme]` でトークン管理。`prefers-color-scheme` を尊重しつつ、将来的に UI トグル追加可能。

### 余白/ラディアス/影トークン
`ui.css` の `:root` にまとめています。調整例:
```css
:root {
  --radius-lg: 14px; /* カード角 */
  --shadow-1: 0 1px 2px rgba(16,24,40,.06), 0 2px 6px rgba(16,24,40,.04);
}
```
**新規コンポーネント追加時の指針:** まず既存トークンを使う → 不足時のみトークンを追加し、要素側にハードコードしない。

### 衝突回避チェックリスト
- フォーム: Pico の `input, select` 初期スタイル → `.btn` / 独自 `input` 上書き済み
- テーブル: 独自 `border-spacing` / 行影 / ラウンドを適用（Pico のフラット表を置換）
- ボタン: `.btn` クラス必須。素の `<button>` にはテーマ差分が出ないようにする

---

## 🏗️ アーキテクチャ

Clean Architecture (Onion Architecture) に基づく4層構造：

```
├── cmd/server/              # Main Application (DI Container)
├── internal/
│   ├── domain/             # Domain Layer (Entities, Repository interfaces)
│   ├── application/        # Application Layer (Use Cases, Services)
│   ├── infrastructure/     # Infrastructure Layer (YouTube API, Events, Logging)
│   └── interfaces/         # Interface Layer (HTTP Handlers, SSE)
```

## 🚀 Cloud Run デプロイ（無料枠最適化済み）

### 📋 前提条件

- Google Cloud Project（無料枠）
- YouTube Data API v3 キー
- Docker
- gcloud CLI

### 🔧 環境設定

1. **環境変数ファイル作成**
```bash
cp .env.cloudrun .env
# .env ファイルを編集して実際の値を設定
```

2. **必要な環境変数**
```bash
# 必須
export YT_API_KEY="your_youtube_api_key_here"
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"

# オプション（無料枠最適化）
export MAX_CHAT_MESSAGES=500    # デフォルト500（メモリ節約）
export MAX_USERS=100            # デフォルト100（メモリ節約）
export LOG_LEVEL=WARN           # ログ量削減
```

### 🚢 デプロイ実行

```bash
# 一括デプロイ（推奨）
./deploy-cloud-run.sh

# または手動デプロイ
docker build -t gcr.io/$GOOGLE_CLOUD_PROJECT/youtube-chat-monitor .
docker push gcr.io/$GOOGLE_CLOUD_PROJECT/youtube-chat-monitor
gcloud run deploy youtube-chat-monitor \
  --image=gcr.io/$GOOGLE_CLOUD_PROJECT/youtube-chat-monitor \
  --memory=256Mi --cpu=0.167 --concurrency=1 --timeout=3600s \
  --max-instances=10 --min-instances=0 \
  --set-secrets="YT_API_KEY=youtube-api-secret:latest"
```

## 💰 無料枠最適化

### 📊 Cloud Run 無料枠制限

- **CPU**: 180,000 vCPU-秒/月
- **メモリ**: 360,000 GB-秒/月  
- **リクエスト**: 2,000,000 回/月

### ⚙️ 最適化設定

| 設定項目 | 値 | 理由 |
|---------|---|------|
| **CPU** | 0.167 vCPU | 最小構成（256Mi RAMに対応） |
| **メモリ** | 256Mi | 最小メモリ設定 |
| **並行実行数** | 1 | SSE接続特性に最適化 |
| **タイムアウト** | 60分 | Cloud Run最大SSE接続時間 |
| **最大インスタンス** | 10 | コスト制御 |
| **最小インスタンス** | 0 | 完全scale-to-zero |

### ⚠️ 重要な制限事項

- **SSE接続は60分で自動切断**されます（Cloud Run制限）
- **24時間常時接続では無料枠を超過**します（月額$2-4程度）
- クライアント側で自動再接続機能の実装が必要

## 🖥️ ローカル開発

### 📦 必要な依存関係

- Go 1.24+
- YouTube Data API v3 キー

### 🏃 実行方法

```bash
# 依存関係インストール
go mod download

# 環境変数設定
cp .env.cloudrun .env
# .env を編集

# アプリケーション起動
go run cmd/server/main.go
```

### 🌐 エンドポイント

| エンドポイント | 説明 |
|---------------|------|
| `GET /` | ホーム画面 |
| `GET /users` | ユーザーリスト表示 |
| `GET /logs` | システムログ表示 |
| `GET /health` | ヘルスチェック |
| `POST /api/monitoring/start` | 監視開始 |
| `DELETE /api/monitoring/stop` | 監視停止 |
| `GET /api/monitoring/active` | 現在監視中のvideoId取得 |
| `GET /api/monitoring/users` | アクティブセッションのユーザー一覧取得（推奨） |
| `GET /api/monitoring/{videoId}/users` | 指定videoIdのユーザー一覧取得（非推奨: `/api/monitoring/users` を使用） |
| `GET /api/monitoring/{videoId}/status` | 動画のライブ配信ステータス取得 |
| `GET /api/sse/{videoId}` | チャットSSE配信 |
| `GET /api/sse/{videoId}/users` | ユーザーリストSSE配信 |

## 🔧 技術スタック

- **言語**: Go 1.24
- **アーキテクチャ**: Clean Architecture
- **外部API**: YouTube Data API v3
- **HTTP Router**: go-chi/chi
- **ログ**: 構造化ログ (slog)
- **コンテナ**: Docker (Distroless)
- **デプロイ**: Google Cloud Run

## 📈 監視・運用

### 🏥 ヘルスチェック

```bash
# ヘルスチェック
curl https://your-service-url/health

# レディネスチェック  
curl https://your-service-url/ready
```

### 📊 使用量監視

Cloud Run使用量は以下で確認：
- [Cloud Run コンソール](https://console.cloud.google.com/run)
- [無料枠使用量](https://console.cloud.google.com/billing)

## 🤝 開発ガイドライン

### 🧪 テスト実行

```bash
go test ./...
go test -race ./...  # 競合状態検査
```

### 🔍 静的解析

```bash
go vet ./...
golangci-lint run
```

### 📝 コミットメッセージ

- feat: 新機能
- fix: バグ修正
- docs: ドキュメント更新
- refactor: リファクタリング

## 🐛 トラブルシューティング

### よくある問題

**Q: SSE接続が60分で切断される**  
A: Cloud Run の制限です。クライアント側で自動再接続を実装してください。

**Q: 無料枠を超過してしまう**  
A: `MAX_CHAT_MESSAGES`と`MAX_USERS`を更に削減するか、監視時間を制限してください。

**Q: YouTube API エラー**  
A: API キーの権限とクォータ制限を確認してください。

## 📄 ライセンス

MIT License

## 🙋 サポート

Issues や質問は [GitHub Issues](../../issues) までお願いします。

- Junie にコードをどう共有する？→ docs/how-to-share-code-with-junie.md をご覧ください。
