import type { LogEntry, LogLevel } from '../hooks/useLogEntries'

interface LogPanelProps {
  entries: LogEntry[]
  onClear: () => void
}

const levelColor: Record<LogLevel, string> = {
  info: 'var(--c-ink)',
  warn: '#b56b00',
  error: 'var(--c-error)',
}

function formatTime(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
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

export function LogPanel({ entries, onClear }: LogPanelProps) {
  return (
    <div className="space-y-3">
      <section className="card-editorial">
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 4px 0 0',
          }}
        >
          <div className="eyebrow" style={{ flex: 1 }}>
            CONSOLE
            <div className="eyebrow__rule" />
          </div>
          <div style={{ padding: '0 16px' }}>
            <span
              style={{
                fontFamily: 'var(--f-mono)',
                fontSize: '11px',
                letterSpacing: '0.14em',
                color: 'var(--c-ink-mute)',
                marginRight: '12px',
              }}
            >
              {entries.length} 件
            </span>
            <button
              onClick={onClear}
              disabled={entries.length === 0}
              style={{
                fontFamily: 'var(--f-mono)',
                fontSize: '11px',
                letterSpacing: '0.14em',
                textTransform: 'uppercase',
                padding: '6px 12px',
                background: 'transparent',
                color: entries.length === 0 ? 'var(--c-ink-mute)' : 'var(--c-ink)',
                border: '1px solid var(--c-line-strong)',
                cursor: entries.length === 0 ? 'not-allowed' : 'pointer',
                opacity: entries.length === 0 ? 0.5 : 1,
              }}
            >
              クリア
            </button>
          </div>
        </div>

        <table className="w-full table-fixed" style={{ fontSize: '13px', lineHeight: '1.7' }}>
          <thead>
            <tr>
              <th style={{ ...thStyle, width: '100px' }}>時刻</th>
              <th style={{ ...thStyle, textAlign: 'left' }}>メッセージ</th>
              <th style={{ ...thStyle, width: '80px' }}>追加</th>
              <th style={{ ...thStyle, width: '80px' }}>スキップ</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry, i) => (
              <tr
                key={entry.id}
                style={{
                  borderBottom: '1px solid var(--c-line)',
                  background: i % 2 === 0 ? 'var(--c-bg)' : 'var(--c-bg-2)',
                }}
              >
                <td
                  style={{
                    padding: '8px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: 'var(--c-ink-mute)',
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {formatTime(entry.timestamp)}
                </td>
                <td
                  style={{
                    padding: '8px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: levelColor[entry.level],
                  }}
                >
                  {entry.message}
                </td>
                <td
                  style={{
                    padding: '8px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: 'var(--c-ink-dim)',
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {entry.addedCount != null ? entry.addedCount : '-'}
                </td>
                <td
                  style={{
                    padding: '8px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color:
                      entry.skippedCount && entry.skippedCount > 0 ? '#b56b00' : 'var(--c-ink-dim)',
                    fontWeight: entry.skippedCount && entry.skippedCount > 0 ? 700 : 400,
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {entry.skippedCount != null ? entry.skippedCount : '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {entries.length === 0 && (
          <p
            style={{
              padding: '20px 16px',
              fontFamily: 'var(--f-mono)',
              fontSize: '12px',
              color: 'var(--c-ink-mute)',
            }}
          >
            ログはありません。
          </p>
        )}
      </section>
    </div>
  )
}
