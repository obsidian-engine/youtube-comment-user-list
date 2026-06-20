import { useState, useMemo } from 'react'
import type { Comment } from '../utils/api'
import type { MatchMode, VoteCounts, VoteVoters } from '../utils/countVotes'
import { parseKeywords } from '../utils/parseKeywords'
import { useVoteTallyCore } from './useVoteTallyCore'

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
  const parsedKeywords = useMemo(() => parseKeywords(keywordsInput), [keywordsInput])

  const { matchMode, setMatchMode, counts, voters, totalVotes } = useVoteTallyCore({
    keywords: parsedKeywords,
    comments: options.comments,
  })

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
