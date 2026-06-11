import { useMemo, useState, useCallback } from 'react'
import type { User } from '../../utils/api'
import { sortUsersStable } from '../../utils/sortUsers'
import { Tooltip } from '../Tooltip'
import { isJapaneseTextTooLong, truncateJapaneseText } from '../../utils/textUtils'

interface HistoryUserTableProps {
  users: User[]
}

function CopyLinkButton({ url, displayName }: { url: string; displayName: string }) {
  const [copied, setCopied] = useState(false)
  const copyText = `${displayName}さん ${url}`

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(copyText)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      try {
        const textarea = document.createElement('textarea')
        textarea.value = copyText
        textarea.style.position = 'fixed'
        textarea.style.opacity = '0'
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      } catch {
        // コピー失敗は無視
      }
    }
  }, [copyText])

  return (
    <button
      onClick={handleCopy}
      title="チャンネルURLをコピー"
      aria-label="チャンネルURLをコピー"
      style={{
        flexShrink: 0,
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        color: copied ? 'var(--c-success)' : 'var(--c-ink-mute)',
        transition: 'color 0.2s',
        padding: '0 2px',
      }}
    >
      {copied ? (
        <svg
          className="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        <svg
          className="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
          />
        </svg>
      )}
    </button>
  )
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

export function HistoryUserTable({ users }: HistoryUserTableProps) {
  const sorted = useMemo(() => sortUsersStable(users), [users])

  const formatDate = (iso: string | undefined): string => {
    if (!iso) return '--'
    const d = new Date(iso)
    if (isNaN(d.getTime())) return '--'
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  if (sorted.length === 0) {
    return (
      <p
        style={{
          padding: '20px',
          textAlign: 'center',
          fontFamily: 'var(--f-mono)',
          fontSize: '12px',
          color: 'var(--c-ink-mute)',
        }}
      >
        視聴者データがありません
      </p>
    )
  }

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
            <th style={{ ...thStyle, width: '60px' }}>#</th>
            <th style={thStyle}>名前</th>
            <th style={{ ...thStyle, width: '100px' }}>発言数</th>
            <th style={{ ...thStyle, width: '160px' }} className="hidden md:table-cell">
              初回コメント
            </th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((user, i) => {
            const channelUrl = user.channelId
              ? `https://www.youtube.com/channel/${user.channelId}`
              : null
            const name = user.displayName || user.channelId || 'Unknown'
            const rowBg = i % 2 === 0 ? 'var(--c-bg)' : 'var(--c-bg-2)'
            return (
              <tr
                key={`${user.channelId || user.displayName}-${i}`}
                style={{
                  borderBottom: '1px solid var(--c-line)',
                  background: rowBg,
                  transition: 'background 0.15s',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLTableRowElement).style.background =
                    'rgba(0,108,138,0.06)'
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLTableRowElement).style.background = rowBg
                }}
              >
                <td
                  style={{
                    padding: '10px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: 'var(--c-ink-mute)',
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {String(i + 1).padStart(2, '0')}
                </td>
                <td style={{ padding: '10px 16px', color: 'var(--c-ink)', fontWeight: 500 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <Tooltip
                      content={name}
                      disabled={!isJapaneseTextTooLong(name, 30)}
                      className="block min-w-0 flex-1 max-w-[280px]"
                    >
                      <span className="block truncate">
                        {isJapaneseTextTooLong(name, 30) ? truncateJapaneseText(name, 30) : name}
                      </span>
                    </Tooltip>
                    {channelUrl && <CopyLinkButton url={channelUrl} displayName={name} />}
                  </div>
                </td>
                <td
                  style={{
                    padding: '10px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: 'var(--c-ink-dim)',
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {user.commentCount ?? 0}
                </td>
                <td
                  className="hidden md:table-cell"
                  style={{
                    padding: '10px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: 'var(--c-ink-dim)',
                    textAlign: 'center',
                  }}
                >
                  {formatDate(user.firstCommentedAt)}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </section>
  )
}
