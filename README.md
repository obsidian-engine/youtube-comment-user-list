# YouTube Comment User List

YouTubeライブ配信のコメント投稿者を時系列で管理・表示するWebアプリケーション

## 🚀 機能概要

- **ライブ配信コメント取得**: YouTube Live Chat APIを使用してリアルタイムでコメントを取得
- **ユーザーリスト管理**: コメント投稿者を参加順（時系列）で表示・管理
- **自動更新**: 配信中は自動的に新規コメント投稿者を追加
- **状態管理**: 配信の開始・終了・リセットに対応

## 🏗️ アーキテクチャ

```
Frontend (React/Vite) ⟷ Backend (Go) ⟷ YouTube API
```

- **フロントエンド**: React + TypeScript + Vite
- **バックエンド**: Go + Clean Architecture
- **データ保存**: インメモリ（サーバー再起動で消失）
- **API**: RESTful API + CORS対応

## 📋 状態管理仕様

### 🔄 アプリケーション状態

| 状態 | 説明 | UI表示 |
|------|------|--------|
| **WAITING** | 待機中（配信開始前） | 🟡 待機中 |
| **ACTIVE** | 配信中（コメント取得中） | 🔴 配信中 |

### 💾 データ保存場所と持続性

#### **サーバーサイド（Go）**
```go
type UserRepo struct {
    mu        sync.RWMutex
    usersByID map[string]domain.User // メモリ内マップ
}
```

- **保存場所**: サーバーのRAMメモリ内
- **持続性**: ❌ 揮発性（サーバー再起動でデータ消失）
- **同期**: RWMutexでスレッドセーフを保証

#### **クライアントサイド（React）**
- **保存場所**: APIから毎回取得（ローカルストレージなし）
- **リロード対応**: ✅ サーバーが稼働中なら復元可能

### 🔧 操作とデータの変化

| 操作 | ユーザーリスト | アプリ状態 | データ持続性 |
|------|---------------|-----------|-------------|
| **ブラウザリロード** | ✅ 保持 | ✅ 保持 | サーバーメモリから復元 |
| **動画切り替え** | ❌ **クリア** | ACTIVE継続 | 新配信用にリセット |
| **配信終了（自動）** | ❌ **クリア** | → WAITING | 自動リセット |
| **リセットボタン** | ❌ **クリア** | → WAITING | 手動初期化 |
| **サーバー再起動** | ❌ **消失** | → WAITING | 完全初期化 |

### 📊 YouTube API制限

#### **コメント取得の制限事項**
- **✅ 取得可能**: アプリ起動以降の新規コメント
- **❌ 取得不可**: 起動前の過去コメント

| アプリ起動タイミング | 取得可能なユーザー | 理由 |
|---------------------|------------------|------|
| 配信開始と同時 | ✅ **ほぼ全員分** | 最初から監視 |
| 配信途中から | ❌ **新規のみ** | APIはリアルタイム取得のみ |
| 配信終了後 | ❌ **取得不可** | Live Chat API無効化 |

### 🎯 最適な使用方法

1. **配信開始前にアプリ起動** - 全視聴者のコメントを確実に取得
2. **配信終了まで稼働継続** - サーバー停止による情報消失を防止
3. **複数配信時は動画切り替え機能使用** - リセットして新しい配信に対応

## 🛠️ API仕様

### エンドポイント一覧

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/status` | アプリ状態・ユーザー数取得 |
| GET | `/users.json` | ユーザーリスト取得（時系列順） |
| POST | `/switch-video` | 配信切り替え |
| POST | `/pull` | コメント手動取得 |
| POST | `/reset` | 状態リセット |

### レスポンス例

#### `/status`
```json
{
  "status": "ACTIVE",
  "videoId": "ABC123",
  "startedAt": "2023-12-01T10:00:00Z",
  "count": 42
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
- Node.js 18+
- Go 1.21+
- YouTube Data API v3キー

### バックエンド起動
```bash
cd backend
cp .env.example .env
# .envファイルにYouTube API キーを設定
go run cmd/server/main.go
```

### フロントエンド起動
```bash
cd frontend
npm install
npm run dev
```

## 📝 開発ガイド

### ディレクトリ構造
```
├── backend/           # Go REST API
│   ├── cmd/          # エントリーポイント
│   ├── internal/     # ビジネスロジック
│   └── tests/        # テストコード
├── frontend/         # React SPA
│   ├── src/          # ソースコード
│   └── public/       # 静的ファイル
└── .github/          # CI/CD設定
```

### 技術スタック

**バックエンド**
- Go 1.21
- Clean Architecture
- YouTube Data API v3
- Google Cloud Run（デプロイ）

**フロントエンド**
- React 18
- TypeScript
- Vite
- CSS3

**インフラ・CI/CD**
- GitHub Actions
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