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


export function UserTable({ users }: UserTableProps) {
  return (
    <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
      <table className="w-full table-auto text-[14px] leading-7">
        <thead className="bg-gradient-to-br from-slate-400 to-slate-500 dark:from-slate-600 dark:to-slate-700 text-white dark:text-slate-100">
          <tr>
            <th className="text-left px-4 py-3.5 w-[72px] font-semibold text-[13px] tracking-wide uppercase">#</th>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">名前</th>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">発言数</th>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">初回コメント</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-200/60 dark:divide-slate-600/40">
          {users.map((user, i) => (
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
      {users.length === 0 && (
        <p className="px-4 py-5 text-[13px] text-slate-500 dark:text-slate-400">ユーザーがいません。</p>
      )}
    </section>
  )
}