export function HelpPanel() {
  return (
    <section
      className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur p-5 md:p-6 space-y-6 text-sm text-slate-700 dark:text-slate-200"
      aria-label="ヘルプ"
    >
      <header>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">ヘルプ</h2>
        <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
          各タブの機能と操作を一覧で説明する。
        </p>
      </header>

      <article className="space-y-2">
        <h3 className="font-semibold text-slate-900 dark:text-slate-100">名前読み上げタブ</h3>
        <p>YouTube Live のチャットに参加したユーザーを一覧表示する。配信ごとの参加者把握用。</p>
        <ul className="list-disc pl-5 space-y-1 text-[13px]">
          <li>
            <strong>入力欄</strong>: YouTube Live の URL または videoId を貼り付ける。
          </li>
          <li>
            <strong>切替</strong>: 指定した配信の監視を開始する。実行と同時にログを初期化する。
          </li>
          <li>
            <strong>今すぐ取得</strong>: 自動更新を待たず手動でコメントを取得する。
          </li>
          <li>
            <strong>リセット</strong>: ユーザーリストを空にする。
          </li>
          <li>
            <strong>更新間隔</strong>: 60 / 90 / 120 秒から選択する。配信中のみ自動取得が走る。
          </li>
          <li>
            <strong>表示項目</strong>:
            通し番号、表示名、発言数、初回コメント時刻、最新コメント時刻。
            ヘッダクリックで発言数・初回時刻でソートできる。
          </li>
          <li>
            <strong>StatsCard</strong>: 総ユーザー数、監視開始からの経過時間、最終更新からの経過、
            スキップ数を表示する。
          </li>
          <li>
            <strong>配信終了検知</strong>:
            配信終了を検知すると自動更新を停止する。ユーザーリストは保持する。
          </li>
          <li>
            <strong>レスポンシブ</strong>: 狭い画面では番号・名前・発言数のみ表示する。
          </li>
        </ul>
      </article>

      <article className="space-y-2">
        <h3 className="font-semibold text-slate-900 dark:text-slate-100">コメント検索タブ</h3>
        <p>キーワードに一致するコメントだけを抽出し、確認・除外しながら読む。</p>
        <ul className="list-disc pl-5 space-y-1 text-[13px]">
          <li>
            <strong>キーワード追加</strong>: フォームに入力して送信するとチップとして追加される。
            複数キーワードは OR 検索。サーバ側は lowercase 部分一致で評価する。
          </li>
          <li>
            <strong>キーワード削除</strong>: チップの × で個別削除する。
          </li>
          <li>
            <strong>検索 / 自動更新</strong>: 手動検索ボタンと自動更新間隔の選択欄を持つ。
          </li>
          <li>
            <strong>リセット</strong>: 現在表示中の全コメントを非表示にする
            （キーワード自体は残る）。非表示は端末ローカルに保存する。
          </li>
          <li>
            <strong>チェック</strong>: 各コメントを既読印として check できる。チェック数は header
            に出る。
          </li>
          <li>
            <strong>表示項目</strong>: 表示名、コメント本文、投稿時刻（トグル可）。
          </li>
        </ul>
      </article>

      <article className="space-y-2">
        <h3 className="font-semibold text-slate-900 dark:text-slate-100">投票タブ</h3>
        <p>キーワードを選択肢にしてチャット投票を集計する。1 コメンター 1 票で確定。</p>
        <ul className="list-disc pl-5 space-y-1 text-[13px]">
          <li>
            <strong>キーワード追加 / 削除 / 全消去</strong>: コメント検索と同じ form 入力 UI
            を使う。 Enter キー単独でのキーワード追加は無効化されている（form submit 経由のみ）。
          </li>
          <li>
            <strong>集計ロジック</strong>: channelId ごとに publishedAt
            昇順で走査し、最初に出現したコメントが キーワードのいずれかと{' '}
            <strong>trim 後完全一致</strong> した場合にそのキーワードへ 1 票投ずる。 同一 channelId
            の以降のコメントは無視する。大文字小文字は厳密に区別する。
          </li>
          <li>
            <strong>自動更新</strong>: 投票タブ表示中かつキーワード設定済みの場合のみ、 15
            秒間隔で再集計する。
          </li>
          <li>
            <strong>再集計</strong>: ボタンで即時に再取得・再集計できる。
          </li>
          <li>
            <strong>表示項目</strong>:
            キーワードごとの票数、投票者リスト（表示名）、総票数、最終更新時刻。
          </li>
        </ul>
      </article>

      <article className="space-y-2">
        <h3 className="font-semibold text-slate-900 dark:text-slate-100">ログタブ</h3>
        <p>backend から届く処理ログを時系列で表示する。問題切り分け用。</p>
        <ul className="list-disc pl-5 space-y-1 text-[13px]">
          <li>
            <strong>表示内容</strong>: 「切替」や「今すぐ取得」実行時に backend が発行する
            進行状況・警告・エラーメッセージを表示する。
          </li>
          <li>
            <strong>クリア</strong>: ログをすべて消去する。
          </li>
          <li>
            <strong>注意</strong>: 現状は <code>/pull</code> 起点のログのみ届く。 他 endpoint
            のログ連携は今後対応予定。
          </li>
        </ul>
      </article>

      <article className="space-y-2">
        <h3 className="font-semibold text-slate-900 dark:text-slate-100">共通操作</h3>
        <ul className="list-disc pl-5 space-y-1 text-[13px]">
          <li>
            <strong>テーマ切替</strong>: 右上のトグルでライト / ダークを切替える。設定は保存される。
          </li>
          <li>
            <strong>タブ記憶</strong>: 最後に開いていたタブを次回起動時に復元する。
          </li>
          <li>
            <strong>エラー表示</strong>: 各タブ上部に赤色バナーで表示する。
          </li>
        </ul>
      </article>
    </section>
  )
}
