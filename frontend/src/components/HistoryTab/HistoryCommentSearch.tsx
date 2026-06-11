import { useState, useMemo } from 'react'
import type { Comment } from '../../utils/api'

interface HistoryCommentSearchProps {
  comments: Comment[]
}

export function HistoryCommentSearch({ comments }: HistoryCommentSearchProps) {
  const [keyword, setKeyword] = useState('')

  const filtered = useMemo(() => {
    const trimmed = keyword.trim()
    if (!trimmed) return comments
    const lower = trimmed.toLowerCase()
    return comments.filter(
      (c) => c.message.toLowerCase().includes(lower) || c.displayName.toLowerCase().includes(lower),
    )
  }, [comments, keyword])

  const formatDate = (iso: string): string => {
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  return (
    <section
      style={{
        border: '1px solid var(--c-line-strong)',
        background: 'var(--c-bg-2)',
        padding: '20px 24px',
      }}
      className="space-y-3"
    >
      <h3
        style={{
          fontFamily: 'var(--f-mono)',
          fontSize: '11px',
          letterSpacing: '0.2em',
          textTransform: 'uppercase',
          color: 'var(--c-accent-2)',
        }}
      >
        コメント検索
      </h3>
      <input
        type="text"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        placeholder="キーワードで絞り込み"
        aria-label="コメント検索キーワード"
        style={{
          width: '100%',
          padding: '9px 12px',
          background: 'var(--c-bg)',
          border: '1px solid var(--c-line-strong)',
          color: 'var(--c-ink)',
          fontFamily: 'var(--f-mono)',
          fontSize: '13px',
        }}
      />
      <div
        style={{
          fontFamily: 'var(--f-mono)',
          fontSize: '11px',
          letterSpacing: '0.1em',
          color: 'var(--c-ink-mute)',
        }}
      >
        {filtered.length} / {comments.length} 件
      </div>
      {filtered.length === 0 ? (
        <p
          style={{
            padding: '16px 0',
            textAlign: 'center',
            fontFamily: 'var(--f-mono)',
            fontSize: '12px',
            color: 'var(--c-ink-mute)',
          }}
        >
          コメントがありません
        </p>
      ) : (
        <ul
          style={{
            maxHeight: '384px',
            overflowY: 'auto',
          }}
        >
          {filtered.map((c) => (
            <li
              key={c.id}
              style={{
                padding: '10px 0',
                borderBottom: '1px solid var(--c-line)',
              }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: '8px',
                  marginBottom: '2px',
                }}
              >
                <span
                  style={{
                    fontSize: '13px',
                    fontWeight: 500,
                    color: 'var(--c-ink)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {c.displayName}
                </span>
                <span
                  style={{
                    fontFamily: 'var(--f-mono)',
                    fontSize: '11px',
                    color: 'var(--c-ink-mute)',
                    flexShrink: 0,
                  }}
                >
                  {formatDate(c.publishedAt)}
                </span>
              </div>
              <p style={{ fontSize: '13px', color: 'var(--c-ink-dim)' }}>{c.message}</p>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
