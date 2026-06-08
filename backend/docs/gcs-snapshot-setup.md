# GCS Snapshot 永続化 setup 手順

## 概要

backend の in-memory state (user / comment) を GCS に snapshot 保存する。
Cloud Run idle 時の state 消失を防ぎ、次の配信開始時に直前の状態を復元する。

## GCP 側の準備

### 1. GCS bucket 作成

- bucket 名: 任意 (例 `yt-livechat-snapshots-<project>`)
- region: Cloud Run と同じ (現状 `us-central1`) ※ egress 無料化のため
- storage class: Standard
- 公開設定: private (Uniform bucket-level access 推奨)

gcloud 例:

```bash
gcloud storage buckets create gs://<bucket-name> --location=us-central1 --uniform-bucket-level-access
```

### 2. Lifecycle Rule (7 日後自動削除)

古い snapshot の自動削除 rule 設定。

gcloud 例:

```bash
cat <<EOF > lifecycle.json
{
  "rule": [
    {
      "action": {"type": "Delete"},
      "condition": {"age": 7}
    }
  ]
}
EOF
gcloud storage buckets update gs://<bucket-name> --lifecycle-file=lifecycle.json
```

### 3. Service Account 権限付与

Cloud Run の service account に bucket scoped で `roles/storage.objectAdmin` 付与。

```bash
gcloud storage buckets add-iam-policy-binding gs://<bucket-name> \
  --member="serviceAccount:<sa-email>" \
  --role="roles/storage.objectAdmin"
```

## GitHub 側の準備

### Secrets 追加

`Settings > Secrets and variables > Actions` に追加:

- `GCS_BUCKET`: 上記 1 で作成した bucket 名

## 動作確認

1. main branch push → deploy 走る
2. Cloud Run service の env vars に `GCS_BUCKET` が反映されているか確認
3. /switchVideo API 叩く → `gs://<bucket-name>/snapshots/current.json` と `snapshots/<videoID>.json` が生成されるか確認
4. Cloud Run instance を強制停止 → 再起動 → 同じ video state が復元されるか確認

## 失敗時挙動

- GCS bucket 不在 / 権限不足 → server 起動時 warn log、空 state で続行 (致命的でない)
- save 失敗 → warn log + 次 trigger で再試行、in-memory は維持
- GCS_BUCKET env が空 → no-op mode (snapshot 機能無効、従来挙動)

## ローカル開発

- `.env` に `GCS_BUCKET=` (空) を設定 → no-op mode
- GCS 連動を試す場合: `gcloud auth application-default login` で ADC 設定 + `GCS_BUCKET=<bucket-name>` 設定
