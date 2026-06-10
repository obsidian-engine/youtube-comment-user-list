import { useState, useCallback, useRef } from 'react'
import { searchComments, type Comment } from '../utils/api'
import { mapHttpError } from '../utils/mapHttpError'

const KEYWORDS_KEY = 'comment-search-keywords'

const loadKeywords = (): string[] => {
  const data = localStorage.getItem(KEYWORDS_KEY)
  return data ? JSON.parse(data) : ['メモ'] // デフォルト値
}

const saveKeywords = (keywords: string[]): void => {
  localStorage.setItem(KEYWORDS_KEY, JSON.stringify(keywords))
}

interface SearchState {
  keywords: string[]
  comments: Comment[]
  isLoading: boolean
  errorMsg: string | null
  lastUpdated: string | null
  intervalSec: number
}

export function useCommentSearch() {
  const [state, setState] = useState<SearchState>({
    keywords: loadKeywords(),
    comments: [],
    isLoading: false,
    errorMsg: '',
    lastUpdated: '--:--:--',
    intervalSec: 0, // デフォルトOFF
  })

  const controllerRef = useRef<AbortController | null>(null)

  const addKeyword = useCallback((word: string) => {
    const trimmed = word.trim()
    if (!trimmed) return

    setState((prev) => {
      if (prev.keywords.includes(trimmed)) return prev
      const updated = [...prev.keywords, trimmed]
      saveKeywords(updated)
      return { ...prev, keywords: updated }
    })
  }, [])

  const removeKeyword = useCallback((word: string) => {
    setState((prev) => {
      const updated = prev.keywords.filter((k) => k !== word)
      saveKeywords(updated)
      return { ...prev, keywords: updated }
    })
  }, [])

  const setIntervalSec = useCallback((value: number) => {
    setState((prev) => ({ ...prev, intervalSec: value }))
  }, [])

  const clearComments = useCallback(() => {
    setState((prev) => ({
      ...prev,
      comments: [],
      errorMsg: null,
      lastUpdated: null,
    }))
  }, [])

  const search = useCallback(async () => {
    if (state.keywords.length === 0) {
      setState((prev) => ({ ...prev, errorMsg: 'キーワードを追加してください' }))
      return
    }

    // 前のリクエストをキャンセル
    if (controllerRef.current) {
      controllerRef.current.abort()
    }

    const controller = new AbortController()
    controllerRef.current = controller

    setState((prev) => ({ ...prev, isLoading: true, errorMsg: '' }))

    try {
      const comments = await searchComments(state.keywords, controller.signal)
      const now = new Date()
      const timeStr = now.toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })

      setState((prev) => ({
        ...prev,
        comments: comments ?? [],
        isLoading: false,
        lastUpdated: timeStr,
      }))
    } catch (e) {
      if (!(e instanceof Error)) return
      if (e.name === 'AbortError') return

      const ERROR_MESSAGES = {
        SERVER_UNREACHABLE: 'サーバーに接続できません。',
        SERVER_ERROR: 'サーバーエラーが発生しました。しばらく待ってから再試行してください。',
        NETWORK: 'ネットワークエラー。接続を確認してください。',
        TIMEOUT: '応答がタイムアウトしました。再試行してください。',
        GENERIC: '検索に失敗しました。再試行してください。',
      } as const

      const code = mapHttpError(e)
      const errorMsg = ERROR_MESSAGES[code]
      setState((prev) => ({
        ...prev,
        isLoading: false,
        errorMsg,
      }))
    }
  }, [state.keywords])

  return {
    ...state,
    addKeyword,
    removeKeyword,
    setIntervalSec,
    search,
    clearComments,
  }
}
