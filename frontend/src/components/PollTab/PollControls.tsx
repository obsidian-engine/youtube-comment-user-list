import { useState } from 'react'
import { LoadingButton } from '../LoadingButton'
import type { MatchMode } from '../../utils/countVotes'

interface PollControlsProps {
  keywords: string[]
  matchMode: MatchMode
  onMatchModeChange: (mode: MatchMode) => void
  onAddKeyword: (word: string) => void
  onRemoveKeyword: (word: string) => void
  onClear: () => void
  onRecount: () => void
  isLoading: boolean
  lastUpdated: string
}

export function PollControls({
  keywords,
  matchMode,
  onMatchModeChange,
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
      <section className="card-editorial">
        <div className="eyebrow">
          TALLY
          <div className="eyebrow__rule" />
        </div>

        <div style={{ padding: '16px 20px 20px' }}>
          <h2
            style={{
              fontFamily: 'var(--f-mono)',
              fontSize: '11px',
              letterSpacing: '0.2em',
              textTransform: 'uppercase',
              color: 'var(--c-ink-dim)',
              marginBottom: '8px',
            }}
          >
            投票キーワード
          </h2>

          <div
            style={{
              display: 'flex',
              gap: '4px',
              marginBottom: '12px',
            }}
            role="group"
            aria-label="マッチモード"
          >
            {(['exact', 'partial'] as const).map((mode) => (
              <button
                key={mode}
                onClick={() => onMatchModeChange(mode)}
                disabled={isLoading}
                aria-pressed={matchMode === mode}
                style={{
                  fontFamily: 'var(--f-mono)',
                  fontSize: '11px',
                  letterSpacing: '0.12em',
                  padding: '4px 12px',
                  border: '1px solid var(--c-line-strong)',
                  cursor: isLoading ? 'not-allowed' : 'pointer',
                  background: matchMode === mode ? 'var(--c-ink)' : 'transparent',
                  color: matchMode === mode ? '#fff' : 'var(--c-ink-dim)',
                  transition: 'background 0.15s, color 0.15s',
                }}
              >
                {mode === 'exact' ? '完全一致' : '部分一致'}
              </button>
            ))}
          </div>

          <p
            style={{
              fontFamily: 'var(--f-mono)',
              fontSize: '11px',
              color: 'var(--c-ink-mute)',
              marginBottom: '14px',
              lineHeight: 1.6,
            }}
          >
            {matchMode === 'exact'
              ? 'キーワードを 1 つずつ追加してください。コメントが完全一致した場合のみ 1 票としてカウントされます。'
              : 'キーワードを 1 つずつ追加してください。コメントにキーワードが含まれる場合に 1 票としてカウントされます。'}
          </p>

          <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
              placeholder="投票キーワードを入力"
              aria-label="キーワード入力"
              disabled={isLoading}
              className="input-rule"
              style={{ flex: 1 }}
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

          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
            {keywords.length === 0 && (
              <span
                style={{
                  fontFamily: 'var(--f-mono)',
                  fontSize: '11px',
                  color: 'var(--c-ink-mute)',
                }}
              >
                キーワード未設定。追加すると一覧表示されます。
              </span>
            )}
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
        </div>
      </section>

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: '12px',
        }}
      >
        <LoadingButton
          onClick={onRecount}
          isLoading={isLoading}
          loadingText="集計中…"
          variant="primary"
          disabled={keywords.length === 0}
          title={keywords.length === 0 ? 'キーワードを追加すると有効になります' : undefined}
          style={keywords.length === 0 ? { cursor: 'not-allowed' } : undefined}
        >
          今すぐ集計
        </LoadingButton>
        <div
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.1em',
            color: 'var(--c-ink-mute)',
          }}
        >
          最終更新: {lastUpdated}
        </div>
      </div>
    </div>
  )
}
