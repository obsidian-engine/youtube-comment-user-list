import type { VoteCounts } from '../../utils/countVotes'

interface PollResultsProps {
  keywords: string[]
  counts: VoteCounts
  totalVotes: number
  isLoading: boolean
}

export function PollResults({ keywords, counts, totalVotes, isLoading }: PollResultsProps) {
  if (keywords.length === 0) {
    return null
  }

  return (
    <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 bg-white/80 backdrop-blur">
      <table className="w-full text-[14px]">
        <thead className="bg-gradient-to-br from-slate-400 to-slate-500 text-white">
          <tr>
            <th className="text-left px-4 py-3.5 font-semibold text-[13px]">キーワード</th>
            <th className="text-right px-4 py-3.5 w-[120px] font-semibold text-[13px]">票数</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-200/60">
          {keywords.map((word) => (
            <tr key={word}>
              <td className="px-4 py-3 text-slate-800">{word}</td>
              <td className="px-4 py-3 text-right tabular-nums font-semibold">
                {counts[word] ?? 0}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot className="bg-slate-50/60">
          <tr>
            <td className="px-4 py-3 text-[13px] text-slate-600">合計</td>
            <td className="px-4 py-3 text-right tabular-nums font-semibold text-slate-700">
              {totalVotes}
            </td>
          </tr>
        </tfoot>
      </table>

      {isLoading && (
        <div className="flex items-center justify-center py-4">
          <div className="animate-spin rounded-full h-6 w-6 border-2 border-slate-300 border-t-slate-600" />
        </div>
      )}
    </section>
  )
}
