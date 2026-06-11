import type { Comment } from '../../utils/api'

interface CommentRowProps {
  comment: Comment
  index: number
  isChecked: boolean
  onToggle: () => void
}

export function CommentRow({ comment, index, isChecked, onToggle }: CommentRowProps) {
  const time = new Date(comment.publishedAt).toLocaleTimeString('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
  })

  const rowBg = isChecked
    ? 'rgba(45, 122, 63, 0.06)'
    : index % 2 === 0
      ? 'var(--c-bg)'
      : 'var(--c-bg-2)'

  return (
    <tr
      style={{
        borderBottom: '1px solid var(--c-line)',
        background: rowBg,
        opacity: isChecked ? 0.55 : 1,
        transition: 'background 0.15s',
      }}
    >
      <td style={{ padding: '10px 16px', textAlign: 'center' }}>
        <input
          type="checkbox"
          checked={isChecked}
          onChange={onToggle}
          style={{
            width: '16px',
            height: '16px',
            cursor: 'pointer',
            accentColor: 'var(--c-accent-2)',
          }}
          aria-label="読み上げ済み"
        />
      </td>
      <td
        style={{
          padding: '10px 16px',
          fontFamily: 'var(--f-mono)',
          fontSize: '12px',
          color: 'var(--c-ink-mute)',
          textAlign: 'center',
        }}
      >
        {time}
      </td>
      <td
        style={{
          padding: '10px 16px',
          fontSize: '13px',
          color: 'var(--c-ink)',
          fontWeight: 500,
        }}
      >
        {comment.displayName}
      </td>
      <td
        style={{
          padding: '10px 16px',
          fontSize: '13px',
          color: 'var(--c-ink-dim)',
        }}
      >
        {comment.message}
      </td>
    </tr>
  )
}
