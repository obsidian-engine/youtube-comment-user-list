📗 Frontend ドキュメント（仕様書・設計書・手順書）

✅1. 概要

YouTubeライブのコメント参加ユーザー一覧を表示するフロントアプリ。
• 技術: React 18 + Vite + Tailwind CSS
• ホスティング: Vercel
• バックエンド: Cloud Run (Go API) と通信
• 目的: 配信中にユーザーを蓄積し、配信終了時点で全員が表示されていればOK。CSV/コピー/履歴保存はしない。

⸻

✅2. 仕様書

🎯 機能要件
• videoId入力→切替（必須）
• 入力欄で videoId を指定し、[切替] を押すとバックエンド /switch-video を呼ぶ。
• 成功時に内部状態をACTIVEへ、以降の取得対象が更新される。
• 一覧表示
• チップ（デフォルト） と 表 の2モード。
• 長い表示名は 1行省略（ellipsis）。
• フィルタ（部分一致）。
• 更新
• 一定間隔（既定30秒）で自動更新（/status と /users.json を取得）。
• 手動取得ボタンで /pull → 直後にrefresh()。
• 終了時の自動反映
• バックエンド側が終了を自動検知し WAITING に戻す。
• フロントは state==='WAITING' を受けたら待機表示に戻す。任意でバナーを出す。
• 不要機能
• CSV/コピー機能なし。履歴保存なし。OBS連携や読み上げ連携なし。

🔌 バックエンドAPI（参照）

メソッド	パス	用途
GET	/status	状態（WAITING/ACTIVE、人数、videoId、時刻）
GET	/users.json	表示名の配列
POST	/switch-video	{"videoId": "<id>"} で対象切替
POST	/pull	即時取得（自動リセットもここで反映可能）
POST	/reset	手動初期化


⸻

✅3. 設計書

🧱 画面構成（単一ページ）
• ヘッダー: タイトル、状態バッジ（WAITING/ACTIVE）、人数、開始/終了/最終更新時刻（表示は最終更新のみでも可）
• 操作バー:
  • videoId 入力 + [切替]
  • 自動間隔選択（停止/3/5/10/30/60秒）
  • [今すぐ取得], [リセット]
  • フィルタ（部分一致）
  • 表示モード選択（チップ/表）
• コンテンツ:
  • チップモード: grid + minmax(180px,1fr) のレスポンシブ
  • 表モード: # + 名前の2列

🗂 ディレクトリ

frontend/
├─ index.html
├─ package.json
├─ vite.config.js
├─ tailwind.config.js
├─ postcss.config.js
├─ .env            # VITE_BACKEND_URL を設定
└─ src/
   ├─ App.jsx
   ├─ main.jsx
   ├─ index.css
   ├─ hooks/       # useAutoRefresh など（任意）
   ├─ utils/       # fetchヘルパ（任意）
   ├─ mocks/       # MSW
   │  ├─ handlers.ts
   │  └─ setup.ts
   └─ __tests__/   # Vitest + RTL + MSW
      ├─ App.ui.spec.tsx
      ├─ App.integration.spec.tsx
      └─ Layout.spec.tsx

🧩 状態管理（最小）
• state: 'WAITING' | 'ACTIVE'
• users: string[]（displayName）
• mode: 'chip' | 'table'
• filter: string
• intervalSec: number（0=停止, 既定30）
• lastUpdated: string（ローカル表記）
• videoId: string（localStorage に保存）

🎨 スタイル（崩れ防止）
• Tailwind ベース。
• .truncate-1（1行省略）をユーザー名に適用。
• チップは grid-cols-[repeat(auto-fill,minmax(180px,1fr))] 相当（tailwind.config.js拡張でもOK）。
• 文字色/背景/アクセントはトークン化（前に提示したカラーパレットを推奨）。

⸻

✅4. 手順書（TDDパターン）

⚙️ セットアップ

```bash
cd frontend
npm i
npm i -D vitest @testing-library/react @testing-library/jest-dom msw whatwg-fetch
echo 'VITE_BACKEND_URL=http://localhost:8080' > .env   # ローカル検証用
npm run dev
```

🧪 ユニット（Red → Green → Refactor）

