import { useEffect, useMemo, useState } from 'react'
import type { Comment } from '../utils/api'
import { countVotes, type MatchMode, type VoteCounts, type VoteVoters } from '../utils/countVotes'
import { loadStoredMatchMode, saveStoredMatchMode } from '../utils/pollMatchMode'

export interface UseVoteTallyCoreOptions {
  keywords: string[]
  comments: Comment[]
}

export interface UseVoteTallyCoreResult {
  matchMode: MatchMode
  setMatchMode: (mode: MatchMode) => void
  counts: VoteCounts
  voters: VoteVoters
  totalVotes: number
}

/**
 * 集計コア: matchMode 管理(load/save) + countVotes 呼び出し + counts/voters/totalVotes 算出。
 * snapshot(useVoteTally) と live(usePollCount) の両方が再利用する。
 * chips/API/AbortController は呼び出し側が保持する。
 */
export function useVoteTallyCore({
  keywords,
  comments,
}: UseVoteTallyCoreOptions): UseVoteTallyCoreResult {
  const [matchMode, setMatchModeState] = useState<MatchMode>(() => loadStoredMatchMode())

  useEffect(() => {
    saveStoredMatchMode(matchMode)
  }, [matchMode])

  const { counts, voters } = useMemo(
    () => countVotes(comments, keywords, matchMode),
    [comments, keywords, matchMode],
  )

  const totalVotes = useMemo(() => Object.values(counts).reduce((a, b) => a + b, 0), [counts])

  const setMatchMode = (mode: MatchMode) => {
    setMatchModeState(mode)
  }

  return { matchMode, setMatchMode, counts, voters, totalVotes }
}
