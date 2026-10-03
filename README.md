# YouTube Comment User List

YouTubeライブ配信のコメント投稿者を時系列で管理・表示するWebアプリケーション

## 🚀 機能概要

- **ライブ配信コメント取得**: YouTube Live Chat APIを使用してリアルタイムでコメントを取得
- **ユーザーリスト管理**: コメント投稿者を参加順（時系列）で表示・管理
- **自動更新**: 配信中は自動的に新規コメント投稿者を追加
- **配信予約**: 開始前の配信を予約すると、サーバーが配信開始を監視して自動でコメント取得を始める
- **投票集計**: キーワードを含むコメントを集計し、結果をTSVでコピーできる
- **配信履歴**: 過去配信のユーザーとコメントをスナップショットから閲覧できる

## 🏗️ アーキテクチャ

```text
Frontend (React/Vite) ⟷ Backend (Go) ⟷ YouTube API
                              ↓
                     Google Cloud Storage（スナップショット）
```

- **フロントエンド**: React + TypeScript + Vite
- **バックエンド**: Go + Clean Architecture
- **データ保存**: インメモリ + GCSスナップショット（`GCS_BUCKET` 設定時）
- **API**: RESTful API + CORS対応 + 共有APIキー認証（`X-API-Key` ヘッダー）

## 📋 状態管理仕様

### 🔄 アプリケーション状態

| 状態 | 説明 | UI表示 |
|------|------|--------|
| **WAITING** | 待機中（配信開始前・配信終了後） | 🟡 待機中 |
| **RESERVED** | 予約中（配信の開始予定時刻を待っている） | 予約中 + 開始予定時刻 |
| **ACTIVE** | 配信中（コメント取得中） | 🔴 配信中 |

予約中の配信は、サーバーのモニターが YouTube 側の実際の開始時刻を確認してから ACTIVE に切り替える。

### 💾 データ保存場所と持続性

#### **サーバーサイド（Go）**

- **保存場所**: サーバーのRAMメモリ内（ユーザー・コメント・配信状態）
- **スナップショット**: `GCS_BUCKET` を設定すると、メモリ内の状態を約60秒ごとと状態の切り替え時にGCSへ保存する
- **起動時の復元**: 起動時にGCSの最新スナップショットから状態を復元する（失敗時は空の状態で起動）
- **未設定時**: `GCS_BUCKET` が空ならスナップショットは無効で、サーバー再起動でデータは消える

GCSの準備手順は [backend/docs/gcs-snapshot-setup.md](backend/docs/gcs-snapshot-setup.md) を参照。

#### **クライアントサイド（React）**
- **保存場所**: APIから毎回取得
- **リロード対応**: ✅ サーバーが稼働中なら復元可能

### 🔧 操作とデータの変化

| 操作 | ユーザーリスト | アプリ状態 | 備考 |
|------|---------------|-----------|------|
| **ブラウザリロード** | ✅ 保持 | ✅ 保持 | サーバーメモリから復元 |
| **別の動画へ切り替え・予約** | ❌ **クリア** | ACTIVE / RESERVED | 同じ動画の再指定ではクリアしない |
| **配信終了（自動）** | ✅ 保持 | → WAITING | 終了時にスナップショットを保存 |
| **リセットボタン** | ❌ **クリア** | → WAITING | 手動で初期化 |
| **サーバー再起動** | GCS設定時は復元 | 復元 | GCS未設定なら消失 |

### 📊 YouTube API制限

#### **コメント取得の制限事項**
- **✅ 取得可能**: 取得を始めた後の新規コメント
- **❌ 取得不可**: 取得を始める前の過去コメント

| 取得を始めるタイミング | 取得可能なユーザー | 理由 |
|---------------------|------------------|------|
| 配信開始前に予約 | ✅ **ほぼ全員分** | 配信開始から監視 |
| 配信途中から | ❌ **新規のみ** | APIはリアルタイム取得のみ |
| 配信終了後 | ❌ **取得不可** | 配信終了で `Live Chat API` が使えなくなる |

### 🎯 最適な使用方法