目的: UIの基本挙動を固定化し、改修時の退行を防ぐ
1. 状態バッジ（App.ui.spec.tsx）
   • Red: state='WAITING'/'ACTIVE' で色/文言が一致するか
   • Green: バッジ実装（Tailwindクラス分岐）
   • Refactor: 小コンポーネント化（任意）
2. 表示モード切替（chip/ table）
   • Red: セレクト変更でDOM構造が切り替わるか
   • Green: 条件分岐描画
   • Refactor: 表コンポーネント分離（任意）
3. フィルタ
   • Red: 「a」入力で ["Alice","あいう"] が表示、["Bob"] は非表示
   • Green: toLowerCase() + includes()
   • Refactor: メモ化（useMemo）
4. videoId入力→localStorage保存
   • Red: 入力→再マウントで初期値復元
   • Green: localStorage 実装
   • Refactor: useEffect/初期化を関数抽出

🔗 結合（UI × API：MSW）

目的: バックエンドと切れた状態でもフローを検証
5. 初期ロード（App.integration.spec.tsx）
   • Red: /status=WAITING, /users.json=[] → 待機UI
   • Green: refresh() 実装（useEffect 初回 + interval）
6. 切替成功
   • Red: POST /switch-video 200 → 次リフレッシュで state=ACTIVE
   • Green: switchVideo() 実装（成功時に refresh()）
7. 今すぐ取得
   • Red: POST /pull 200 → 即 refresh() で人数が増えている
   • Green: pullNow() 実装
8. 切替失敗
   • Red: POST /switch-video 502 → エラー表示（alert か バナー）
   • Green: エラー処理追加
9. 終了の自動リセット反映
   • Red: POST /pull 後、MSWで /status=WAITING を返す → UIが待機に戻る
   • Green: refresh() の表示更新（stateとusers）

⏱ 自動更新（タイマー）
10. 30s発火／停止（fake timers）

   • Red: intervalSec=30 で refresh が定期実行、0 にすると止まる
   • Green: useEffect と clearInterval
   • Refactor: useAutoRefresh() に抽出可

🎨 レイアウト保護
11. 長文/emoji（Layout.spec.tsx）

   • Red: 100文字＋絵文字で ellipsis が効く（getComputedStyle or snapshot）
   • Green: .truncate-1 を適用
   • Refactor: チップ最小幅を定数化

⸻

✅5. デプロイ（Vercel）

手順
1. Vercel サインアップ → GitHub 連携
2. Import Project:
   • Framework: Vite（自動認識）
   • Root Directory: frontend/（※モノレポなので重要）
   • Build: npm run build / Output: dist（自動）
3. Environment Variables:
   • VITE_BACKEND_URL = https://<Cloud Run サービスURL>
4. Deploy → https://<project>.vercel.app 発行
5. Cloud Run 側の FRONTEND_ORIGIN を 発行URLに締める（CORS最終化）

⸻

✅6. 受け入れ条件（Definition of Done）
• UIから videoId入力→[切替] で ACTIVE表示になる
• 今すぐ取得で人数が即時更新される
• 配信終了時、次回 refresh で WAITINGに戻り、一覧が空になる
• CSV/コピーUIが存在しない
• 30s自動更新が動作し、0で停止できる
• npm run test（Vitest）が 緑
• 本番で CORSが本番Originのみに制限されている

⸻

✅7. よくある詰まり & 対処
• CORSエラー: Cloud Run の FRONTEND_ORIGIN が Vercel の本番URLに締められているか確認
• Root Dirミス: Vercel インポートで frontend/ を指定したか
• 環境変数: VITE_BACKEND_URL の末尾スラ不要・https 必須
• 切替失敗: 配信が未開始だと activeLiveChatId が取れない → 配信開始後に再トライ

⸻

✅8. 追加メモ（任意拡張）
• UIバナー: state==='WAITING' && users.length===0 で「配信が終了しました。次の videoId を入力して『切替』してください。」
• アクセシビリティ: ボタンに aria-label、フォームに label 紐付け
• パフォーマンス: 大量ユーザー時は virtual list 検討（今回の要件では不要想定）

⸻

以上が フロントエンドの最終ドキュメントです。
必要なら TypeScript版の雛形や、Vitest/MSW のサンプルテストファイルもすぐ出します。
