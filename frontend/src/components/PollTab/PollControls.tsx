import { useState } from 'react'
import { LoadingButton } from '../LoadingButton'

interface PollControlsProps {
  keywords: string[]
  onAddKeyword: (word: string) => void
  onRemoveKeyword: (word: string) => void
  onClear: () => void
  onRecount: () => void
  isLoading: boolean
  lastUpdated: string
}

export function PollControls({
  keywords,
  onAddKeyword,
  onRemoveKeyword,
  onClear,
  onRecount,
  isLoading,
  lastUpdated,
}: PollControlsProps) {
  const [input, setInput] = useState('')

  const handleAdd = () => {
    if (input.trim()) {
      onAddKeyword(input)
      setInput('')
    }
  }

  return (
    <div className="space-y-4">
      <section className="rounded-lg shadow-subtle ring-1 ring-black/5 bg-white/80 backdrop-blur p-5">
        <h2 className="text-sm font-semibold mb-3 text-slate-700">
          投票キーワード（完全一致でカウント）
        </h2>

        <p className="text-[12px] text-slate-500 mb-3">
          キーワードを 1 つずつ追加してください。コメントが完全一致した場合のみ 1
          票としてカウントされます。
        </p>

        <div className="flex gap-2 mb-4">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="投票キーワードを入力"
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
          {keywords.length > 0 && (
            <LoadingButton onClick={onClear} disabled={isLoading} variant="outline">
              クリア
            </LoadingButton>
          )}
        </div>

        <div className="flex flex-wrap gap-2">
          {keywords.length === 0 && (
            <span className="text-[12px] text-slate-500">
              キーワード未設定。追加すると一覧表示されます。
            </span>
          )}
          {keywords.map((word) => (
            <span
              key={word}
              className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-slate-200 text-sm"
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

      <div className="flex items-center justify-between">
        <LoadingButton
          onClick={onRecount}
          isLoading={isLoading}
          loadingText="集計中..."
          variant="primary"
          disabled={keywords.length === 0}
        >
          今すぐ集計
        </LoadingButton>
        <div className="text-[12px] text-slate-500">最終更新: {lastUpdated}</div>
      </div>
    </div>
  )
}
