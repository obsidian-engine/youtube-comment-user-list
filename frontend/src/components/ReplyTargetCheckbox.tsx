interface ReplyTargetCheckboxProps {
  channelId: string
  displayName: string
  checked: boolean
  onToggle: (channelId: string) => void
}

// 自動返信の対象にするかを切り替えるチェックボックス
export function ReplyTargetCheckbox({
  channelId,
  displayName,
  checked,
  onToggle,
}: ReplyTargetCheckboxProps) {
  return (
    <input
      type="checkbox"
      checked={checked}
      onChange={() => onToggle(channelId)}
      title="初めてのコメントに bot が自動で返信する"
      aria-label={`${displayName}さんを自動返信の対象にする`}
      style={{ flexShrink: 0, cursor: 'pointer', accentColor: 'var(--c-ink)' }}
    />
  )
}

export function ReplyTargetError({ message }: { message: string | null }) {
  if (!message) return null
  return (
    <p
      role="alert"
      style={{
        padding: '8px 16px',
        fontSize: '12px',
        color: 'var(--c-danger, #c0392b)',
      }}
    >
      {message}
    </p>
  )
}
