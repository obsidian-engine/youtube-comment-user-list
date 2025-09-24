import { useState, useMemo } from 'react'
import { useCurrentTime } from '../hooks/useCurrentTime'

interface User {
  channelId?: string
  displayName?: string
  joinedAt?: string
  commentCount?: number
  firstCommentedAt?: string
}

// Union type to handle both object and string users
type UserData = User | string

interface UserTableProps {
  users: UserData[]
  intervalSec?: number
  setIntervalSec?: (value: number) => void
  lastUpdated?: string
  lastFetchTime?: string
  isRefreshing?: boolean
}

type SortField = 'commentCount' | 'firstCommentedAt'
type SortOrder = 'asc' | 'desc'

interface SortState {
  field: SortField | null
  order: SortOrder
}

// Helper functions to handle mixed data types
const getUserDisplayName = (user: UserData): string => {
  if (typeof user === 'string') return user
  return user.displayName || user.channelId || 'Unknown User'
}

const getUserKey = (user: UserData, index: number): string => {
  if (typeof user === 'string') return `${user}-${index}`
  return `${user.channelId || user.displayName}-${index}`
}

const getUserCommentCount = (user: UserData): number => {
  if (typeof user === 'string') return 0
  return user.commentCount ?? 0
}

const getUserFirstComment = (user: UserData): string => {
  if (typeof user === 'string') return '--:--'
  if (user.firstCommentedAt && user.firstCommentedAt !== '') {
    return new Date(user.firstCommentedAt).toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
      timeZone: 'Asia/Tokyo'
    })
  }
  return '--:--'
}

const getUserFirstCommentTime = (user: UserData): Date | null => {
  if (typeof user === 'string') return null
  if (user.firstCommentedAt && user.firstCommentedAt !== '') {
    return new Date(user.firstCommentedAt)
  }
  return null
}

interface SortButtonProps {
  field: SortField
  currentSort: SortState
  onSort: (field: SortField) => void
  children: React.ReactNode
}

function SortButton({ field, currentSort, onSort, children }: SortButtonProps) {
  const isActive = currentSort.field === field
  const isDesc = isActive && currentSort.order === 'desc'
  const isAsc = isActive && currentSort.order === 'asc'

  const ariaLabel = field === 'commentCount' ? '発言数でソート' : '初回コメントでソート'

  return (
    <button
      className="flex items-center gap-1 hover:text-white/80 transition-colors"
      onClick={() => onSort(field)}
      aria-label={ariaLabel}
    >
      {children}
      <span className="flex flex-col w-3 h-3">
        <svg
          className={`w-3 h-1.5 transition-colors ${
            isAsc ? 'text-white' : 'text-white/40'
          }`}
          fill="currentColor"
          viewBox="0 0 12 6"
        >
          <path d="M6 0L12 6H0z" />
        </svg>
        <svg
          className={`w-3 h-1.5 transition-colors ${
            isDesc ? 'text-white' : 'text-white/40'
          }`}
          fill="currentColor"
          viewBox="0 0 12 6"
        >
          <path d="M6 6L0 0h12z" />
        </svg>
      </span>
    </button>
  )
}


