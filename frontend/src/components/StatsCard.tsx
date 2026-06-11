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

  const cardStyle: React.CSSProperties = {
    background: 'var(--c-bg-2)',
    border: '1px solid var(--c-line-strong)',
    padding: '20px 24px',
  }

  const labelStyle: React.CSSProperties = {
    fontFamily: 'var(--f-mono)',
    fontSize: '10px',
    letterSpacing: '0.2em',
    textTransform: 'uppercase',
    color: 'var(--c-ink-mute)',
    marginBottom: '4px',
  }

  const valueStyle: React.CSSProperties = {
    fontFamily: 'var(--f-mono)',
    fontSize: '18px',
    fontWeight: 700,
    color: 'var(--c-ink)',
    tabularNums: true,
  } as React.CSSProperties

  return (
    <div style={cardStyle}>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {/* 総ユーザー数 */}
        <div>
          <div style={labelStyle}>総ユーザー数</div>
          <div style={valueStyle}>
            {totalUsers}
            <span
              style={{
                fontSize: '12px',
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
          <div style={{ ...valueStyle, fontSize: '13px' }}>{monitoringStartTime}</div>
        </div>

        {/* 画面最終更新 */}
        <div>
          <div style={labelStyle}>最終更新</div>
          <div style={valueStyle}>{lastUpdated || '--:--:--'}</div>
        </div>

        {/* クラウド保存 */}
        <div>
          <div style={labelStyle}>クラウド保存</div>
          <div style={valueStyle}>{lastSnapshotAt || '--:--'}</div>
        </div>
      </div>

      {/* ステータスインジケーター */}
      <div
        style={{
          marginTop: '16px',
          paddingTop: '16px',
          borderTop: '1px solid var(--c-line)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              background: active ? 'var(--c-success)' : 'var(--c-ink-mute)',
              animation: active ? 'pulse 1.6s ease-in-out infinite' : 'none',
            }}
          />
          <span
            style={{
              fontFamily: 'var(--f-mono)',
              fontSize: '11px',
              letterSpacing: '0.16em',
              textTransform: 'uppercase',
              color: active ? 'var(--c-success)' : 'var(--c-ink-mute)',
            }}
          >
            {active ? '監視中' : '停止中'}
          </span>
        </div>
        {skippedCount > 0 && (
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
        )}
      </div>
    </div>
  )
}
