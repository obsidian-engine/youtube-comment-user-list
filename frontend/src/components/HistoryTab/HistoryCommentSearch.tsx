import { useState, useMemo } from 'react'
import type { Comment } from '../../utils/api'

interface HistoryCommentSearchProps {
  comments: Comment[]
}

export function HistoryCommentSearch({ comments }: HistoryCommentSearchProps) {
  const [keyword, setKeyword] = useState('')

  const filtered = useMemo(() => {
    const trimmed = keyword.trim()
    if (!trimmed) return comments
    const lower = trimmed.toLowerCase()
    return comments.filter(
      (c) => c.message.toLowerCase().includes(lower) || c.displayName.toLowerCase().includes(lower),
    )
  }, [comments, keyword])

  const formatDate = (iso: string): string => {
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  return (
    <section className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur p-4 space-y-3">
      <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-200">コメント検索</h3>
      <input
        type="text"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        placeholder="キーワードで絞り込み"
        aria-label="コメント検索キーワード"
        className="w-full px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 focus:outline-none focus:ring-2 focus:ring-neutral-400/60 text-[14px]"
      />
      <div className="text-[12px] text-slate-500 dark:text-slate-400">
        {filtered.length} / {comments.length} 件
      </div>
      {filtered.length === 0 ? (
        <p className="py-4 text-center text-slate-500 dark:text-slate-400 text-[13px]">
          コメントがありません
        </p>
      ) : (
        <ul className="divide-y divide-slate-200/60 dark:divide-slate-600/40 max-h-96 overflow-y-auto">
          {filtered.map((c) => (
            <li key={c.id} className="py-2.5 space-y-0.5">
              <div className="flex items-center justify-between gap-2">
                <span className="text-[13px] font-medium text-slate-700 dark:text-slate-200 truncate">
                  {c.displayName}
                </span>
                <span className="text-[11px] text-slate-400 dark:text-slate-500 shrink-0">
                  {formatDate(c.publishedAt)}
                </span>
              </div>
              <p className="text-[13px] text-slate-600 dark:text-slate-300">{c.message}</p>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
