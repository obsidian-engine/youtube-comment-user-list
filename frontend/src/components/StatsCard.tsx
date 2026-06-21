import type { User } from '../utils/api'

interface StatsCardProps {
  users: User[]
  active: boolean
  reserved?: boolean
  startTime?: string
  scheduledStartTime?: string
  lastUpdated?: string
  lastSnapshotAt?: string
  skippedCount: number
}

const formatDateTime = (value?: string): string | null => {
  if (!value) return null
  try {
    const d = new Date(value)
    if (isNaN(d.getTime())) return null
    return d.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return null
  }
}

export function StatsCard({
  users,
  active,
  reserved = false,
  startTime,
  scheduledStartTime,
  lastUpdated,
  lastSnapshotAt,
  skippedCount,
}: StatsCardProps) {
  const totalUsers = users.length
  const startTimeLabel = reserved ? '開始予定' : '監視開始'
  const startTimeValue = reserved
    ? (formatDateTime(scheduledStartTime) ?? '未定')
    : active
      ? (formatDateTime(startTime) ?? '未開始')
      : '未開始'

  const labelStyle: React.CSSProperties = {
    fontFamily: 'var(--f-mono)',
    fontSize: '10px',
    letterSpacing: '0.2em',
    textTransform: 'uppercase',
    color: 'var(--c-ink-mute)',
    marginBottom: '6px',
  }

  const valueStyle: React.CSSProperties = {
    fontFamily: 'var(--f-display)',
    fontSize: '36px',
    fontWeight: 400,
    color: 'var(--c-ink)',
    fontVariantNumeric: 'tabular-nums',
    lineHeight: 1,
  } as React.CSSProperties

  const valueSmStyle: React.CSSProperties = {
    fontFamily: 'var(--f-display)',
    fontSize: '28px',
    fontWeight: 400,
    color: 'var(--c-ink)',
    fontVariantNumeric: 'tabular-nums',
    lineHeight: 1,
  } as React.CSSProperties

  const statusClass = active
    ? 'eyebrow__status eyebrow__status--live'
    : reserved
      ? 'eyebrow__status eyebrow__status--reserved'
      : 'eyebrow__status'
  const statusLabel = active ? '監視中' : reserved ? '予約中' : '停止中'

  return (
    <div className="card-editorial">
      {/* Eyebrow with status */}
      <div className="eyebrow">
        STATS
        <div className="eyebrow__rule" />
        <div className={statusClass}>
          <div className="eyebrow__status-dot" />
          {statusLabel}
        </div>
      </div>

      <div style={{ padding: '16px 20px 20px' }}>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          {/* 総ユーザー数 */}
          <div>
            <div style={labelStyle}>総ユーザー数</div>
            <div style={valueStyle}>
              {totalUsers}
              <span
                style={{
                  fontSize: '13px',
                  fontFamily: 'var(--f-mono)',
                  fontWeight: 400,
                  color: 'var(--c-ink-dim)',
                  marginLeft: '4px',
                }}
              >
                人
              </span>
            </div>
          </div>

          {/* 監視開始時間 or 開始予定 */}
          <div>
            <div style={labelStyle}>{startTimeLabel}</div>
            <div
              style={{
                fontFamily: 'var(--f-mono)',
                fontSize: '13px',
                fontWeight: 600,
                color: 'var(--c-ink)',
                lineHeight: 1.4,
                marginTop: '4px',
              }}
            >
              {startTimeValue}
            </div>
          </div>

          {/* 画面最終更新 */}
          <div>
            <div style={labelStyle}>最終更新</div>
            <div style={valueSmStyle}>{lastUpdated || '--:--:--'}</div>
          </div>

          {/* クラウド保存 */}
          <div>
            <div style={labelStyle}>クラウド保存</div>
            <div style={valueSmStyle}>{lastSnapshotAt || '--:--'}</div>
          </div>
        </div>

        {skippedCount > 0 && (
          <div
            style={{
              marginTop: '12px',
              paddingTop: '12px',
              borderTop: '1px solid var(--c-line)',
            }}
          >
            <span
              style={{
                fontFamily: 'var(--f-mono)',
                fontSize: '11px',
                letterSpacing: '0.14em',
                color: 'var(--c-ink-dim)',
              }}
            >
              スキップ: {skippedCount}件
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
