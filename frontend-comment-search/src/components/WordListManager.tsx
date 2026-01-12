import { useState } from 'react'
import { LoadingButton } from './LoadingButton'

interface WordListManagerProps {
  keywords: string[]
  onAdd: (word: string) => void
  onRemove: (word: string) => void
  disabled?: boolean
}

export function WordListManager({ keywords, onAdd, onRemove, disabled }: WordListManagerProps) {
  const [input, setInput] = useState('')

  const handleAdd = () => {
    if (input.trim()) {
      onAdd(input)
      setInput('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleAdd()
    }
  }

  return (
    <section className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur p-5">
      <h2 className="text-sm font-semibold mb-3 text-slate-700 dark:text-slate-200">
        検索キーワード（OR検索）
      </h2>

      <div className="flex gap-2 mb-4">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="キーワードを入力"
          disabled={disabled}
          className="flex-1 px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 focus:outline-none focus:ring-2 focus:ring-neutral-400/60 text-[14px]"
        />
        <LoadingButton onClick={handleAdd} disabled={disabled || !input.trim()} variant="primary">
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
              onClick={() => onRemove(word)}
              disabled={disabled}
              className="hover:text-red-500 disabled:opacity-50"
              aria-label={`${word}を削除`}
            >
              ×
            </button>
          </span>
        ))}
      </div>
    </section>
  )
}
