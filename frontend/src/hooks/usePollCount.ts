import { useCallback, useRef, useState } from 'react'
import { searchComments, type Comment } from '../utils/api'
import { mapHttpError } from '../utils/mapHttpError'
import { countVotes, type VoteCounts, type VoteVoters } from '../utils/countVotes'

export const POLL_INTERVAL_SEC = 15

const ERROR_MESSAGES = {
  GENERIC: '集計に失敗しました。再試行してください。',
  SERVER_UNREACHABLE: 'サーバーに接続できません。',
  SERVER_ERROR: 'サーバーエラーが発生しました。しばらく待ってから再試行してください。',
  NETWORK: 'ネットワークエラー。接続を確認してください。',
  NO_KEYWORDS: 'キーワードを追加してください。',
  TIMEOUT: '応答がタイムアウトしました。再試行してください。',
} as const

interface PollState {
  keywords: string[]
  counts: VoteCounts
  voters: VoteVoters
  isLoading: boolean
  errorMsg: string
  lastUpdated: string
}

const initialState: PollState = {
  keywords: [],
  counts: {},
  voters: {},
  isLoading: false,
  errorMsg: '',
  lastUpdated: '--:--:--',
}

export function usePollCount() {
  const [state, setState] = useState<PollState>(initialState)
  const controllerRef = useRef<AbortController | null>(null)

  const addKeyword = useCallback((word: string) => {
    const trimmed = word.trim()
    if (!trimmed) return
    setState((prev) => {
      if (prev.keywords.includes(trimmed)) return prev
      const keywords = [...prev.keywords, trimmed]
      return {
        ...prev,
        keywords,
        counts: { ...prev.counts, [trimmed]: prev.counts[trimmed] ?? 0 },
        voters: { ...prev.voters, [trimmed]: prev.voters[trimmed] ?? [] },
        errorMsg: '',
      }
    })
  }, [])

  const removeKeyword = useCallback((word: string) => {
    setState((prev) => {
      const keywords = prev.keywords.filter((k) => k !== word)
      const counts = { ...prev.counts }
      const voters = { ...prev.voters }
      delete counts[word]
      delete voters[word]
      return { ...prev, keywords, counts, voters }
    })
  }, [])

  const clearKeywords = useCallback(() => {
    if (controllerRef.current) controllerRef.current.abort()
    setState(initialState)
  }, [])

  const recount = useCallback(async () => {
    const keywords = state.keywords
    if (keywords.length === 0) {
      setState((prev) => ({ ...prev, errorMsg: ERROR_MESSAGES.NO_KEYWORDS }))
      return
    }

    if (controllerRef.current) controllerRef.current.abort()
    const controller = new AbortController()
    controllerRef.current = controller

    setState((prev) => ({ ...prev, isLoading: true, errorMsg: '' }))

    try {
      const comments: Comment[] = (await searchComments(keywords, controller.signal)) ?? []
      const { counts, voters } = countVotes(comments, keywords)
      const timeStr = new Date().toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
      setState((prev) => ({
        ...prev,
        counts,
        voters,
        isLoading: false,
        lastUpdated: timeStr,
      }))
    } catch (e) {
      try {
        const code = mapHttpError(e)
        const errorMsg = ERROR_MESSAGES[code]
        setState((prev) => ({ ...prev, isLoading: false, errorMsg }))
      } catch {
        return
      }
    }
  }, [state.keywords])

  return {
    ...state,
    totalVotes: Object.values(state.counts).reduce((a, b) => a + b, 0),
    addKeyword,
    removeKeyword,
    clearKeywords,
    recount,
  }
}
