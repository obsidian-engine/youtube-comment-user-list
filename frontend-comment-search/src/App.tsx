import { useState } from 'react'
import { useSearchState } from './hooks/useSearchState'
import { useCheckState } from './hooks/useCheckState'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { WordListManager } from './components/WordListManager'
import { CommentList } from './components/CommentList'

export default function App() {
  const search = useSearchState()
  const check = useCheckState()
  const [intervalSec, setIntervalSec] = useState(10)

  useAutoRefresh(intervalSec, search.search)

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6">
        <h1 className="text-xl font-bold">コメント検索</h1>

        {search.errorMsg && (
          <div className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3">
            {search.errorMsg}
          </div>
        )}

        <WordListManager
          keywords={search.keywords}
          onAdd={search.addKeyword}
          onRemove={search.removeKeyword}
          disabled={search.isLoading}
        />

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={search.search}
              disabled={search.isLoading}
              className="px-4 py-2 rounded-md bg-slate-600 text-white hover:bg-slate-700 disabled:opacity-50"
            >
              今すぐ検索
            </button>
            <button
              onClick={check.clear}
              className="px-4 py-2 rounded-md bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600"
            >
              チェックをリセット
            </button>
            <select
              value={intervalSec}
              onChange={(e) => setIntervalSec(Number(e.target.value))}
              className="px-2 py-1 rounded-md border"
            >
              <option value="0">自動更新: 停止</option>
              <option value="10">10秒</option>
              <option value="15">15秒</option>
              <option value="30">30秒</option>
            </select>
          </div>
          <div className="text-sm text-slate-500">
            {search.comments.length}件中 {check.checkedCount}件済 | 最終更新: {search.lastUpdated}
          </div>
        </div>

        <CommentList
          comments={search.comments}
          isChecked={check.isChecked}
          onToggle={check.toggle}
          isLoading={search.isLoading}
        />
      </main>
    </div>
  )
}
