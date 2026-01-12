import { useState } from 'react'
import { useSearchState } from './hooks/useSearchState'
import { useCheckState } from './hooks/useCheckState'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { WordListManager } from './components/WordListManager'
import { CommentList } from './components/CommentList'
import { ThemeToggle } from './components/ThemeToggle'
import { LoadingButton } from './components/LoadingButton'

export default function App() {
  const search = useSearchState()
  const check = useCheckState()
  const [intervalSec, setIntervalSec] = useState(10)

  useAutoRefresh(intervalSec, search.search)

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <div className="fixed inset-0 -z-10 bg-field" />
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-xl font-bold">コメント検索</h1>
          <ThemeToggle />
        </div>

        {search.errorMsg && (
          <div
            role="alert"
            aria-live="assertive"
            className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3"
          >
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
            <LoadingButton
              onClick={search.search}
              isLoading={search.isLoading}
              loadingText="検索中..."
              variant="primary"
            >
              今すぐ検索
            </LoadingButton>
            <LoadingButton onClick={check.clear} variant="outline">
              チェックをリセット
            </LoadingButton>
            <select
              aria-label="自動更新間隔"
              value={intervalSec}
              onChange={(e) => setIntervalSec(Number(e.target.value))}
              disabled={search.isLoading}
              className="text-[12px] px-2 py-1 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10"
            >
              <option value="0">停止</option>
              <option value="10">10s</option>
              <option value="15">15s</option>
              <option value="30">30s</option>
            </select>
          </div>
          <div className="text-[12px] text-slate-500 dark:text-slate-400">
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
