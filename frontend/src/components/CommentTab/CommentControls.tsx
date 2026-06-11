import { useState } from 'react'
import { LoadingButton } from '../LoadingButton'
import { SELECT_CLASS } from '../../utils/styles'

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
      <section
        style={{
          background: 'var(--c-bg-2)',
          border: '1px solid var(--c-line-strong)',
          padding: '20px 24px',
        }}
      >
        <h2
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.2em',
            textTransform: 'uppercase',
            color: 'var(--c-accent-2)',
            marginBottom: '14px',
          }}
        >
          検索キーワード（OR検索）
        </h2>

        <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
            placeholder="キーワードを入力"
            disabled={isLoading}
            style={{
              flex: 1,
              padding: '9px 12px',
              background: 'var(--c-bg)',
              border: '1px solid var(--c-line-strong)',
              color: 'var(--c-ink)',
              fontFamily: 'var(--f-mono)',
              fontSize: '13px',
              outline: 'none',
            }}
          />
          <LoadingButton
            onClick={handleAdd}
            disabled={isLoading || !input.trim()}
            variant="primary"
          >
            追加
          </LoadingButton>
        </div>

        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
          {keywords.map((word) => (
            <span
              key={word}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                padding: '4px 10px',
                background: 'var(--c-ink)',
                color: '#fff',
                fontFamily: 'var(--f-mono)',
                fontSize: '12px',
                letterSpacing: '0.08em',
              }}
            >
              {word}
              <button
                onClick={() => onRemoveKeyword(word)}
                disabled={isLoading}
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'rgba(255,255,255,0.6)',
                  cursor: 'pointer',
                  padding: '0 2px',
                  fontSize: '14px',
                  lineHeight: 1,
                }}
                aria-label={`${word}を削除`}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      </section>

      {/* コントロールバー */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: '12px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
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
            className={SELECT_CLASS}
          >
            <option value="0">停止</option>
            <option value="60">60s</option>
            <option value="90">90s</option>
            <option value="120">120s</option>
          </select>
        </div>
        <div
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.1em',
            color: 'var(--c-ink-mute)',
          }}
        >
          {commentsCount}件中 {checkedCount}件済 | 最終更新: {lastUpdated}
        </div>
      </div>
    </div>
  )
}
