import type { HistorySnapshot } from '../../utils/api'
import { formatSnapshotSavedAt } from '../../hooks/useAppState'
import { HistoryUserTable } from './HistoryUserTable'
import { HistoryCommentSearch } from './HistoryCommentSearch'

interface HistoryDetailProps {
  snapshot: HistorySnapshot
  onBack: () => void
}

export function HistoryDetail({ snapshot, onBack }: HistoryDetailProps) {
  return (
    <div className="space-y-4">
      {/* ヘッダー */}
      <div className="flex items-center gap-3 flex-wrap">
        <button
          aria-label="戻る"
          onClick={onBack}
          className="px-3 py-1.5 text-[13px] rounded-md bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600 transition-colors"
        >
          ← 戻る
        </button>
        <span className="inline-flex items-center px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 border border-amber-300 dark:border-amber-700">
          閲覧モード (read-only)
        </span>
        <span className="text-[13px] text-slate-600 dark:text-slate-300 font-mono">
          {snapshot.videoId}
        </span>
        {snapshot.savedAt && (
          <span className="text-[12px] text-slate-500 dark:text-slate-400">
            保存日時: {formatSnapshotSavedAt(snapshot.savedAt)}
          </span>
        )}
      </div>

      {/* 視聴者一覧 */}
      <section>
        <h3 className="text-sm font-semibold mb-2 text-slate-700 dark:text-slate-200">
          視聴者一覧 ({snapshot.users.length} 人)
        </h3>
        <HistoryUserTable users={snapshot.users} />
      </section>

      {/* コメント検索 */}
      <HistoryCommentSearch comments={snapshot.comments} />
    </div>
  )
}
