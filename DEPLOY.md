# 🚀 YouTube Live Chat アプリ 無料デプロイ手順

## 📋 必要な準備

### 1. Google Cloud Platform セットアップ

#### GCPプロジェクト作成
```bash
# Google Cloud SDKをインストール済みの場合
gcloud projects create [PROJECT_ID] --name="YouTube Live Chat"
gcloud config set project [PROJECT_ID]
```

#### 必要なAPIの有効化
```bash
gcloud services enable run.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

#### Artifact Registry リポジトリ作成
```bash
gcloud artifacts repositories create yt-livechat \
  --repository-format=docker \
  --location=us-central1
```

#### サービスアカウント作成と権限設定
```bash
# サービスアカウント作成
gcloud iam service-accounts create github-actions \
  --description="GitHub Actions deployment" \
  --display-name="GitHub Actions"

# 必要な権限を付与
gcloud projects add-iam-policy-binding [PROJECT_ID] \
  --member="serviceAccount:github-actions@[PROJECT_ID].iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding [PROJECT_ID] \
  --member="serviceAccount:github-actions@[PROJECT_ID].iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding [PROJECT_ID] \
  --member="serviceAccount:github-actions@[PROJECT_ID].iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# サービスアカウントキーを生成
gcloud iam service-accounts keys create key.json \
  --iam-account=github-actions@[PROJECT_ID].iam.gserviceaccount.com
```

### 2. GitHub Secrets 設定

GitHubリポジトリの Settings > Secrets and variables > Actions で以下を設定：

| Secret名 | 値 | 説明 |
|---------|---|------|
| `GCP_PROJECT_ID` | `[あなたのプロジェクトID]` | GCPプロジェクトID |
| `GCP_SA_KEY` | `key.json`の内容 | サービスアカウントの認証キー（JSON全体） |
| `YT_API_KEY` | `AIzaSy...` | YouTube Data API v3 キー |
| `YT_VIDEO_ID` | `kXpv3asP0Qw` | 対象のライブ配信動画ID |

#### YouTube API キー取得方法
1. [Google Cloud Console](https://console.cloud.google.com/) で YouTube Data API v3 を有効化
2. 「認証情報」→「認証情報を作成」→「APIキー」
3. 作成されたAPIキーを `YT_API_KEY` に設定

### 3. デプロイ実行

#### 自動デプロイ
```bash
git add .
git commit -m "feat: setup deployment configuration"
git push origin main
```

#### 手動デプロイ
GitHub リポジトリの Actions タブから「Deploy to Google Cloud Run」ワークフローを手動実行

## 🛡️ セキュリティ対策

### API キー保護
- ✅ `.env` ファイルを `.gitignore` で除外
- ✅ GitHub Secrets で環境変数を管理
- ✅ Cloud Run 環境変数として安全に注入

### ネットワーク制限
```bash
# 特定のIPからのみアクセス許可する場合（オプション）
gcloud run services update yt-livechat \
  --region=us-central1 \
  --ingress=internal-and-cloud-load-balancing
```

### 監査ログ
```bash
# Cloud Run のログを確認
gcloud logs read "resource.type=cloud_run_revision AND resource.labels.service_name=yt-livechat" \
  --limit=50 --format="table(timestamp,textPayload)"
```

## 💰 コスト管理

### Always Free 枠内運用
- **CPU**: 0.1 vCPU × 24h × 30日 = 72 vCPU時間（枠内: 180,000 vCPU秒）
- **メモリ**: 256MB × 24h × 30日 = 184GB時間（枠内: 360,000 GB秒）
- **リクエスト**: 個人利用なら月間数千リクエスト（枠内: 200万リクエスト）

### コスト監視設定
```bash
# 課金アラート設定（月$1超過で通知）
gcloud alpha billing budgets create \
  --billing-account=[BILLING_ACCOUNT_ID] \
  --display-name="YouTube Live Chat Budget" \
  --budget-amount=1.00USD \
  --threshold-rules-percent=50,90,100
```

## 🔧 トラブルシューティング

### デプロイエラー
```bash
# GitHub Actions のログを確認
# Cloud Run のログを確認
gcloud logs read "resource.type=cloud_run_revision" --limit=20

# サービスの状態確認
gcloud run services describe yt-livechat --region=us-central1
```

### 権限エラー
```bash
# サービスアカウントの権限確認
gcloud projects get-iam-policy [PROJECT_ID] \
  --filter="bindings.members:github-actions@[PROJECT_ID].iam.gserviceaccount.com"
```

### API制限エラー
```bash
# YouTube API の使用量確認
gcloud services list --enabled --filter="name:youtube.googleapis.com"
```

## 📊 運用監視

### アクセス確認
```bash
# デプロイされたURL確認
gcloud run services describe yt-livechat \
  --region=us-central1 \
  --format='value(status.url)'
```

### パフォーマンス監視
- Cloud Run コンソールでメトリクス確認
- `/users.json` エンドポイントで動作確認
- `/overlay` でOBS用画面確認

## 🔄 アップデート方法

### 動画IDの変更
1. GitHub Secrets で `YT_VIDEO_ID` を更新
2. Actions から手動デプロイ実行

### コードの更新
1. ローカルで修正・テスト
2. `git push origin main` で自動デプロイ

---

## 📞 サポート

- **GitHub Issues**: バグ報告・機能要求
- **Cloud Console**: GCP関連の監視・設定
- **YouTube Creator Studio**: 配信設定の確認

**🎉 これで個人用YouTube Live Chatアプリが無料で24時間稼働します！**