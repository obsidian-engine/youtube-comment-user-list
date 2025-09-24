import { useState, useCallback, useRef } from 'react'
import { getStatus, getUsers, postPull, postReset, postSwitchVideo } from '../utils/api'
import { sortUsersStable } from '../utils/sortUsers'
import { logger } from '../utils/logger'
import type { User } from '../utils/api'

interface LoadingStates {
  switching: boolean
  pulling: boolean
  resetting: boolean
  refreshing: boolean
}

interface AppState {
  active: boolean
  users: User[]
  videoId: string
  intervalSec: number
  lastUpdated: string
  lastFetchTime: string
  errorMsg: string
  infoMsg: string
  startTime?: string
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
}

export function useAppState() {
  const [state, setState] = useState<AppState>({
    active: false,
    users: [],
    videoId: localStorage.getItem('videoId') || '',
    intervalSec: 15,
    lastUpdated: '--:--:--',
    lastFetchTime: '',
    errorMsg: '',
    infoMsg: '',
    startTime: undefined,
    loadingStates: {
      switching: false,
      pulling: false,
      resetting: false,
      refreshing: false
    }
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
    setState(prev => ({ ...prev, lastUpdated: timeString }))
  }, [])

  // 切替・リセット時専用のrefresh（ユーザーリストを強制クリア）
  const refreshWithClear = useCallback(async () => {
    logger.log('🎯 refreshWithClear function called - will clear user list')
    
    try {
      // 前のリクエストをキャンセル
      if (refreshControllerRef.current) {
        logger.log('🛑 Aborting previous refresh request')
        refreshControllerRef.current.abort()
      }

      logger.log('🔄 Refresh with clear starting...', new Date().toLocaleTimeString())
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, refreshing: true }
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
      
      const status = st.status || st.Status || 'WAITING'
      const fetched = Array.isArray(us) ? us : []
      
      setState(prev => {
        const sortedUsers = sortUsersStable(fetched)
        logger.log('📋 Clearing and updating with fresh users:', { count: sortedUsers.length })
        
        return {
          ...prev,
          active: status === 'ACTIVE',
          users: sortedUsers, // 強制的に新しいリストに置き換え
          startTime: st.startedAt,
          errorMsg: ''
        }
      })
      
      logger.log('✅ Refresh with clear completed:', { status, userCount: fetched.length })
    } catch (e) {
      // AbortErrorの場合はエラーメッセージを表示しない
      if (e instanceof Error && e.name === 'AbortError') {
        logger.log('🚫 Refresh with clear aborted')
        return
      }
      
      logger.error('❌ Refresh with clear failed:', e)
      setState(prev => ({
        ...prev,
        errorMsg: '更新に失敗しました。しばらくしてから再試行してください。'
      }))
    } finally {
      updateClock()
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, refreshing: false }
      }))
    }
  }, [updateClock])

  const refresh = useCallback(async () => {
    logger.log('🎯 refresh function called from useAppState')
    
    try {
      // 前のリクエストをキャンセル
      if (refreshControllerRef.current) {
        logger.log('🛑 Aborting previous refresh request')
        refreshControllerRef.current.abort()
      }

      logger.log('🔄 Auto refresh starting...', new Date().toLocaleTimeString())
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, refreshing: true }
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
      
      const status = st.status || st.Status || 'WAITING'
      const fetched = Array.isArray(us) ? us : []
      
      setState(prev => {
        const sortedUsers = sortUsersStable(fetched)
        logger.log('📋 Updating state with users:', { count: sortedUsers.length, firstThree: sortedUsers.slice(0, 3).map(u => u.displayName) })
        
        // ユーザーリスト保持ロジック：
        // 1. サーバーから新しいユーザーがある場合は更新
        // 2. サーバーが空でも既存ユーザーがいれば保持（停止中でもリストを保持）
        const shouldKeepExistingUsers = fetched.length === 0 && prev.users.length > 0
        const finalUsers = shouldKeepExistingUsers ? prev.users : sortedUsers
        
        logger.log('📋 User list decision:', {
          serverUsers: fetched.length,
          existingUsers: prev.users.length,
          keepExisting: shouldKeepExistingUsers,
          finalCount: finalUsers.length
        })
        
        return {
          ...prev,
          active: status === 'ACTIVE',
          users: finalUsers,
          startTime: st.startedAt,
          errorMsg: ''
        }
      })
      
      logger.log('✅ Auto refresh completed:', { status, userCount: fetched.length })
    } catch (e) {
      // AbortErrorの場合はエラーメッセージを表示しない
      if (e instanceof Error && e.name === 'AbortError') {
        logger.log('🚫 Refresh aborted')
        return
      }
      
      logger.error('❌ Auto refresh failed:', e)
      setState(prev => ({
        ...prev,
        errorMsg: '更新に失敗しました。しばらくしてから再試行してください。'
      }))
    } finally {
      updateClock()
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, refreshing: false }
      }))
    }
  }, [updateClock])

  const handleAsyncAction = useCallback(async (
    action: (signal: AbortSignal) => Promise<void>,
    loadingKey: keyof LoadingStates,
    successMsg: string,
    errorMsgPrefix: string = '',
    controllerRef: React.MutableRefObject<AbortController | null>,
    shouldClearUsers: boolean = false // 切替・リセット時のフラグ
  ) => {
    try {
      // 前のリクエストをキャンセル
      if (controllerRef.current) {
        controllerRef.current.abort()
      }

      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, [loadingKey]: true }
      }))
      
      // 新しいAbortControllerを作成
      const controller = new AbortController()
      controllerRef.current = controller

      await action(controller.signal)
      
      // リクエストが成功したらcontrollerをクリア
      controllerRef.current = null
      
      setState(prev => ({ ...prev, errorMsg: '', infoMsg: successMsg }))

      // 取得系アクション（pulling）の場合は取得時刻を更新
      if (loadingKey === 'pulling') {
        const now = new Date()
        const pad = (n: number) => String(n).padStart(2, '0')
        const timeString = `${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
        setState(prev => ({ ...prev, lastFetchTime: timeString }))
      }

      // 切替・リセット時はユーザーリストをクリア、それ以外は保持
      if (shouldClearUsers) {
        await refreshWithClear()
      } else {
        await refresh()
      }
    } catch (e) {
      // AbortErrorの場合はエラーメッセージを表示しない
      if (e instanceof Error && e.name === 'AbortError') {
        logger.log(`🚫 ${loadingKey} action aborted`)
        return
      }
      
      const errorMessage = `${errorMsgPrefix}に失敗しました。${loadingKey === 'switching' ? '配信開始後に再度お試しください。' : ''}`
      setState(prev => ({ ...prev, errorMsg: errorMessage }))
    } finally {
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, [loadingKey]: false }
      }))

    }
  }, [refresh])

  const actions: AppActions = {
    setVideoId: useCallback((value: string) => {
      setState(prev => ({ ...prev, videoId: value }))
    }, []),

    setIntervalSec: useCallback((value: number) => {
      setState(prev => ({ ...prev, intervalSec: value }))
    }, []),

    refresh,

    onSwitch: useCallback(async () => {
      if (!state.videoId) {
        setState(prev => ({ ...prev, errorMsg: 'videoId を入力してください。' }))
        return
      }
      await handleAsyncAction(
        async (signal) => {
          await postSwitchVideo(state.videoId, signal)
          localStorage.setItem('videoId', state.videoId)
        },
        'switching',
        '切替しました',
        '切替',
        switchControllerRef,
        true // 切替時はユーザーリストをクリア
      )
    }, [state.videoId, handleAsyncAction]),

    onPull: useCallback(async () => {
      await handleAsyncAction(
        (signal) => postPull(signal),
        'pulling',
        '取得しました',
        '取得',
        pullControllerRef
      )
    }, [handleAsyncAction]),

    onPullSilent: useCallback(async () => {
      await handleAsyncAction(
        (signal) => postPull(signal),
        'pulling',
        '', // 自動更新時はメッセージなし
        '取得',
        pullControllerRef
      )
    }, [handleAsyncAction]),

    onReset: useCallback(async () => {
      await handleAsyncAction(
        (signal) => postReset(signal),
        'resetting',
        'リセットしました',
        'リセット',
        resetControllerRef,
        true // リセット時もユーザーリストをクリア
      )
    }, [handleAsyncAction]),

    clearInfoMsg: useCallback(() => {
      setState(prev => ({ ...prev, infoMsg: '' }))
    }, [])
  }

  return { state, actions }
}