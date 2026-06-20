import { useCallback, useEffect, useRef, useState } from 'react'
import { searchComments, type Comment } from '../utils/api'
import { mapHttpError } from '../utils/mapHttpError'
import { countVotes, type MatchMode, type VoteCounts, type VoteVoters } from '../utils/countVotes'

export const POLL_INTERVAL_SEC = 15
const STORAGE_KEY = 'pollKeywords'
const MATCH_MODE_STORAGE_KEY = 'pollMatchMode'

function loadStoredMatchMode(): MatchMode {
  try {
    const raw = localStorage.getItem(MATCH_MODE_STORAGE_KEY)
    if (raw === 'exact' || raw === 'partial') return raw
  } catch {
    // ignore
  }
  return 'exact'
}

function loadStoredKeywords(): string[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) return []
    return parsed.filter((v): v is string => typeof v === 'string' && v.trim() !== '')
  } catch {
    return []
  }
}

function saveStoredKeywords(keywords: string[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(keywords))
  } catch {
    // localStorage 不可環境 (Safari private mode 等) は黙って無視
  }
}

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
  matchMode: MatchMode
  counts: VoteCounts
  voters: VoteVoters
  isLoading: boolean
  errorMsg: string
  lastUpdated: string
}

const initialState: PollState = {
  keywords: [],
  matchMode: 'exact',
  counts: {},
  voters: {},
  isLoading: false,
  errorMsg: '',
  lastUpdated: '--:--:--',
}

export function usePollCount() {
  const [state, setState] = useState<PollState>(() => {
    const stored = loadStoredKeywords()
    const matchMode = loadStoredMatchMode()
    if (stored.length === 0) return { ...initialState, matchMode }
    return {
      ...initialState,
      matchMode,
      keywords: stored,
      counts: Object.fromEntries(stored.map((k) => [k, 0])),
      voters: Object.fromEntries(stored.map((k) => [k, []])),
    }
  })
  const controllerRef = useRef<AbortController | null>(null)
  const stateRef = useRef(state)
  stateRef.current = state

  useEffect(() => {
    saveStoredKeywords(state.keywords)
  }, [state.keywords])

  useEffect(() => {
    try {
      localStorage.setItem(MATCH_MODE_STORAGE_KEY, state.matchMode)
    } catch {
      // ignore
    }
  }, [state.matchMode])

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

  const clearResults = useCallback(() => {
    setState((prev) => ({
      ...prev,
      counts: {},
      voters: {},
      errorMsg: '',
      lastUpdated: '--:--:--',
    }))
  }, [])

  const performRecount = useCallback(async (keywords: string[], matchMode: MatchMode) => {
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
      const { counts, voters } = countVotes(comments, keywords, matchMode)
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
  }, [])

  const recount = useCallback(async () => {
    await performRecount(state.keywords, state.matchMode)
  }, [state.keywords, state.matchMode, performRecount])

  const setMatchMode = useCallback(
    (mode: MatchMode) => {
      const { keywords, matchMode: currentMode } = stateRef.current
      if (currentMode === mode) return

      setState((prev) => ({ ...prev, matchMode: mode }))

      if (keywords.length > 0) {
        void performRecount(keywords, mode)
      }
    },
    [performRecount],
  )

  return {
    ...state,
    totalVotes: Object.values(state.counts).reduce((a, b) => a + b, 0),
    addKeyword,
    removeKeyword,
    clearKeywords,
    clearResults,
    recount,
    setMatchMode,
  }
}
