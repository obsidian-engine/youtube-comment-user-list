import type { User } from '../utils/api'

interface StatsCardProps {
  users: User[]
  active: boolean
  startTime?: string
  lastUpdated?: string
  lastSnapshotAt?: string
  skippedCount: number
}

const getMonitoringStartTime = (startTime?: string): string => {
  if (!startTime) return '未開始'
  try {
    const start = new Date(startTime)
    if (isNaN(start.getTime())) return '未開始'
    return start.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return '未開始'
  }
}

export function StatsCard({
  users,
  active,
  startTime,
  lastUpdated,
  lastSnapshotAt,
  skippedCount,
}: StatsCardProps) {
  const totalUsers = users.length
  const monitoringStartTime = getMonitoringStartTime(active ? startTime : undefined)

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

  const statusClass = active ? 'eyebrow__status eyebrow__status--live' : 'eyebrow__status'

  return (
    <div className="card-editorial">
      {/* Eyebrow with status */}
      <div className="eyebrow">
        STATS
        <div className="eyebrow__rule" />
        <div className={statusClass}>
          <div className="eyebrow__status-dot" />
          {active ? '監視中' : '停止中'}
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

          {/* 監視開始時間 */}
          <div>
            <div style={labelStyle}>監視開始</div>
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
              {monitoringStartTime}
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
