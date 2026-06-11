import { useState, useMemo } from 'react'
import type { HistorySummary } from '../../utils/api'
import { formatSnapshotSavedAt } from '../../hooks/useAppState'

interface HistoryListProps {
  snapshots: HistorySummary[]
  loading: boolean
  error: string
  onSelect: (videoId: string) => Promise<void>
}

function toLocalDateString(savedAt: string): string {
  const d = new Date(savedAt)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
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

const inputStyle: React.CSSProperties = {
  padding: '6px 10px',
  background: 'var(--c-bg-2)',
  border: '1px solid var(--c-line-strong)',
  color: 'var(--c-ink)',
  fontFamily: 'var(--f-mono)',
  fontSize: '13px',
}

export function HistoryList({ snapshots, loading, error, onSelect }: HistoryListProps) {
  const [fromDate, setFromDate] = useState<string>('')
  const [toDate, setToDate] = useState<string>('')

  const filteredSnapshots = useMemo(() => {
    if (!fromDate && !toDate) return snapshots
    return snapshots.filter((snap) => {
      const d = toLocalDateString(snap.savedAt)
      if (fromDate && d < fromDate) return false
      if (toDate && d > toDate) return false
      return true
    })
  }, [snapshots, fromDate, toDate])

  const isFilterActive = fromDate !== '' || toDate !== ''

  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '12px',
          padding: '32px',
          justifyContent: 'center',
          color: 'var(--c-ink-dim)',
          fontFamily: 'var(--f-mono)',
          fontSize: '12px',
        }}
      >
        <div
          data-testid="history-loading-spinner"
          style={{
            width: '22px',
            height: '22px',
            border: '2px solid var(--c-line-strong)',
            borderTopColor: 'var(--c-ink)',
            borderRadius: '50%',
            animation: 'spin 0.7s linear infinite',
          }}
        />
        <span>読み込み中…</span>
      </div>
    )
  }

  if (error) {
    return (
      <div
        role="alert"
        style={{
          padding: '16px',
          background: 'rgba(179,0,27,0.06)',
          border: '1px solid rgba(179,0,27,0.25)',
          borderLeft: '3px solid var(--c-error)',
          color: 'var(--c-error)',
          fontFamily: 'var(--f-mono)',
          fontSize: '12px',
        }}
      >
        {error}
      </div>
    )
  }

  if (snapshots.length === 0) {
    return (
      <p
        style={{
          padding: '32px',
          textAlign: 'center',
          fontFamily: 'var(--f-mono)',
          fontSize: '12px',
          color: 'var(--c-ink-mute)',
        }}
      >
        履歴がありません
      </p>
    )
  }

  return (
    <div className="space-y-3">
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.12em',
            color: 'var(--c-ink-dim)',
          }}
        >
          From
          <input
            type="date"
            value={fromDate}
            onChange={(e) => setFromDate(e.target.value)}
            style={inputStyle}
          />
        </label>
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.12em',
            color: 'var(--c-ink-dim)',
          }}
        >
          To
          <input
            type="date"
            value={toDate}
            onChange={(e) => setToDate(e.target.value)}
            style={inputStyle}
          />
        </label>
        {isFilterActive && (
          <button
            onClick={() => {
              setFromDate('')
              setToDate('')
            }}
            style={{
              fontFamily: 'var(--f-mono)',
              fontSize: '11px',
              letterSpacing: '0.12em',
              textTransform: 'uppercase',
              padding: '5px 10px',
              background: 'transparent',
              color: 'var(--c-ink-dim)',
              border: '1px solid var(--c-line-strong)',
              cursor: 'pointer',
            }}
          >
            クリア
          </button>
        )}
      </div>

      {filteredSnapshots.length === 0 ? (
        <p
          style={{
            padding: '32px',
            textAlign: 'center',
            fontFamily: 'var(--f-mono)',
            fontSize: '12px',
            color: 'var(--c-ink-mute)',
          }}
        >
          該当する履歴がありません
        </p>
      ) : (
        <section className="card-editorial">
          <div className="eyebrow">
            ARCHIVE
            <div className="eyebrow__rule" />
          </div>
          <table
            className="w-full table-fixed"
            style={{ fontSize: '14px', lineHeight: 1.7, marginTop: '8px' }}
          >
            <thead>
              <tr>
                <th style={{ ...thStyle, textAlign: 'left' }}>動画ID</th>
                <th style={{ ...thStyle, width: '160px' }}>保存日時</th>
                <th style={{ ...thStyle, width: '100px' }}>視聴者数</th>
                <th style={{ ...thStyle, width: '100px' }}>コメント数</th>
                <th style={{ ...thStyle, width: '80px', textAlign: 'center' }} />
              </tr>
            </thead>
            <tbody>
              {filteredSnapshots.map((snap, i) => {
                const rowBg = i % 2 === 0 ? 'var(--c-bg)' : 'var(--c-bg-2)'
                return (
                  <tr
                    key={snap.videoId}
                    style={{
                      borderBottom: '1px solid var(--c-line)',
                      background: rowBg,
                      transition: 'background 0.15s',
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLTableRowElement).style.background =
                        'rgba(0,95,120,0.06)'
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
                        color: 'var(--c-ink)',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {snap.videoId}
                    </td>
                    <td
                      style={{
                        padding: '10px 16px',
                        fontFamily: 'var(--f-mono)',
                        fontSize: '12px',
                        color: 'var(--c-ink-dim)',
                        textAlign: 'center',
                      }}
                    >
                      {formatSnapshotSavedAt(snap.savedAt)}
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
                      {snap.userCount}
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
                      {snap.commentCount}
                    </td>
                    <td style={{ padding: '10px 16px', textAlign: 'center' }}>
                      <button
                        aria-label="表示"
                        onClick={() => {
                          void onSelect(snap.videoId)
                        }}
                        style={{
                          fontFamily: 'var(--f-mono)',
                          fontSize: '11px',
                          letterSpacing: '0.1em',
                          textTransform: 'uppercase',
                          padding: '4px 10px',
                          background: 'transparent',
                          color: 'var(--c-accent-2)',
                          border: '1px solid var(--c-accent-2)',
                          cursor: 'pointer',
                          transition: 'background 0.2s',
                        }}
                        onMouseEnter={(e) => {
                          (e.currentTarget as HTMLButtonElement).style.background =
                            'rgba(0,95,120,0.08)'
                        }}
                        onMouseLeave={(e) => {
                          (e.currentTarget as HTMLButtonElement).style.background = 'transparent'
                        }}
                      >
                        表示
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </section>
      )}
    </div>
  )
}
