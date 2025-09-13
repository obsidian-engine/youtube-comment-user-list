#!/bin/bash

# YouTube Live Chat Monitor - Cloud Run デプロイスクリプト
# 無料枠最適化済み

set -e

# 設定
PROJECT_ID="${GOOGLE_CLOUD_PROJECT}"
REGION="${REGION:-us-central1}"
SERVICE_NAME="youtube-chat-monitor"
IMAGE_TAG="${IMAGE_TAG:-latest}"

# プロジェクトIDの確認
if [ -z "$PROJECT_ID" ]; then
    echo "❌ GOOGLE_CLOUD_PROJECT環境変数が設定されていません"
    echo "gcloud config set project YOUR_PROJECT_ID を実行してください"
    exit 1
fi

echo "🚀 Cloud Run デプロイを開始します..."
echo "プロジェクトID: $PROJECT_ID"
echo "リージョン: $REGION"
echo "サービス名: $SERVICE_NAME"

# YouTube API キーの確認
if [ -z "$YT_API_KEY" ]; then
    echo "❌ YT_API_KEY環境変数が設定されていません"
    echo "export YT_API_KEY='your_api_key_here' を実行してください"
    exit 1
fi

# Docker イメージをビルド
echo "📦 Docker イメージをビルド中..."
docker build -t gcr.io/${PROJECT_ID}/${SERVICE_NAME}:${IMAGE_TAG} .

# Container Registry にプッシュ
echo "⬆️  Container Registry にプッシュ中..."
docker push gcr.io/${PROJECT_ID}/${SERVICE_NAME}:${IMAGE_TAG}

# Secret を作成（存在しない場合のみ）
echo "🔐 Secret を作成中..."
gcloud secrets describe youtube-api-secret --project=$PROJECT_ID >/dev/null 2>&1 || \
    echo -n "$YT_API_KEY" | gcloud secrets create youtube-api-secret \
        --project=$PROJECT_ID \
        --data-file=-

# Cloud Run サービスをデプロイ
echo "🌐 Cloud Run サービスをデプロイ中..."
gcloud run deploy ${SERVICE_NAME} \
    --image=gcr.io/${PROJECT_ID}/${SERVICE_NAME}:${IMAGE_TAG} \
    --project=${PROJECT_ID} \
    --region=${REGION} \
    --platform=managed \
    --allow-unauthenticated \
    --memory=256Mi \
    --cpu=0.167 \
    --concurrency=1 \
    --timeout=3600s \
    --max-instances=10 \
    --min-instances=0 \
    --execution-environment=gen2 \
    --cpu-throttling \
    --set-env-vars="MAX_CHAT_MESSAGES=500,MAX_USERS=100,LOG_LEVEL=WARN" \
    --set-secrets="YT_API_KEY=youtube-api-secret:latest"

# サービスURLを取得
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} \
    --project=${PROJECT_ID} \
    --region=${REGION} \
    --format="value(status.url)")

echo ""
echo "✅ デプロイが完了しました！"
echo "🌐 サービスURL: ${SERVICE_URL}"
echo "🏥 ヘルスチェック: ${SERVICE_URL}/health"
echo "📊 ユーザーリスト: ${SERVICE_URL}/users"
echo ""
echo "💡 使用量監視のために以下をブックマークしてください："
echo "   Cloud Run コンソール: https://console.cloud.google.com/run/detail/${REGION}/${SERVICE_NAME}/metrics?project=${PROJECT_ID}"
echo "   無料枠使用量: https://console.cloud.google.com/billing/consumption?project=${PROJECT_ID}"
echo ""
echo "📋 無料枠制限："
echo "   - CPU: 180,000 vCPU-秒/月"
echo "   - メモリ: 360,000 GB-秒/月"
echo "   - リクエスト: 2,000,000 回/月"
echo ""
echo "⚠️  注意: SSE接続は60分で自動切断されます。クライアント側で自動再接続を実装してください。"