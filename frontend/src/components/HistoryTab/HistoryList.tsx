import { useState, useMemo } from 'react'
import type { HistorySummary } from '../../utils/api'
import { formatSnapshotSavedAt } from '../../hooks/useAppState'

interface HistoryListProps {
  snapshots: HistorySummary[]
  loading: boolean
  error: string
  onSelect: (videoId: string) => Promise<void>
}

// savedAt (ISO 8601) をローカルタイムゾーンの YYYY-MM-DD 文字列に変換する
function toLocalDateString(savedAt: string): string {
  const d = new Date(savedAt)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
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
      <div className="flex items-center gap-3 py-8 justify-center text-slate-600 dark:text-slate-300">
        <div
          data-testid="history-loading-spinner"
          className="animate-spin rounded-full h-6 w-6 border-2 border-slate-300 border-t-slate-600 dark:border-slate-600 dark:border-t-slate-300"
        />
        <span>読み込み中...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div
        role="alert"
        className="rounded-lg p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 text-red-700 dark:text-red-300"
      >
        {error}
      </div>
    )
  }

  if (snapshots.length === 0) {
    return (
      <p className="py-8 text-center text-slate-500 dark:text-slate-400 text-[13px]">
        履歴がありません
      </p>
    )
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3 flex-wrap">
        <label className="flex items-center gap-2 text-[13px] text-slate-600 dark:text-slate-300">
          From
          <input
            type="date"
            value={fromDate}
            onChange={(e) => setFromDate(e.target.value)}
            className="rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 px-2 py-1 text-[13px]"
          />
        </label>
        <label className="flex items-center gap-2 text-[13px] text-slate-600 dark:text-slate-300">
          To
          <input
            type="date"
            value={toDate}
            onChange={(e) => setToDate(e.target.value)}
            className="rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 px-2 py-1 text-[13px]"
          />
        </label>
        {isFilterActive && (
          <button
            onClick={() => {
              setFromDate('')
              setToDate('')
            }}
            className="px-3 py-1 text-[12px] rounded-md bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600 transition-colors"
          >
            クリア
          </button>
        )}
      </div>

      {filteredSnapshots.length === 0 ? (
        <p className="py-8 text-center text-slate-500 dark:text-slate-400 text-[13px]">
          該当する履歴がありません
        </p>
      ) : (
        <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
          <table className="w-full table-fixed text-[14px] leading-7">
            <thead className="bg-gradient-to-br from-slate-400 to-slate-500 dark:from-slate-600 dark:to-slate-700 text-white dark:text-slate-100">
              <tr>
                <th className="text-left px-4 py-3 font-semibold text-[13px] tracking-wide uppercase">
                  動画ID
                </th>
                <th className="text-center px-4 py-3 font-semibold text-[13px] tracking-wide uppercase w-[160px]">
                  保存日時
                </th>
                <th className="text-center px-4 py-3 font-semibold text-[13px] tracking-wide uppercase w-[100px]">
                  視聴者数
                </th>
                <th className="text-center px-4 py-3 font-semibold text-[13px] tracking-wide uppercase w-[100px]">
                  コメント数
                </th>
                <th className="px-4 py-3 w-[80px]" />
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200/60 dark:divide-slate-600/40">
              {filteredSnapshots.map((snap, i) => (
                <tr
                  key={snap.videoId}
                  className={`transition-colors duration-150 hover:bg-sky-100 dark:hover:bg-sky-900/30 ${
                    i % 2 === 0
                      ? 'bg-slate-100/50 dark:bg-slate-800/20'
                      : 'bg-slate-200/40 dark:bg-slate-700/25'
                  }`}
                >
                  <td className="px-4 py-3 text-slate-800 dark:text-slate-200 font-mono text-[13px] truncate">
                    {snap.videoId}
                  </td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300 text-center text-[13px]">
                    {formatSnapshotSavedAt(snap.savedAt)}
                  </td>
                  <td className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 text-center">
                    {snap.userCount}
                  </td>
                  <td className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 text-center">
                    {snap.commentCount}
                  </td>
                  <td className="px-4 py-3 text-center">
                    <button
                      aria-label="表示"
                      onClick={() => {
                        void onSelect(snap.videoId)
                      }}
                      className="px-3 py-1 text-[12px] rounded-md bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600 transition-colors"
                    >
                      表示
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}
    </div>
  )
}