export function UserTable({ users, intervalSec = 0, setIntervalSec, lastUpdated = '--:--:--', lastFetchTime = '', isRefreshing = false }: UserTableProps) {
  const [sortState, setSortState] = useState<SortState>({ field: null, order: 'asc' })
  const currentTime = useCurrentTime()

  const handleSort = (field: SortField) => {
    setSortState(prevState => {
      if (prevState.field === field) {
        return { field, order: prevState.order === 'asc' ? 'desc' : 'asc' }
      } else {
        // 発言数は降順から開始、初回コメントは昇順から開始
        return { field, order: field === 'commentCount' ? 'desc' : 'asc' }
      }
    })
  }

  const handleReset = () => {
    setSortState({ field: null, order: 'asc' })
  }

  const isSorted = sortState.field !== null

  const sortedUsers = useMemo(() => {
    if (!sortState.field) {
      return users
    }

    const sorted = [...users].sort((a, b) => {
      if (sortState.field === 'commentCount') {
        const countA = getUserCommentCount(a)
        const countB = getUserCommentCount(b)
        return sortState.order === 'asc' ? countA - countB : countB - countA
      } else if (sortState.field === 'firstCommentedAt') {
        const timeA = getUserFirstCommentTime(a)
        const timeB = getUserFirstCommentTime(b)

        // null値の処理（コメントしていない場合は最後に配置）
        if (!timeA && !timeB) return 0
        if (!timeA) return 1
        if (!timeB) return -1

        const diff = timeA.getTime() - timeB.getTime()
        return sortState.order === 'asc' ? diff : -diff
      }

      return 0
    })

    return sorted
  }, [users, sortState])

  return (
    <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
      {/* コントロールヘッダー */}
      <div className="px-4 py-3 border-b border-slate-200/60 dark:border-slate-600/40 bg-slate-50/50 dark:bg-slate-800/30">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="bg-white/60 dark:bg-white/5 rounded-lg px-3 py-2 border border-slate-200/50 dark:border-white/10">
              <div className="text-[10px] text-slate-500 dark:text-slate-400 mb-0.5">現在時刻</div>
              <div className="text-xs font-medium text-slate-700 dark:text-slate-200 tabular-nums">
                {currentTime}
              </div>
            </div>
            <div className="bg-white/60 dark:bg-white/5 rounded-lg px-3 py-2 border border-slate-200/50 dark:border-white/10">
              <div className="text-[10px] text-slate-500 dark:text-slate-400 mb-0.5">画面最終更新</div>
              <div className="text-xs font-medium text-slate-700 dark:text-slate-200 tabular-nums">
                {lastUpdated}
              </div>
            </div>
            <button
              onClick={handleReset}
              disabled={!isSorted || isRefreshing}
              aria-label="ソートリセット"
              className={`text-[12px] px-3 py-1.5 rounded-md transition-colors ${
                isSorted
                  ? 'bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600'
                  : 'bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-500 cursor-not-allowed'
              }`}
            >
              ↻ デフォルト順
            </button>
            {setIntervalSec && (
              <div className="flex items-center gap-2">
                <label htmlFor="interval-select" className="text-[11px] text-slate-500 dark:text-slate-400">更新間隔</label>
                <select
                  id="interval-select"
                  aria-label="更新間隔"
                  value={intervalSec}
                  onChange={(e) => setIntervalSec(Number(e.target.value))}
                  disabled={isRefreshing}
                  className="text-[12px] px-2 py-1 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10"
                >
                  <option value="0">停止</option>
                  <option value="10">10s</option>
                  <option value="15">15s</option>
                  <option value="30">30s</option>
                  <option value="60">60s</option>
                </select>
              </div>
            )}
          </div>
          <div className="flex items-center gap-3">

            {isRefreshing && (
              <div className="flex items-center gap-3">
                <div 
                  data-testid="loading-spinner"
                  className="animate-spin rounded-full h-8 w-8 border-2 border-slate-300 border-t-slate-600 dark:border-slate-600 dark:border-t-slate-300"
                />
                <span className="text-sm text-slate-600 dark:text-slate-400">
                  データ更新中...
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
      <table className="w-full table-auto text-[14px] leading-7">
        <thead className="bg-gradient-to-br from-slate-400 to-slate-500 dark:from-slate-600 dark:to-slate-700 text-white dark:text-slate-100">
          <tr>
            <th className="text-left px-4 py-3.5 w-[72px] font-semibold text-[13px] tracking-wide uppercase">#</th>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">名前</th>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">
              <SortButton field="commentCount" currentSort={sortState} onSort={handleSort}>
                発言数
              </SortButton>
            </th>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">
              <SortButton field="firstCommentedAt" currentSort={sortState} onSort={handleSort}>
                初回コメント
              </SortButton>
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-200/60 dark:divide-slate-600/40">
          {sortedUsers.map((user, i) => (
            <tr
              key={getUserKey(user, i)}
              className={`transition-colors duration-150 hover:bg-slate-200/40 dark:hover:bg-slate-700/20 ${
                i % 2 === 0
                  ? 'bg-slate-100/50 dark:bg-slate-800/20'
                  : 'bg-slate-200/40 dark:bg-slate-700/25'
              }`}
            >
              <td className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 font-medium">
                {String(i + 1).padStart(2, '0')}
              </td>
              <td
                className="px-4 py-3 truncate-1 text-slate-800 dark:text-slate-200 font-medium"
                title={getUserDisplayName(user)}
              >
                {getUserDisplayName(user)}
              </td>
              <td
                className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 font-medium"
                data-testid={`comment-count-${i}`}
              >
                {getUserCommentCount(user)}
              </td>
              <td
                className="px-4 py-3 text-slate-600 dark:text-slate-300 font-mono text-[13px]"
                data-testid={`first-comment-${i}`}
              >
                {getUserFirstComment(user)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {sortedUsers.length === 0 && (
        <p className="px-4 py-5 text-[13px] text-slate-500 dark:text-slate-400">ユーザーがいません。</p>
      )}
    </section>
  )
}