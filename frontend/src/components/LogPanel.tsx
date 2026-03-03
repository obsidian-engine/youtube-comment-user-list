import type { LogEntry, LogLevel } from '../hooks/useLogEntries'

interface LogPanelProps {
  entries: LogEntry[]
  onClear: () => void
}

const levelClass: Record<LogLevel, string> = {
  info: 'text-slate-700 dark:text-slate-300',
  warn: 'text-amber-700 dark:text-amber-400',
  error: 'text-rose-700 dark:text-rose-400',
}

function formatTime(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export function LogPanel({ entries, onClear }: LogPanelProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-[13px] text-slate-500 dark:text-slate-400">{entries.length} 件</span>
        <button
          onClick={onClear}
          disabled={entries.length === 0}
          className="px-3 py-1.5 text-[13px] rounded-md bg-slate-100 dark:bg-white/10 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-white/20 disabled:opacity-40 disabled:cursor-not-allowed transition"
        >
          クリア
        </button>
      </div>

      <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
        <table className="w-full table-fixed text-[14px] leading-7">
          <thead className="bg-gradient-to-br from-slate-400 to-slate-500 dark:from-slate-600 dark:to-slate-700 text-white dark:text-slate-100">
            <tr>
              <th className="text-center px-4 py-3.5 w-[100px] font-semibold text-[13px]">時刻</th>
              <th className="text-center px-4 py-3.5 font-semibold text-[13px]">メッセージ</th>
              <th className="text-center px-4 py-3.5 w-[80px] font-semibold text-[13px]">追加</th>
              <th className="text-center px-4 py-3.5 w-[80px] font-semibold text-[13px]">
                スキップ
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200/60 dark:divide-slate-600/40">
            {entries.map((entry) => (
              <tr
                key={entry.id}
                className="hover:bg-slate-50/50 dark:hover:bg-white/[0.02] transition-colors"
              >
                <td className="text-center px-4 py-2.5 text-[13px] text-slate-500 dark:text-slate-400 tabular-nums">
                  {formatTime(entry.timestamp)}
                </td>
                <td className={`px-4 py-2.5 text-[13px] ${levelClass[entry.level]}`}>
                  {entry.message}
                </td>
                <td className="text-center px-4 py-2.5 text-[13px] text-slate-600 dark:text-slate-300 tabular-nums">
                  {entry.addedCount != null ? entry.addedCount : '-'}
                </td>
                <td
                  className={`text-center px-4 py-2.5 text-[13px] tabular-nums ${
                    entry.skippedCount && entry.skippedCount > 0
                      ? 'text-amber-700 dark:text-amber-400 font-semibold'
                      : 'text-slate-600 dark:text-slate-300'
                  }`}
                >
                  {entry.skippedCount != null ? entry.skippedCount : '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {entries.length === 0 && (
          <p className="px-4 py-5 text-[13px] text-slate-500 dark:text-slate-400">
            ログはありません。
          </p>
        )}
      </section>
    </div>
  )
}
