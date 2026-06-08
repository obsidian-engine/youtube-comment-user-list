import { Fragment, useState } from 'react'
import type { VoteCounts, VoteVoters } from '../../utils/countVotes'

interface PollResultsProps {
  keywords: string[]
  counts: VoteCounts
  voters: VoteVoters
  totalVotes: number
  isLoading: boolean
}

function voterListToTsv(voters: Array<{ displayName: string; channelId: string }>): string {
  return voters.map((v) => `${v.displayName}\t${v.channelId}`).join('\n')
}

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

export function PollResults({ keywords, counts, voters, totalVotes, isLoading }: PollResultsProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [copiedKeyword, setCopiedKeyword] = useState<string | null>(null)

  if (keywords.length === 0) {
    return null
  }

  const toggleExpand = (word: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(word)) next.delete(word)
      else next.add(word)
      return next
    })
  }

  const handleCopy = async (word: string) => {
    const list = voters[word] ?? []
    if (list.length === 0) return
    const ok = await copyToClipboard(voterListToTsv(list))
    if (ok) {
      setCopiedKeyword(word)
      setTimeout(() => setCopiedKeyword((cur) => (cur === word ? null : cur)), 1500)
    }
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
          {keywords.map((word) => {
            const list = voters[word] ?? []
            const isOpen = expanded.has(word)
            return (
              <Fragment key={word}>
                <tr className="cursor-pointer hover:bg-slate-50" onClick={() => toggleExpand(word)}>
                  <td className="px-4 py-3 text-slate-800">
                    <span className="inline-block w-4 text-slate-400">{isOpen ? '▼' : '▶'}</span>
                    {word}
                  </td>
                  <td className="px-4 py-3 text-right tabular-nums font-semibold">
                    {counts[word] ?? 0}
                  </td>
                </tr>
                {isOpen && (
                  <tr className="bg-slate-50/60">
                    <td colSpan={2} className="px-4 py-3">
                      {list.length === 0 ? (
                        <div className="text-[12px] text-slate-500">投票したユーザーはいません</div>
                      ) : (
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="text-[12px] text-slate-600">
                              投票ユーザー ({list.length}人)
                            </span>
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                void handleCopy(word)
                              }}
                              className="text-[12px] px-2 py-1 rounded-md bg-white border border-slate-300/80 hover:bg-slate-100"
                            >
                              {copiedKeyword === word ? 'コピー済' : '名前+channelId をコピー'}
                            </button>
                          </div>
                          <ul className="space-y-1 text-[13px]">
                            {list.map((v) => (
                              <li
                                key={v.channelId}
                                className="flex items-center gap-2 text-slate-700"
                              >
                                <span>{v.displayName}</span>
                                <span className="text-[11px] text-slate-400 font-mono">
                                  {v.channelId}
                                </span>
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </td>
                  </tr>
                )}
              </Fragment>
            )
          })}
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
