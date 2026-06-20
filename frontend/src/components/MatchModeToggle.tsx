import type { MatchMode } from '../utils/countVotes'

interface MatchModeToggleProps {
  matchMode: MatchMode
  onMatchModeChange: (mode: MatchMode) => void
  disabled?: boolean
}

export function MatchModeToggle({
  matchMode,
  onMatchModeChange,
  disabled = false,
}: MatchModeToggleProps) {
  return (
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
          type="button"
          onClick={() => onMatchModeChange(mode)}
          disabled={disabled}
          aria-pressed={matchMode === mode}
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.12em',
            padding: '4px 12px',
            border: '1px solid var(--c-line-strong)',
            cursor: disabled ? 'not-allowed' : 'pointer',
            background: matchMode === mode ? 'var(--c-ink)' : 'transparent',
            color: matchMode === mode ? '#fff' : 'var(--c-ink-dim)',
            transition: 'background 0.15s, color 0.15s',
          }}
        >
          {mode === 'exact' ? '完全一致' : '部分一致'}
        </button>
      ))}
    </div>
  )
}
