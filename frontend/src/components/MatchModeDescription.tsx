import type { MatchMode } from '../utils/countVotes'

interface MatchModeDescriptionProps {
  matchMode: MatchMode
  variant: 'poll' | 'history'
}

const descriptions: Record<'poll' | 'history', Record<MatchMode, string>> = {
  poll: {
    exact:
      'キーワードを 1 つずつ追加してください。コメントが完全一致した場合のみ 1 票としてカウントされます。',
    partial:
      'キーワードを 1 つずつ追加してください。コメントにキーワードが含まれる場合に 1 票としてカウントされます。',
  },
  history: {
    exact: 'コメントがキーワードと完全一致した場合に 1 票としてカウントします。',
    partial: 'コメントにキーワードが含まれる場合に 1 票としてカウントします。',
  },
}

export function MatchModeDescription({ matchMode, variant }: MatchModeDescriptionProps) {
  const text = descriptions[variant][matchMode]

  if (variant === 'poll') {
    return (
      <p
        style={{
          fontFamily: 'var(--f-mono)',
          fontSize: '11px',
          color: 'var(--c-ink-mute)',
          marginBottom: '14px',
          lineHeight: 1.6,
        }}
      >
        {text}
      </p>
    )
  }

  return <p className="text-[12px] text-slate-500">{text}</p>
}
