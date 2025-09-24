import type { User } from '../utils/api'

interface StatsCardProps {
  users: User[]
  active: boolean
  startTime?: string
}



const getLatestCommentTime = (users: User[]): string => {
  let latestTime: Date | null = null
  
  for (const user of users) {
    if (user.firstCommentedAt && user.firstCommentedAt !== '') {
      try {
        const commentTime = new Date(user.firstCommentedAt)
        // 不正な日付の場合はInvalid Dateになるのでチェック
        if (!isNaN(commentTime.getTime())) {
          if (!latestTime || commentTime > latestTime) {
            latestTime = commentTime
          }
        }
      } catch (error) {
        // 不正な日付フォーマットの場合はスキップ
        console.warn('Invalid date format:', user.firstCommentedAt)
      }
    }
  }
  
  if (!latestTime) return 'なし'
  
  const now = new Date()
  const diffMs = now.getTime() - latestTime.getTime()
  const diffMinutes = Math.floor(diffMs / (1000 * 60))
  
  if (diffMinutes < 1) return '1分未満前'
  return `${diffMinutes}分前`
}

const getMonitoringDuration = (startTime?: string): string => {
  if (!startTime) return '停止中'
  
  try {
    const start = new Date(startTime)
    if (isNaN(start.getTime())) return '停止中'
    
    const now = new Date()
    const diffMs = now.getTime() - start.getTime()
    const diffMinutes = Math.floor(diffMs / (1000 * 60))
    
    if (diffMinutes < 1) return '1分未満'
    if (diffMinutes < 60) return `${diffMinutes}分`
    
    const hours = Math.floor(diffMinutes / 60)
    const minutes = diffMinutes % 60
    return `${hours}時間${minutes}分`
  } catch (error) {
    console.warn('Invalid start time format:', startTime)
    return '停止中'
  }
}

export function StatsCard({ users, active, startTime }: StatsCardProps) {
  const totalUsers = users.length
  const latestComment = getLatestCommentTime(users)
  const monitoringDuration = getMonitoringDuration(active ? startTime : undefined)

  return (
    <div className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-slate-800/50 dark:to-slate-700/50 backdrop-blur">
      <div className="px-6 py-5">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* 総ユーザー数 */}
          <div className="flex items-center gap-3">
            <div className="flex-shrink-0">
              <div className="w-10 h-10 rounded-lg bg-blue-500/10 dark:bg-blue-400/10 flex items-center justify-center">
                <span className="text-lg">👥</span>
              </div>
            </div>
            <div>
              <div className="text-sm font-medium text-slate-600 dark:text-slate-400">総ユーザー数</div>
              <div className="text-2xl font-bold text-slate-900 dark:text-white tabular-nums">
                {totalUsers}
                <span className="text-sm font-normal text-slate-500 dark:text-slate-400 ml-1">人</span>
              </div>
            </div>
          </div>



          {/* 監視時間 */}
          <div className="flex items-center gap-3">
            <div className="flex-shrink-0">
              <div className="w-10 h-10 rounded-lg bg-purple-500/10 dark:bg-purple-400/10 flex items-center justify-center">
                <span className="text-lg">⏰</span>
              </div>
            </div>
            <div>
              <div className="text-sm font-medium text-slate-600 dark:text-slate-400">監視時間</div>
              <div className="text-2xl font-bold text-slate-900 dark:text-white">
                <span className="text-lg">{monitoringDuration}</span>
              </div>
            </div>
          </div>

          {/* 最新コメント */}
          <div className="flex items-center gap-3">
            <div className="flex-shrink-0">
              <div className="w-10 h-10 rounded-lg bg-orange-500/10 dark:bg-orange-400/10 flex items-center justify-center">
                <span className="text-lg">💭</span>
              </div>
            </div>
            <div>
              <div className="text-sm font-medium text-slate-600 dark:text-slate-400">最新コメント</div>
              <div className="text-2xl font-bold text-slate-900 dark:text-white">
                <span className="text-lg">{latestComment}</span>
              </div>
            </div>
          </div>
        </div>

        {/* ステータスインジケーター */}
        <div className="mt-4 pt-4 border-t border-slate-200/60 dark:border-slate-600/40">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className={`w-3 h-3 rounded-full ${
                active 
                  ? 'bg-green-400 animate-pulse shadow-lg shadow-green-400/50' 
                  : 'bg-slate-300 dark:bg-slate-600'
              }`} />
              <span className="text-sm font-medium text-slate-600 dark:text-slate-400">
                {active ? '監視中' : '停止中'}
              </span>
            </div>

          </div>
        </div>
      </div>
    </div>
  )
}