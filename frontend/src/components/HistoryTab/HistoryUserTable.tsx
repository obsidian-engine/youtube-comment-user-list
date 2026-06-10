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
        // copy failed silently
      }
    }
  }, [copyText])

  return (
    <button
      onClick={handleCopy}
      title="チャンネルURLをコピー"
      aria-label="チャンネルURLをコピー"
      className={`flex-shrink-0 transition-colors ${
        copied
          ? 'text-green-500'
          : 'text-slate-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-300'
      }`}
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
      <p className="py-5 text-center text-slate-500 dark:text-slate-400 text-[13px]">
        視聴者データがありません
      </p>
    )
  }

  return (
    <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
      <table className="w-full table-fixed text-[14px] leading-7">
        <thead className="bg-gradient-to-br from-slate-400 to-slate-500 dark:from-slate-600 dark:to-slate-700 text-white dark:text-slate-100">
          <tr>
            <th className="text-center px-4 py-3 w-[60px] font-semibold text-[13px] tracking-wide uppercase">
              #
            </th>
            <th className="text-center px-4 py-3 font-semibold text-[13px] tracking-wide uppercase">
              名前
            </th>
            <th className="text-center px-4 py-3 font-semibold text-[13px] tracking-wide uppercase w-[100px]">
              発言数
            </th>
            <th className="text-center px-4 py-3 font-semibold text-[13px] tracking-wide uppercase w-[160px] hidden md:table-cell">
              初回コメント
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-200/60 dark:divide-slate-600/40">
          {sorted.map((user, i) => {
            const channelUrl = user.channelId
              ? `https://www.youtube.com/channel/${user.channelId}`
              : null
            const name = user.displayName || user.channelId || 'Unknown'
            return (
              <tr
                key={`${user.channelId || user.displayName}-${i}`}
                className={`transition-colors duration-150 hover:bg-sky-100 dark:hover:bg-sky-900/30 ${
                  i % 2 === 0
                    ? 'bg-slate-100/50 dark:bg-slate-800/20'
                    : 'bg-slate-200/40 dark:bg-slate-700/25'
                }`}
              >
                <td className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 font-medium text-center">
                  {String(i + 1).padStart(2, '0')}
                </td>
                <td className="px-4 py-3 text-slate-800 dark:text-slate-200 font-medium">
                  <div className="flex items-center gap-1.5">
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
                <td className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 font-medium text-center">
                  {user.commentCount ?? 0}
                </td>
                <td className="px-4 py-3 text-slate-600 dark:text-slate-300 font-mono text-[13px] hidden md:table-cell text-center">
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
