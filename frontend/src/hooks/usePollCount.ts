import { useCallback, useRef, useState } from 'react'
import { searchComments, type Comment, HttpError } from '../utils/api'
import { countVotes, type VoteCounts } from '../utils/countVotes'

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
  isLoading: boolean
  errorMsg: string
  lastUpdated: string
}

const initialState: PollState = {
  keywords: [],
  counts: {},
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
        errorMsg: '',
      }
    })
  }, [])

  const removeKeyword = useCallback((word: string) => {
    setState((prev) => {
      const keywords = prev.keywords.filter((k) => k !== word)
      const counts = { ...prev.counts }
      delete counts[word]
      return { ...prev, keywords, counts }
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
      const counts = countVotes(comments, keywords)
      const timeStr = new Date().toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
      setState((prev) => ({
        ...prev,
        counts,
        isLoading: false,
        lastUpdated: timeStr,
      }))
    } catch (e) {
      if (!(e instanceof Error)) return
      if (e.name === 'AbortError') return

      let errorMsg: string = ERROR_MESSAGES.GENERIC
      if (e.name === 'TimeoutError') {
        errorMsg = ERROR_MESSAGES.TIMEOUT
      } else if (e instanceof HttpError) {
        if (e.status === 404) errorMsg = ERROR_MESSAGES.SERVER_UNREACHABLE
        else if (e.status >= 500) errorMsg = ERROR_MESSAGES.SERVER_ERROR
      } else if (e instanceof TypeError && e.message.includes('Failed to fetch')) {
        errorMsg = ERROR_MESSAGES.NETWORK
      }
      setState((prev) => ({ ...prev, isLoading: false, errorMsg }))
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
