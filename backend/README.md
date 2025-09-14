# YouTube Live Comment User List - Backend API

YouTube Live配信のコメント参加者を収集・管理するバックエンドAPI

## 🚀 セットアップ

### 1. 環境変数設定

```bash
# 環境変数ファイルを作成
cp .env.example .env

# .env ファイルを編集して必要な値を設定
vi .env
```

### 2. 必須設定項目

- **YT_API_KEY**: YouTube Data API v3 キー（必須）
  - [Google Cloud Console](https://console.cloud.google.com/apis/credentials) で取得
  - YouTube Data API v3 を有効化してAPIキーを作成

### 3. 依存関係インストール

```bash
go mod download
```

### 4. サーバー起動

```bash
# 開発環境
go run cmd/server/main.go

# または
go build -o server cmd/server/main.go
./server
```

## 📋 API エンドポイント

| Method | Endpoint | 説明 |
|--------|----------|------|
| GET | `/status` | 現在のライブ状態とユーザー数を取得 |
| GET | `/users.json` | 参加者一覧を取得 |
| POST | `/switch-video` | 配信URLを切り替え |
| POST | `/pull` | コメント参加者を収集 |
| POST | `/reset` | 参加者リストをリセット |

## 🧪 テスト実行

```bash
# 全テスト実行
go test ./...

# 詳細出力
go test -v ./...
```

## 🔧 開発ツール

```bash
# コードフォーマット
go fmt ./...

# 静的解析
go vet ./...

# Lint（golangci-lint）
golangci-lint run
```

## 🏗️ アーキテクチャ

Clean Architecture パターンを採用

```
backend/
├── cmd/server/          # エントリーポイント
├── internal/
│   ├── adapter/         # 外部システムとの接続層
│   │   ├── http/        # HTTPハンドラー
│   │   ├── memory/      # インメモリ実装
│   │   └── youtube/     # YouTube API
│   ├── usecase/         # ビジネスロジック
│   ├── domain/          # ドメインモデル
│   └── port/            # インターフェース定義
└── Dockerfile           # Cloud Run用コンテナ
```

## 🌐 デプロイ

### Cloud Run

```bash
# Cloud Runにデプロイ
gcloud run deploy youtube-comment-backend \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## 📝 環境変数詳細

| 変数名 | 必須 | デフォルト | 説明 |
|--------|------|------------|------|
| YT_API_KEY | ✅ | - | YouTube Data API v3 キー |
| PORT | - | 8080 | サーバーポート |
| FRONTEND_ORIGIN | - | - | CORS許可オリジン |
| LOG_LEVEL | - | info | ログレベル (debug/info/warn/error) |
| GO_ENV | - | development | 環境識別子 |