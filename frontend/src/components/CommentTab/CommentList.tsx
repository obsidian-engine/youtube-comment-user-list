import { CommentRow } from './CommentRow'
import type { Comment } from '../../utils/api'

interface CommentListProps {
  comments: Comment[]
  isChecked: (id: string) => boolean
  onToggle: (id: string) => void
  isLoading: boolean
}

const thStyle: React.CSSProperties = {
  padding: '12px 16px',
  fontFamily: 'var(--f-mono)',
  fontWeight: 700,
  fontSize: '11px',
  letterSpacing: '0.18em',
  textTransform: 'uppercase',
  color: '#fff',
  background: 'var(--c-ink)',
  textAlign: 'center',
}

export function CommentList({ comments, isChecked, onToggle, isLoading }: CommentListProps) {
  return (
    <section
      style={{
        overflow: 'hidden',
        border: '1px solid var(--c-line-strong)',
        background: 'var(--c-bg-2)',
      }}
    >
      <table className="w-full table-fixed" style={{ fontSize: '14px', lineHeight: 1.7 }}>
        <thead>
          <tr>
            <th style={{ ...thStyle, width: '80px' }}>済</th>
            <th style={{ ...thStyle, width: '100px' }}>時刻</th>
            <th style={{ ...thStyle, width: '150px' }}>投稿者</th>
            <th style={{ ...thStyle, textAlign: 'left' }}>コメント</th>
          </tr>
        </thead>
        <tbody>
          {comments.map((comment, i) => (
            <CommentRow
              key={comment.id}
              comment={comment}
              index={i}
              isChecked={isChecked(comment.id)}
              onToggle={() => onToggle(comment.id)}
            />
          ))}
        </tbody>
      </table>

      {comments.length === 0 && !isLoading && (
        <p
          style={{
            padding: '20px 16px',
            fontFamily: 'var(--f-mono)',
            fontSize: '12px',
            color: 'var(--c-ink-mute)',
          }}
        >
          該当するコメントがありません。
        </p>
      )}

      {isLoading && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '32px',
          }}
        >
          <div
            style={{
              width: '28px',
              height: '28px',
              border: '2px solid var(--c-line-strong)',
              borderTopColor: 'var(--c-ink)',
              borderRadius: '50%',
              animation: 'spin 0.7s linear infinite',
            }}
          />
        </div>
      )}
    </section>
  )
}
