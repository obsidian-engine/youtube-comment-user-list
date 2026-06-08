import { useCallback, useRef, useState } from 'react'
import { searchComments, type Comment, HttpError } from '../utils/api'
import { countVotes, type VoteCounts } from '../utils/countVotes'
import { parseKeywordsTxt } from '../utils/parseKeywordsTxt'

const ERROR_MESSAGES = {
  GENERIC: '集計に失敗しました。再試行してください。',
  SERVER_UNREACHABLE: 'サーバーに接続できません。',
  SERVER_ERROR: 'サーバーエラーが発生しました。しばらく待ってから再試行してください。',
  NETWORK: 'ネットワークエラー。接続を確認してください。',
  NO_KEYWORDS: 'キーワードを読み込んでください。',
  EMPTY_FILE: 'txt ファイルにキーワードが含まれていません。',
  READ_FAILED: 'txt ファイルの読み込みに失敗しました。',
  INVALID_LINES: '「,」を含む行は使用できません（除外しました）：',
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

  const loadKeywordsFromFile = useCallback(async (file: File) => {
    // 進行中の recount があれば cancel。旧 keywords でキャプチャされた結果が
    // 新 keywords の counts に上書きされる race を防ぐ。
    if (controllerRef.current) controllerRef.current.abort()
    try {
      const text = await file.text()
      const { keywords, invalid } = parseKeywordsTxt(text)
      if (keywords.length === 0) {
        setState((prev) => ({ ...prev, errorMsg: ERROR_MESSAGES.EMPTY_FILE }))
        return
      }
      const errorMsg =
        invalid.length > 0 ? `${ERROR_MESSAGES.INVALID_LINES}${invalid.join(' / ')}` : ''
      setState((prev) => ({
        ...prev,
        keywords,
        counts: Object.fromEntries(keywords.map((k) => [k, 0])),
        errorMsg,
      }))
    } catch {
      setState((prev) => ({ ...prev, errorMsg: ERROR_MESSAGES.READ_FAILED }))
    }
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
      // 移植先の searchComments は Promise<Comment[] | null> のため null を [] に吸収
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
      // ユーザー操作 / hook 内部での abort は state を変更しない
      if (e.name === 'AbortError') return

      let errorMsg: string = ERROR_MESSAGES.GENERIC
      // AbortSignal.timeout 経由のタイムアウト（10 秒）。ユーザー abort と区別してメッセージ提示
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
    loadKeywordsFromFile,
    clearKeywords,
    recount,
  }
}
