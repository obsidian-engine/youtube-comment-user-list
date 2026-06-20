import { useEffect, useMemo, useState } from 'react'
import type { Comment } from '../utils/api'
import { countVotes, type MatchMode, type VoteCounts, type VoteVoters } from '../utils/countVotes'
import { loadStoredMatchMode, saveStoredMatchMode } from '../utils/pollMatchMode'
import { parseKeywords } from '../utils/parseKeywords'

type SnapshotOptions = {
  mode: 'snapshot'
  comments: Comment[]
}

// discriminated union: live モードは次 PR で追加
type UseVoteTallyOptions = SnapshotOptions

type UseVoteTallyResult = {
  keywordsInput: string
  setKeywordsInput: (input: string) => void
  matchMode: MatchMode
  setMatchMode: (mode: MatchMode) => void
  parsedKeywords: string[]
  counts: VoteCounts
  voters: VoteVoters
  totalVotes: number
}

export function useVoteTally(options: UseVoteTallyOptions): UseVoteTallyResult {
  const [keywordsInput, setKeywordsInput] = useState('')
  const [matchMode, setMatchMode] = useState<MatchMode>(() => loadStoredMatchMode())

  const parsedKeywords = useMemo(() => parseKeywords(keywordsInput), [keywordsInput])

  useEffect(() => {
    saveStoredMatchMode(matchMode)
  }, [matchMode])

  const { counts, voters } = useMemo(
    () => countVotes(options.comments, parsedKeywords, matchMode),
    [options.comments, parsedKeywords, matchMode],
  )

  const totalVotes = useMemo(() => Object.values(counts).reduce((a, b) => a + b, 0), [counts])

  return {
    keywordsInput,
    setKeywordsInput,
    matchMode,
    setMatchMode,
    parsedKeywords,
    counts,
    voters,
    totalVotes,
  }
}
