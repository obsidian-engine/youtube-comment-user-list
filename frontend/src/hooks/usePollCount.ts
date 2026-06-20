import { useCallback, useEffect, useRef, useState } from 'react'
import { searchComments, type Comment } from '../utils/api'
import { mapHttpError } from '../utils/mapHttpError'
import { useVoteTallyCore } from './useVoteTallyCore'
import type { MatchMode } from '../utils/countVotes'

export const POLL_INTERVAL_SEC = 15
const STORAGE_KEY = 'pollKeywords'

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

interface PollMeta {
  keywords: string[]
  isLoading: boolean
  errorMsg: string
  lastUpdated: string
  /** clearResults 呼出後 true になり、次の recount で false に戻る */
  isResultCleared: boolean
}

const initialMeta: PollMeta = {
  keywords: [],
  isLoading: false,
  errorMsg: '',
  lastUpdated: '--:--:--',
  isResultCleared: false,
}

export function usePollCount() {
  const [meta, setMeta] = useState<PollMeta>(() => {
    const stored = loadStoredKeywords()
    return { ...initialMeta, keywords: stored }
  })

  // core に渡す comments — performRecount 完了時に更新
  const [fetchedComments, setFetchedComments] = useState<Comment[]>([])

  const controllerRef = useRef<AbortController | null>(null)
  const metaRef = useRef(meta)
  metaRef.current = meta

  // 集計コア: matchMode 管理(load/save) + countVotes + totalVotes
  const core = useVoteTallyCore({ keywords: meta.keywords, comments: fetchedComments })
  const coreRef = useRef(core)
  coreRef.current = core

  useEffect(() => {
    saveStoredKeywords(meta.keywords)
  }, [meta.keywords])

  const addKeyword = useCallback((word: string) => {
    const trimmed = word.trim()
    if (!trimmed) return
    setMeta((prev) => {
      if (prev.keywords.includes(trimmed)) return prev
      return {
        ...prev,
        keywords: [...prev.keywords, trimmed],
        errorMsg: '',
      }
    })
  }, [])

  const removeKeyword = useCallback((word: string) => {
    setMeta((prev) => ({
      ...prev,
      keywords: prev.keywords.filter((k) => k !== word),
    }))
  }, [])

  const clearKeywords = useCallback(() => {
    if (controllerRef.current) controllerRef.current.abort()
    setMeta({ ...initialMeta, keywords: [] })
    setFetchedComments([])
  }, [])

  const clearResults = useCallback(() => {
    setFetchedComments([])
    setMeta((prev) => ({
      ...prev,
      errorMsg: '',
      lastUpdated: '--:--:--',
      isResultCleared: true,
    }))
  }, [])

  const performRecount = useCallback(async (keywords: string[]) => {
    if (keywords.length === 0) {
      setMeta((prev) => ({ ...prev, errorMsg: ERROR_MESSAGES.NO_KEYWORDS }))
      return
    }

    if (controllerRef.current) controllerRef.current.abort()
    const controller = new AbortController()
    controllerRef.current = controller

    setMeta((prev) => ({ ...prev, isLoading: true, errorMsg: '' }))

    try {
      const comments: Comment[] = (await searchComments(keywords, controller.signal)) ?? []
      const timeStr = new Date().toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
      setFetchedComments(comments)
      setMeta((prev) => ({
        ...prev,
        isLoading: false,
        lastUpdated: timeStr,
        isResultCleared: false,
      }))
    } catch (e) {
      try {
        const code = mapHttpError(e)
        const errorMsg = ERROR_MESSAGES[code]
        setMeta((prev) => ({ ...prev, isLoading: false, errorMsg }))
      } catch {
        return
      }
    }
  }, [])

  const recount = useCallback(async () => {
    await performRecount(metaRef.current.keywords)
  }, [performRecount])

  const setMatchMode = useCallback(
    (mode: MatchMode) => {
      const keywords = metaRef.current.keywords
      const currentMode = coreRef.current.matchMode
      if (currentMode === mode) return

      coreRef.current.setMatchMode(mode)

      if (keywords.length > 0) {
        void performRecount(keywords)
      }
    },
    [performRecount],
  )

  // clearResults 後は counts/voters を空にして旧 API の振る舞いを維持
  const counts = meta.isResultCleared ? {} : core.counts
  const voters = meta.isResultCleared ? {} : core.voters
  const totalVotes = meta.isResultCleared ? 0 : core.totalVotes

  return {
    keywords: meta.keywords,
    matchMode: core.matchMode,
    counts,
    voters,
    isLoading: meta.isLoading,
    errorMsg: meta.errorMsg,
    lastUpdated: meta.lastUpdated,
    totalVotes,
    addKeyword,
    removeKeyword,
    clearKeywords,
    clearResults,
    recount,
    setMatchMode,
  }
}
