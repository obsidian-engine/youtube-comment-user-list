import { useEffect, useMemo, useState } from 'react'
import type { Comment } from '../../utils/api'
import { countVotes, type MatchMode } from '../../utils/countVotes'
import { loadStoredMatchMode, saveStoredMatchMode } from '../../utils/pollMatchMode'
import { MatchModeDescription } from '../MatchModeDescription'
import { MatchModeToggle } from '../MatchModeToggle'
import { PollResults } from '../PollTab/PollResults'

interface HistoryVotesProps {
  comments: Comment[]
}

function parseKeywords(input: string): string[] {
  return [
    ...new Set(
      input
        .split(/[\n,]/)
        .map((k) => k.trim())
        .filter((k) => k.length > 0),
    ),
  ]
}

export function HistoryVotes({ comments }: HistoryVotesProps) {
  const [keywordsInput, setKeywordsInput] = useState('')
  const [matchMode, setMatchMode] = useState<MatchMode>(() => loadStoredMatchMode())

  const parsedKeywords = useMemo(() => parseKeywords(keywordsInput), [keywordsInput])

  useEffect(() => {
    saveStoredMatchMode(matchMode)
  }, [matchMode])

  const { counts, voters } = useMemo(
    () => countVotes(comments, parsedKeywords, matchMode),
    [comments, parsedKeywords, matchMode],
  )

  const totalVotes = useMemo(() => Object.values(counts).reduce((a, b) => a + b, 0), [counts])

  return (
    <div className="space-y-3">
      <MatchModeToggle matchMode={matchMode} onMatchModeChange={setMatchMode} />
      <MatchModeDescription matchMode={matchMode} variant="history" />
      <div className="flex gap-3 items-start">
        <textarea
          value={keywordsInput}
          onChange={(e) => setKeywordsInput(e.target.value)}
          placeholder="キーワード（改行またはカンマ区切り）"
          rows={3}
          aria-label="投票キーワード入力"
          className="flex-1 text-[13px] px-3 py-2 rounded-md border border-slate-300 bg-white text-slate-800 resize-none focus:ring-2 focus:ring-slate-400"
        />
        {parsedKeywords.length === 0 && (
          <p className="text-[12px] text-slate-500 pt-2">キーワードを入力すると集計します</p>
        )}
      </div>
      <PollResults
        keywords={parsedKeywords}
        counts={counts}
        voters={voters}
        totalVotes={totalVotes}
        isLoading={false}
      />
    </div>
  )
}
