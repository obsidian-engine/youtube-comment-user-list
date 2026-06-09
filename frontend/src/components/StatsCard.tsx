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
  } catch (error) {
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

  return (
    <div className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-slate-800/50 dark:to-slate-700/50 backdrop-blur">
      <div className="px-4 py-4 md:px-6 md:py-5">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-6">
          {/* 総ユーザー数 */}
          <div className="flex items-center gap-2 md:gap-3 min-w-0">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 md:w-10 md:h-10 rounded-lg bg-blue-500/10 dark:bg-blue-400/10 flex items-center justify-center">
                <span className="text-base md:text-lg">👥</span>
              </div>
            </div>
            <div className="min-w-0">
              <div className="text-xs md:text-sm font-medium text-slate-600 dark:text-slate-400 truncate">
                総ユーザー数
              </div>
              <div className="text-xl md:text-2xl font-bold text-slate-900 dark:text-white tabular-nums">
                {totalUsers}
                <span className="text-xs md:text-sm font-normal text-slate-500 dark:text-slate-400 ml-1">
                  人
                </span>
              </div>
            </div>
          </div>

          {/* 監視開始時間 */}
          <div className="flex items-center gap-2 md:gap-3 min-w-0">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 md:w-10 md:h-10 rounded-lg bg-purple-500/10 dark:bg-purple-400/10 flex items-center justify-center">
                <span className="text-base md:text-lg">⏰</span>
              </div>
            </div>
            <div className="min-w-0">
              <div className="text-xs md:text-sm font-medium text-slate-600 dark:text-slate-400 truncate">
                監視開始時間
              </div>
              <div className="font-bold text-slate-900 dark:text-white">
                <span className="text-sm md:text-lg break-all">{monitoringStartTime}</span>
              </div>
            </div>
          </div>

          {/* 画面最終更新 */}
          <div className="flex items-center gap-2 md:gap-3 min-w-0">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 md:w-10 md:h-10 rounded-lg bg-green-500/10 dark:bg-green-400/10 flex items-center justify-center">
                <span className="text-base md:text-lg">🔄</span>
              </div>
            </div>
            <div className="min-w-0">
              <div className="text-xs md:text-sm font-medium text-slate-600 dark:text-slate-400 truncate">
                画面最終更新
              </div>
              <div className="font-bold text-slate-900 dark:text-white">
                <span className="text-sm md:text-lg tabular-nums">{lastUpdated || '--:--:--'}</span>
              </div>
            </div>
          </div>

          {/* クラウド保存 */}
          <div className="flex items-center gap-2 md:gap-3 min-w-0">
            <div className="flex-shrink-0">
              <div className="w-8 h-8 md:w-10 md:h-10 rounded-lg bg-sky-500/10 dark:bg-sky-400/10 flex items-center justify-center">
                <span className="text-base md:text-lg">☁️</span>
              </div>
            </div>
            <div className="min-w-0">
              <div className="text-xs md:text-sm font-medium text-slate-600 dark:text-slate-400 truncate">
                クラウド保存
              </div>
              <div className="font-bold text-slate-900 dark:text-white">
                <span className="text-sm md:text-lg tabular-nums">{lastSnapshotAt || '--:--'}</span>
              </div>
            </div>
          </div>
        </div>

        {/* ステータスインジケーター */}
        <div className="mt-4 pt-4 border-t border-slate-200/60 dark:border-slate-600/40">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div
                className={`w-3 h-3 rounded-full ${
                  active
                    ? 'bg-green-400 animate-pulse shadow-lg shadow-green-400/50'
                    : 'bg-slate-300 dark:bg-slate-600'
                }`}
              />
              <span className="text-sm font-medium text-slate-600 dark:text-slate-400">
                {active ? '監視中' : '停止中'}
              </span>
            </div>
            {skippedCount > 0 && (
              <div className="flex items-center gap-1.5">
                <span className="text-sm font-medium text-amber-600 dark:text-amber-400">
                  スキップ: {skippedCount}件
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
