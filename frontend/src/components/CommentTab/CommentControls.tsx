import { useState } from 'react'
import { LoadingButton } from '../LoadingButton'

interface CommentControlsProps {
  keywords: string[]
  onAddKeyword: (word: string) => void
  onRemoveKeyword: (word: string) => void
  onSearch: () => void
  onReset: () => void
  isLoading: boolean
  intervalSec: number
  setIntervalSec: (value: number) => void
  commentsCount: number
  checkedCount: number
  lastUpdated: string
}

export function CommentControls({
  keywords,
  onAddKeyword,
  onRemoveKeyword,
  onSearch,
  onReset,
  isLoading,
  intervalSec,
  setIntervalSec,
  commentsCount,
  checkedCount,
  lastUpdated,
}: CommentControlsProps) {
  const [input, setInput] = useState('')

  const handleAdd = () => {
    if (input.trim()) {
      onAddKeyword(input)
      setInput('')
    }
  }



  return (
    <div className="space-y-4">
      {/* キーワード管理 */}
      <section className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur p-5">
        <h2 className="text-sm font-semibold mb-3 text-slate-700 dark:text-slate-200">
          検索キーワード（OR検索）
        </h2>

        <div className="flex gap-2 mb-4">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="キーワードを入力"
            disabled={isLoading}
            className="flex-1 px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 focus:outline-none focus:ring-2 focus:ring-neutral-400/60 text-[14px]"
          />
          <LoadingButton
            onClick={handleAdd}
            disabled={isLoading || !input.trim()}
            variant="primary"
          >
            追加
          </LoadingButton>
        </div>

        <div className="flex flex-wrap gap-2">
          {keywords.map((word) => (
            <span
              key={word}
              className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-slate-200 dark:bg-slate-700 text-sm"
            >
              {word}
              <button
                onClick={() => onRemoveKeyword(word)}
                disabled={isLoading}
                className="hover:text-red-500 disabled:opacity-50"
                aria-label={`${word}を削除`}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      </section>

      {/* コントロールバー */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <LoadingButton
            onClick={onSearch}
            isLoading={isLoading}
            loadingText="検索中..."
            variant="primary"
          >
            今すぐ検索
          </LoadingButton>
          <LoadingButton onClick={onReset} variant="outline">
            リセット
          </LoadingButton>
          <select
            aria-label="自動更新間隔"
            value={intervalSec}
            onChange={(e) => setIntervalSec(Number(e.target.value))}
            disabled={isLoading}
            className="text-[12px] px-2 py-1 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10"
          >
            <option value="0">停止</option>
            <option value="10">10s</option>
            <option value="15">15s</option>
            <option value="30">30s</option>
          </select>
        </div>
        <div className="text-[12px] text-slate-500 dark:text-slate-400">
          {commentsCount}件中 {checkedCount}件済 | 最終更新: {lastUpdated}
        </div>
      </div>
    </div>
  )
}
