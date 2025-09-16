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
            <div className="space-y-3 text-sm text-slate-600 dark:text-slate-300">
              <div>
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200">
                  <span className="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>
                  基本の使い方
                </span>
                <p className="mt-1 ml-3">YouTube動画のURLまたはvideoIdを下の入力欄に貼り付けて「切替」ボタンをクリックしてください。</p>
              </div>
              <div className="grid gap-2 md:grid-cols-3 ml-3">
                <div className="flex items-start gap-2">
                  <span className="text-blue-600 dark:text-blue-400 font-medium">切替:</span>
                  <span className="text-xs">指定した動画の監視を開始</span>
                </div>
                <div className="flex items-start gap-2">
                  <span className="text-green-600 dark:text-green-400 font-medium">今すぐ取得:</span>
                  <span className="text-xs">手動でコメントを取得</span>
                </div>
                <div className="flex items-start gap-2">
                  <span className="text-amber-600 dark:text-amber-400 font-medium">リセット:</span>
                  <span className="text-xs">参加者リストをクリア</span>
                </div>
              </div>
              <div className="ml-3">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  💡 コツ:
                </span>
                <span className="text-xs ml-1">配信開始前にアプリを起動すると、より多くの参加者を取得できます</span>
              </div>
              <div className="ml-3 pt-2 border-t border-slate-200/40 dark:border-slate-600/30">
                <span className="inline-flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-200 text-xs">
                  🔄 配信終了後:
                </span>
                <div className="text-xs mt-1 space-y-1">
                  <p>• 配信終了は<span className="font-medium text-slate-700 dark:text-slate-200">自動検知</span>され、状態が「WAITING」に戻ります</p>
                  <p>• 参加者リストは<span className="font-medium text-slate-700 dark:text-slate-200">自動的にクリア</span>されます</p>
                  <p>• 新しい配信を始める場合は、新しいvideoIdを入力して「切替」してください</p>
                  <p>• 手動でリセットしたい場合は「リセット」ボタンをお使いください</p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  )
}