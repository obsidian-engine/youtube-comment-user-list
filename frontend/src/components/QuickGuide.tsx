import { useState } from 'react'

export function QuickGuide() {
  const [guideExpanded, setGuideExpanded] = useState(false)

  return (
    <section className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur" aria-label="操作ガイド">
      <div className="p-4 md:p-5">
        <button
          onClick={() => setGuideExpanded(!guideExpanded)}
          className="flex items-center gap-2 text-sm font-medium text-slate-700 dark:text-slate-200 hover:text-slate-900 dark:hover:text-white transition-colors"
          aria-expanded={guideExpanded}
          aria-controls="operation-guide"
        >
          <svg
            className={`w-4 h-4 transition-transform duration-200 ${guideExpanded ? 'rotate-90' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
          <span>はじめての方へ - 操作ガイド</span>
        </button>

{guideExpanded && (
          <div id="operation-guide" className="mt-4 pt-4 border-t border-slate-200/60 dark:border-slate-600/40" role="region">
            <div className="space-y-4 text-sm text-slate-600 dark:text-slate-300">
              
              {/* 基本の使い方 */}
              <div>
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200">
                  <span className="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>
                  基本の使い方
                </span>
                <p className="mt-1 ml-3">YouTube Live配信のURLまたはvideoIdを入力欄に貼り付けて「切替」ボタンをクリックしてください。</p>
              </div>
              
              {/* ボタン説明 */}
              <div className="grid gap-2 md:grid-cols-3 ml-3">
                <div className="flex items-start gap-2">
                  <span className="text-blue-600 dark:text-blue-400 font-medium">切替:</span>
                  <span className="text-xs">指定した配信の監視を開始</span>
                </div>
                <div className="flex items-start gap-2">
                  <span className="text-green-600 dark:text-green-400 font-medium">今すぐ取得:</span>
                  <span className="text-xs">手動でコメント取得実行</span>
                </div>
                <div className="flex items-start gap-2">
                  <span className="text-red-600 dark:text-red-400 font-medium">リセット:</span>
                  <span className="text-xs">ユーザーリストを全クリア</span>
                </div>
              </div>

              {/* 表示される情報 */}
              <div className="ml-3">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  📊 表示される情報:
                </span>
                <div className="text-xs mt-1 grid gap-1 md:grid-cols-2">
                  <p>• <strong>総ユーザー数:</strong> コメントした人の総数</p>
                  <p>• <strong>監視時間:</strong> 監視開始からの経過時間</p>
                  <p>• <strong>最新コメント:</strong> 最後のコメントからの経過時間</p>
                  <p>• <strong>ユーザーリスト:</strong> 名前・発言数・初回/最新コメント時間</p>
                </div>
              </div>

              {/* 自動更新機能 */}
              <div className="ml-3">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  🔄 自動更新機能:
                </span>
                <div className="text-xs mt-1 space-y-1">
                  <p>• 更新間隔を15秒〜60秒で設定可能（デフォルト15秒）</p>
                  <p>• 配信中は自動でコメントを取得し続けます</p>
                  <p>• 配信終了時は<span className="font-medium text-orange-600 dark:text-orange-400">自動で停止</span>します</p>
                </div>
              </div>

              {/* レスポンシブ対応 */}
              <div className="ml-3">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  📱 スマホ対応:
                </span>
                <p className="text-xs mt-1">狭い画面では#・名前・発言数のみ表示され、見やすくなります</p>
              </div>

              {/* 配信終了後の動作 */}
              <div className="ml-3 pt-2 border-t border-slate-200/40 dark:border-slate-600/30">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  ⚠️ 配信終了時の動作:
                </span>
                <div className="text-xs mt-1 space-y-1">
                  <p>• 配信終了は<span className="font-medium text-slate-700 dark:text-slate-200">自動検知</span>され、ステータスが「停止中」に戻ります</p>
                  <p>• 自動更新も同時に停止します</p>
                  <p>• ユーザーリストはそのまま残るので、配信データを確認できます</p>
                  <p>• 新しい配信を始める場合は、新しいvideoIdを入力して「切替」してください</p>
                  <p>• データをクリアしたい場合は「リセット」ボタンをお使いください</p>
                </div>
              </div>

              {/* コツとヒント */}
              <div className="ml-3 pt-2 border-t border-slate-200/40 dark:border-slate-600/30">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  💡 効果的な使い方:
                </span>
                <div className="text-xs mt-1 space-y-1">
                  <p>• 配信開始前にアプリを起動すると、初期参加者を逃さずキャッチできます</p>
                  <p>• ユーザーリストはソート可能（発言数・初回コメント時間）</p>
                  <p>• 長時間配信では更新間隔を30秒〜60秒に設定すると負荷軽減になります</p>
                </div>
              </div>
              
            </div>
          </div>
        )}
      </div>
    </section>
  )
}