1. **配信開始前に予約する** - 開始と同時に取得が始まり、全視聴者のコメントを取得できる
2. **GCSスナップショットを有効にする** - Cloud Run の再起動でデータを失わない
3. **複数配信時は動画を切り替える** - 前の配信のデータをクリアして新しい配信に対応

## 🛠️ API仕様

全エンドポイントで `X-API-Key` ヘッダーが必要（バックエンドの `API_KEY` と一致させる）。

### エンドポイント一覧

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/status` | アプリ状態・ユーザー数取得 |
| GET | `/users.json` | ユーザーリスト取得（時系列順） |
| POST | `/switch-video` | 配信切り替え（配信前の動画は予約として扱う） |
| POST | `/reserve` | 配信予約（body: `{"videoId": "..."}`、URLも可） |
| POST | `/cancel-reserve` | 予約を解除する |
| POST | `/pull` | コメント手動取得 |
| POST | `/reset` | 状態リセット |
| GET | `/comments?keywords=a,b` | キーワードを含むコメントの検索（最大20語、各100文字まで） |
| GET | `/history/snapshots` | 過去配信のスナップショット一覧 |
| GET | `/history/snapshots/{videoID}` | 過去配信のスナップショット詳細 |

### レスポンス例

#### `/status`
```json
{
  "status": "ACTIVE",
  "count": 42,
  "videoId": "ABC123",
  "liveChatId": "...",
  "startedAt": "2023-12-01T10:00:00Z",
  "endedAt": null,
  "lastPulledAt": "2023-12-01T10:05:00Z",
  "reservedAt": null,
  "scheduledStartTime": null,
  "autonomousMonitoring": true,
  "snapshotSavedAt": "2023-12-01T10:04:30Z"
}
```

#### `/users.json`
```json
[
  {
    "channelId": "UC...",
    "displayName": "ユーザー1",
    "joinedAt": "2023-12-01T10:05:30Z"
  }
]
```

## 🚀 起動方法

### 前提条件
- Node.js 20+
- Go 1.25+
- YouTube Data API v3キー

### バックエンド起動
```bash
cd backend
cp .env.example .env
# .env に YT_API_KEY / API_KEY / FRONTEND_ORIGIN を設定（GCS_BUCKET は任意）
go run cmd/server/main.go
```

### フロントエンド起動
```bash
cd frontend
cp .env.example .env
# .env に VITE_BACKEND_URL / VITE_API_KEY を設定
npm install
npm run dev
```

### テスト・Lint
```bash
cd backend && go test ./internal/...
cd frontend && npm run lint && npm run typecheck && npm run test
```

## 📝 開発ガイド

### ディレクトリ構造
```text
├── backend/              # Go REST API
│   ├── cmd/server/      # エントリーポイント
│   ├── internal/
│   │   ├── domain/      # 配信状態・ユーザー・コメント
│   │   ├── usecase/     # 切り替え・予約・取得・リセット・監視・スナップショット
│   │   ├── port/        # usecase が使う interface
│   │   └── adapter/     # HTTP / YouTube API / メモリ / GCS の実装
│   └── docs/            # GCS のセットアップ手順
├── frontend/             # React SPA
│   └── src/             # components / hooks / utils
└── .github/workflows/    # CI とバックエンドのデプロイ
```

### 技術スタック

**バックエンド**
- Go 1.25
- chi（HTTPルーター）
- Clean Architecture
- YouTube Data API v3
- Google Cloud Storage（スナップショット）
- Google Cloud Run（デプロイ）

**フロントエンド**
- React 18
- TypeScript
- Vite
- Tailwind CSS
- Vitest

**インフラ・CI/CD**
- GitHub Actions（main への push でバックエンドを Cloud Run へデプロイ）
- Google Cloud Run
- Docker

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. 機能ブランチを作成: `git checkout -b feature/amazing-feature`
3. 変更をコミット: `git commit -m 'Add: 素晴らしい機能'`
4. ブランチにプッシュ: `git push origin feature/amazing-feature`
5. プルリクエストを作成

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。
