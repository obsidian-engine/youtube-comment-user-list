import { useState, useCallback } from 'react'
import { getStatus, getUsers, postPull, postReset, postSwitchVideo } from '../utils/api'
import { sortUsersStable } from '../utils/sortUsers'
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
  loadingStates: LoadingStates
}

interface AppActions {
  setVideoId: (value: string) => void
  setIntervalSec: (value: number) => void
  refresh: () => Promise<void>
  onSwitch: () => Promise<void>
  onPull: () => Promise<void>
  onReset: () => Promise<void>
}

export function useAppState() {
  const [state, setState] = useState<AppState>({
    active: false,
    users: [],
    videoId: localStorage.getItem('videoId') || '',
    intervalSec: 30,
    lastUpdated: '--:--:--',
    lastFetchTime: '',
    errorMsg: '',
    infoMsg: '',
    loadingStates: {
      switching: false,
      pulling: false,
      resetting: false,
      refreshing: false
    }
  })

  const updateClock = useCallback(() => {
    const d = new Date()
    const pad = (n: number) => String(n).padStart(2, '0')
    const timeString = `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    setState(prev => ({ ...prev, lastUpdated: timeString }))
  }, [])

  const refresh = useCallback(async () => {
    try {
      console.log('🔄 Auto refresh starting...', new Date().toLocaleTimeString())
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, refreshing: true }
      }))
      
      const ac = new AbortController()
      const [st, us] = await Promise.all([
        getStatus(ac.signal),
        getUsers(ac.signal),
      ])
      
      const status = st.status || st.Status || 'WAITING'
      const fetched = Array.isArray(us) ? us : []
      
      setState(prev => ({
        ...prev,
        active: status === 'ACTIVE',
        users: sortUsersStable(fetched),
        errorMsg: ''
      }))
      
      console.log('✅ Auto refresh completed:', { status, userCount: fetched.length })
    } catch (e) {
      console.error('❌ Auto refresh failed:', e)
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
    action: () => Promise<void>,
    loadingKey: keyof LoadingStates,
    successMsg: string,
    errorMsgPrefix: string = ''
  ) => {
    try {
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, [loadingKey]: true }
      }))
      
      await action()
      
      setState(prev => ({ ...prev, errorMsg: '', infoMsg: successMsg }))

      // 取得系アクション（pulling）の場合は取得時刻を更新
      if (loadingKey === 'pulling') {
        const now = new Date()
        const pad = (n: number) => String(n).padStart(2, '0')
        const timeString = `最終取得: ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
        setState(prev => ({ ...prev, lastFetchTime: timeString }))
      }

      await refresh()
    } catch (e) {
      const errorMessage = `${errorMsgPrefix}に失敗しました。${loadingKey === 'switching' ? '配信開始後に再度お試しください。' : ''}`
      setState(prev => ({ ...prev, errorMsg: errorMessage }))
    } finally {
      setState(prev => ({ 
        ...prev, 
        loadingStates: { ...prev.loadingStates, [loadingKey]: false }
      }))
      setTimeout(() => {
        setState(prev => ({ ...prev, infoMsg: '' }))
      }, 2000)
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
        async () => {
          await postSwitchVideo(state.videoId)
          localStorage.setItem('videoId', state.videoId)
        },
        'switching',
        '切替しました',
        '切替'
      )
    }, [state.videoId, handleAsyncAction]),

    onPull: useCallback(async () => {
      await handleAsyncAction(
        () => postPull(),
        'pulling',
        '取得しました',
        '取得'
      )
    }, [handleAsyncAction]),

    onReset: useCallback(async () => {
      await handleAsyncAction(
        () => postReset(),
        'resetting',
        'リセットしました',
        'リセット'
      )
    }, [handleAsyncAction])
  }

  return { state, actions }
}