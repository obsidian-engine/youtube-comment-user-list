# トラブルシューティングガイド

## "更新に失敗しました" エラー

### 症状
- フロントエンドで「今すぐ取得」ボタンを押すと「更新に失敗しました。しばらくしてから再試行してください。」というエラーが表示される

### 原因
- YouTube API キーが正しく読み込まれていない
- バックエンドサーバーが起動していない、または古い状態

### 解決方法

1. **環境変数の確認**
   ```bash
   cd backend
   cat .env
   ```
   - `YT_API_KEY` が正しく設定されているか確認

2. **バックエンドサーバーの再起動**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

3. **動作確認**
   ```bash
   # ステータス確認
   curl http://localhost:8080/status

   # 配信切り替え
   curl -X POST -H "Content-Type: application/json" \
     -d '{"videoId":"YOUR_VIDEO_ID"}' \
     http://localhost:8080/switch-video

   # コメント取得
   curl -X POST http://localhost:8080/pull
   ```

### ログの確認方法

サーバー起動時に以下のようなログが表示されることを確認：
```
Config loaded successfully:
  Port: 8080
  Frontend Origin: http****3000
  YouTube API Key: AIza****WYsE
  Log Level: info
Server starting on port 8080
```

### 関連する設定ファイル

- `backend/.env` - 環境変数設定
- `backend/internal/config/config.go` - 設定読み込み処理
- `backend/internal/adapter/youtube/api.go` - YouTube API実装