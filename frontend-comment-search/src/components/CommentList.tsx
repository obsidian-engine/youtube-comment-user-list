import { CommentRow } from './CommentRow'
import type { Comment } from '../types'

interface CommentListProps {
  comments: Comment[]
  isChecked: (id: string) => boolean
  onToggle: (id: string) => void
  isLoading: boolean
}

export function CommentList({ comments, isChecked, onToggle, isLoading }: CommentListProps) {
  return (
    <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
      <table className="w-full table-fixed text-[14px] leading-7">
        <thead className="bg-gradient-to-br from-slate-400 to-slate-500 dark:from-slate-600 dark:to-slate-700 text-white dark:text-slate-100">
          <tr>
            <th className="text-center px-4 py-3.5 w-[80px] font-semibold text-[13px]">済</th>
            <th className="text-center px-4 py-3.5 w-[100px] font-semibold text-[13px]">時刻</th>
            <th className="text-center px-4 py-3.5 w-[150px] font-semibold text-[13px]">投稿者</th>
            <th className="text-center px-4 py-3.5 font-semibold text-[13px]">コメント</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-200/60 dark:divide-slate-600/40">
          {comments.map((comment, i) => (
            <CommentRow
              key={comment.id}
              comment={comment}
              index={i}
              isChecked={isChecked(comment.id)}
              onToggle={() => onToggle(comment.id)}
            />
          ))}
        </tbody>
      </table>

      {comments.length === 0 && !isLoading && (
        <p className="px-4 py-5 text-[13px] text-slate-500 dark:text-slate-400">
          該当するコメントがありません。
        </p>
      )}

      {isLoading && (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-slate-300 border-t-slate-600" />
        </div>
      )}
    </section>
  )
}
