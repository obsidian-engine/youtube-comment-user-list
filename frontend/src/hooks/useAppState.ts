import { useState, useCallback, useRef, useEffect, useMemo } from 'react'
import {
  getStatus,
  getUsers,
  postPull,
  postReset,
  postSwitchVideo,
  BackendError,
} from '../utils/api'
import { logger } from '../utils/logger'
import { decideUsers } from '../utils/decideUsers'
import type { User } from '../utils/api'
import type { LogLevel } from './useLogEntries'

const TOAST_KEY = 'snapshotRestoreToastShown'

/**
 * snapshotSavedAt を HH:MM (今日) または MM/DD HH:MM (それ以外) にフォーマットする。
 * 未指定またはフォーマット不能の場合は空文字列を返す。
 */
export function formatSnapshotSavedAt(snapshotSavedAt?: string): string {
  if (!snapshotSavedAt) return ''
  const d = new Date(snapshotSavedAt)
  if (isNaN(d.getTime())) return ''

  const pad = (n: number) => String(n).padStart(2, '0')
  const today = new Date()
  const isToday =
    d.getFullYear() === today.getFullYear() &&
    d.getMonth() === today.getMonth() &&
    d.getDate() === today.getDate()
  return isToday
    ? `${pad(d.getHours())}:${pad(d.getMinutes())}`
    : `${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/**
 * snapshotSavedAt を元に toast メッセージを生成する。
 * 未指定・既表示・フォーマット不能の場合は空文字列を返す。
 * sessionStorage への書込は行わない（呼び出し側の useEffect で副作用化）。
 */
function consumeSnapshotRestoreMsg(snapshotSavedAt?: string): string {
  if (!snapshotSavedAt) return ''
  const alreadyShown =
    typeof window !== 'undefined' && window.sessionStorage?.getItem(TOAST_KEY) === '1'
  if (alreadyShown) return ''

  const formatted = formatSnapshotSavedAt(snapshotSavedAt)
  if (!formatted) return ''
  return `${formatted} の保存状態を取得しました`
}

interface LoadingStates {
  switching: boolean
  pulling: boolean
  resetting: boolean
  refreshing: boolean
}

interface AppState {
  status: string
  users: User[]
  videoId: string
  currentVideoId?: string
  intervalSec: number
  lastUpdated: string
  lastFetchTime: string
  errorMsg: string
  infoMsg: string
  snapshotRestoreMsg: string
  lastSnapshotAt: string
  startTime?: string
  scheduledStartTime?: string
  skippedCount: number
  loadingStates: LoadingStates
}

interface AppActions {
  setVideoId: (value: string) => void
  setIntervalSec: (value: number) => void
  refresh: () => Promise<void>
  onSwitch: () => Promise<void>
  onPull: () => Promise<void>
  onPullSilent: () => Promise<void>
  onReset: () => Promise<void>
  clearInfoMsg: () => void
  clearSnapshotRestoreMsg: () => void
}

type AddEntryFn = (
  level: LogLevel,
  message: string,
  data?: { addedCount?: number; skippedCount?: number },
) => void

export function useAppState(addEntry?: AddEntryFn) {
  const [state, setState] = useState<AppState>({
    status: 'WAITING',
    users: [],
    videoId: localStorage.getItem('videoId') || '',
    currentVideoId: undefined,
    intervalSec: 60,
    lastUpdated: '--:--:--',
    lastFetchTime: '',
    errorMsg: '',
    infoMsg: '',
    snapshotRestoreMsg: '',
    lastSnapshotAt: '',
    startTime: undefined,
    scheduledStartTime: undefined,
    skippedCount: 0,
    loadingStates: {
      switching: false,
      pulling: false,
      resetting: false,
      refreshing: false,
    },
  })

  // AbortController管理用のref
  const refreshControllerRef = useRef<AbortController | null>(null)
  const switchControllerRef = useRef<AbortController | null>(null)
  const pullControllerRef = useRef<AbortController | null>(null)
  const resetControllerRef = useRef<AbortController | null>(null)

  const updateClock = useCallback(() => {
    const d = new Date()
    const pad = (n: number) => String(n).padStart(2, '0')
    const timeString = `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    setState((prev) => ({ ...prev, lastUpdated: timeString }))
  }, [])

  const refresh = useCallback(
    async (options: { clearUsers?: boolean } = {}) => {
      logger.log('🎯 refresh({ clearUsers:', options.clearUsers, '}) called from useAppState')

      try {
        // 前のリクエストをキャンセル
        if (refreshControllerRef.current) {
          logger.log('🛑 Aborting previous refresh request')
          refreshControllerRef.current.abort()
        }

        logger.log('🔄 Refresh starting...', new Date().toLocaleTimeString())
        setState((prev) => ({
          ...prev,
          loadingStates: { ...prev.loadingStates, refreshing: true },
        }))

        // 新しいAbortControllerを作成
        const controller = new AbortController()
        refreshControllerRef.current = controller

        const [st, us] = await Promise.all([
          getStatus(controller.signal),
          getUsers(controller.signal),
        ])

        // リクエストが成功したらcontrollerをクリア
        refreshControllerRef.current = null

        const status = st.status || 'WAITING'
        const fetched = Array.isArray(us) ? us : []

        // snapshotSavedAt: 初回ロード時に 1 度だけ toast を表示（sessionStorage への書込は useEffect で副作用化）
        const snapshotRestoreMsg = consumeSnapshotRestoreMsg(st.snapshotSavedAt)
        const newSnapshotAt = st.snapshotSavedAt
          ? formatSnapshotSavedAt(st.snapshotSavedAt)
          : undefined

        setState((prev) => {
          const finalUsers = decideUsers(prev.users, fetched, {
            clearUsers: options.clearUsers ?? false,
          })
          logger.log('📋 User list decision:', {
            clearUsers: options.clearUsers,
            serverUsers: fetched.length,
            existingUsers: prev.users.length,
            finalCount: finalUsers.length,
          })

          return {
            ...prev,
            status,
            users: finalUsers,
            startTime: st.startedAt,
            scheduledStartTime: st.scheduledStartTime,
            currentVideoId: st.videoId,
            errorMsg: '',
            lastSnapshotAt: newSnapshotAt ?? prev.lastSnapshotAt,
            ...(snapshotRestoreMsg ? { snapshotRestoreMsg } : {}),
          }
        })

        logger.log('✅ Refresh completed:', {
          status,
          userCount: fetched.length,
          clearUsers: options.clearUsers,
        })
      } catch (e) {
        // AbortErrorの場合はエラーメッセージを表示しない
        if (e instanceof Error && e.name === 'AbortError') {
          logger.log('🚫 Refresh aborted')
          return
        }

        logger.error('❌ Refresh failed:', e)
        setState((prev) => ({
          ...prev,
          errorMsg: '更新に失敗しました。しばらくしてから再試行してください。',
        }))
      } finally {
        updateClock()
        setState((prev) => ({
          ...prev,
          loadingStates: { ...prev.loadingStates, refreshing: false },
        }))
      }
    },
    [updateClock],
  )

  const handleAsyncAction = useCallback(
    async (
      action: (signal: AbortSignal) => Promise<void>,
      loadingKey: keyof LoadingStates,
      successMsg: string,
      errorMsgPrefix: string = '',
      controllerRef: React.MutableRefObject<AbortController | null>,
      shouldClearUsers: boolean = false, // 切替・リセット時のフラグ
    ) => {
      try {
        // 前のリクエストをキャンセル
        if (controllerRef.current) {
          controllerRef.current.abort()
        }

        setState((prev) => ({
          ...prev,
          loadingStates: { ...prev.loadingStates, [loadingKey]: true },
        }))

        // 新しいAbortControllerを作成
        const controller = new AbortController()
        controllerRef.current = controller

        await action(controller.signal)

        // リクエストが成功したらcontrollerをクリア
        controllerRef.current = null

        setState((prev) => ({ ...prev, errorMsg: '', infoMsg: successMsg }))

        // 取得系アクション（pulling）の場合は取得時刻を更新
        if (loadingKey === 'pulling') {
          const now = new Date()
          const pad = (n: number) => String(n).padStart(2, '0')
          const timeString = `${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
          setState((prev) => ({ ...prev, lastFetchTime: timeString }))
        }

        // 切替・リセット時はユーザーリストをクリア、それ以外は保持
        await refresh({ clearUsers: shouldClearUsers })
      } catch (e) {
        // AbortErrorの場合はエラーメッセージを表示しない
        if (e instanceof Error && e.name === 'AbortError') {
          logger.log(`🚫 ${loadingKey} action aborted`)
          return
        }

        const errorMessage = `${errorMsgPrefix}に失敗しました。${loadingKey === 'switching' ? '配信開始後に再度お試しください。' : ''}`
        setState((prev) => ({ ...prev, errorMsg: errorMessage }))
        addEntry?.('error', `${errorMsgPrefix}に失敗しました`)
        if (e instanceof BackendError && e.logs.length > 0) {
          e.logs.forEach((entry) => {
            const level: LogLevel =
              entry.level === 'warn' || entry.level === 'error' ? entry.level : 'info'
            addEntry?.(level, `[${entry.source}] ${entry.message}`)
          })
        }
      } finally {
        setState((prev) => ({
          ...prev,
          loadingStates: { ...prev.loadingStates, [loadingKey]: false },
        }))
      }
    },
    [refresh, addEntry],
  )

  const pullAction = useCallback(
    async (signal: AbortSignal) => {
      const res = await postPull(signal)
      setState((prev) => ({ ...prev, skippedCount: prev.skippedCount + res.skippedCount }))
      if (res.autoReset) {
        addEntry?.('warn', '配信終了を検知しました。再度切替してください。')
      }
      const level = res.skippedCount > 0 ? 'warn' : 'info'
      addEntry?.(level, res.skippedCount > 0 ? 'Pull完了（スキップあり）' : 'Pull完了', {
        addedCount: res.addedCount,
        skippedCount: res.skippedCount,
      })
      res.logs?.forEach((entry) => {
        const level: LogLevel =
          entry.level === 'warn' || entry.level === 'error' ? entry.level : 'info'
        addEntry?.(level, `[${entry.source}] ${entry.message}`)
      })
    },
    [addEntry],
  )

  const actions: AppActions = {
    setVideoId: useCallback((value: string) => {
      setState((prev) => ({ ...prev, videoId: value }))
    }, []),

    setIntervalSec: useCallback((value: number) => {
      setState((prev) => ({ ...prev, intervalSec: value }))
    }, []),

    refresh: useCallback(() => refresh(), [refresh]),

    onSwitch: useCallback(async () => {
      if (!state.videoId) {
        setState((prev) => ({ ...prev, errorMsg: 'videoId を入力してください。' }))
        return
      }
      setState((prev) => ({ ...prev, skippedCount: 0 }))
      await handleAsyncAction(
        async (signal) => {
          await postSwitchVideo(state.videoId, signal)
          localStorage.setItem('videoId', state.videoId)
          addEntry?.('info', `配信切替: ${state.videoId}`)
          await pullAction(signal)
        },
        'switching',
        '切替しました',
        '切替',
        switchControllerRef,
        true, // 切替時はユーザーリストをクリア
      )
    }, [state.videoId, handleAsyncAction, addEntry, pullAction]),

    onPull: useCallback(async () => {
      await handleAsyncAction(pullAction, 'pulling', '取得しました', '取得', pullControllerRef)
    }, [handleAsyncAction, pullAction]),

    onPullSilent: useCallback(async () => {
      await handleAsyncAction(
        pullAction,
        'pulling',
        '', // 自動更新時はメッセージなし
        '取得',
        pullControllerRef,
      )
    }, [handleAsyncAction, pullAction]),

    onReset: useCallback(async () => {
      setState((prev) => ({ ...prev, skippedCount: 0 }))
      await handleAsyncAction(
        async (signal) => {
          await postReset(signal)
          addEntry?.('info', 'リセット')
        },
        'resetting',
        'リセットしました',
        'リセット',
        resetControllerRef,
        true, // リセット時もユーザーリストをクリア
      )
    }, [handleAsyncAction, addEntry]),

    clearInfoMsg: useCallback(() => {
      setState((prev) => ({ ...prev, infoMsg: '' }))
    }, []),

    clearSnapshotRestoreMsg: useCallback(() => {
      setState((prev) => ({ ...prev, snapshotRestoreMsg: '' }))
    }, []),
  }

  // snapshotRestoreMsg が実際に描画された後に sessionStorage へ副作用を書き込む
  useEffect(() => {
    if (state.snapshotRestoreMsg) {
      sessionStorage.setItem(TOAST_KEY, '1')
    }
  }, [state.snapshotRestoreMsg])

  const derivedState = useMemo(
    () => ({
      ...state,
      active: state.status === 'ACTIVE',
      reserved: state.status === 'RESERVED',
    }),
    [state],
  )

  return { state: derivedState, actions }
}
