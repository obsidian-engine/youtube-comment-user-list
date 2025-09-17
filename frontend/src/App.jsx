import { useEffect, useState, useCallback } from 'react'
import { getStatus, getUsers, postPull, postReset, postSwitchVideo } from './utils/api'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { sortUsersStable } from './utils/sortUsers'
import { Header } from './components/Header'
import { QuickGuide } from './components/QuickGuide'
import { Controls } from './components/Controls'
import { UserTable } from './components/UserTable'


export default function App() {
  const [active, setActive] = useState(false) // ACTIVE / WAITING
  const [users, setUsers] = useState([])
  const [videoId, setVideoId] = useState(() => localStorage.getItem('videoId') || '')
  const [intervalSec, setIntervalSec] = useState(30)
  const [lastUpdated, setLastUpdated] = useState('--:--:--')
  const [lastFetchTime, setLastFetchTime] = useState('')
  const [errorMsg, setErrorMsg] = useState('')
  const [infoMsg, setInfoMsg] = useState('')
  const [loadingStates, setLoadingStates] = useState({
    switching: false,
    pulling: false,
    resetting: false,
    refreshing: false
  })

  // 並び順ユーティリティ（TS実装）を使用

  const updateClock = () => {
    const d = new Date();
    const pad = (n) => String(n).padStart(2, '0')
    setLastUpdated(`${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`)
  }

  const refresh = useCallback(async () => {
    try {
      console.log('🔄 Auto refresh starting...', new Date().toLocaleTimeString())
      setLoadingStates(prev => ({ ...prev, refreshing: true }))
      const ac = new AbortController()
      const [st, us] = await Promise.all([
        getStatus(ac.signal),
        getUsers(ac.signal),
      ])
      const status = st.status || st.Status || 'WAITING'
      setActive(status === 'ACTIVE')
      const fetched = Array.isArray(us) ? us : []
      setUsers(sortUsersStable(fetched))
      setErrorMsg('')
      console.log('✅ Auto refresh completed:', { status, userCount: (Array.isArray(us) ? us : []).length })
    } catch (e) {
      console.error('❌ Auto refresh failed:', e)
      setErrorMsg('更新に失敗しました。しばらくしてから再試行してください。')
    } finally {
      updateClock()
      setLoadingStates(prev => ({ ...prev, refreshing: false }))
    }
  }, [])

  // 共通のasyncActionハンドラー
  const handleAsyncAction = useCallback(async (action, loadingKey, successMsg, errorMsgPrefix = '') => {
    try {
      setLoadingStates(prev => ({ ...prev, [loadingKey]: true }))
      await action()
      setErrorMsg('')
      setInfoMsg(successMsg)

      // 取得系アクション（pulling）の場合は取得時刻を更新
      if (loadingKey === 'pulling') {
        const now = new Date()
        const pad = (n) => String(n).padStart(2, '0')
        setLastFetchTime(`最終取得: ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`)
      }

      await refresh()
    } catch(e) {
      setErrorMsg(`${errorMsgPrefix}に失敗しました。${loadingKey === 'switching' ? '配信開始後に再度お試しください。' : ''}`)
    } finally {
      setLoadingStates(prev => ({ ...prev, [loadingKey]: false }))
      setTimeout(() => setInfoMsg(''), 2000)
    }
  }, [refresh])

  const onSwitch = useCallback(async () => {
    if (!videoId) {
      setErrorMsg('videoId を入力してください。');
      return
    }
    await handleAsyncAction(
      async () => {
        await postSwitchVideo(videoId)
        localStorage.setItem('videoId', videoId)
      },
      'switching',
      '切替しました',
      '切替'
    )
  }, [videoId, handleAsyncAction])

  const onPull = useCallback(async () => {
    await handleAsyncAction(
      () => postPull(),
      'pulling',
      '取得しました',
      '取得'
    )
  }, [handleAsyncAction])

  const onReset = useCallback(async () => {
    await handleAsyncAction(
      () => postReset(),
      'resetting',
      'リセットしました',
      'リセット'
    )
  }, [handleAsyncAction])

  useEffect(() => { refresh() }, [refresh])

  useAutoRefresh(intervalSec, refresh)

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <div className="fixed inset-0 -z-10 bg-field" />
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        <Header
          active={active}
          userCount={users.length}
          lastUpdated={lastUpdated}
        />

        <QuickGuide />

        {errorMsg && (
          <div role="alert" aria-live="assertive" className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3">
            {errorMsg}
          </div>
        )}

        <Controls
          videoId={videoId}
          setVideoId={setVideoId}
          intervalSec={intervalSec}
          setIntervalSec={setIntervalSec}
          lastFetchTime={lastFetchTime}
          loadingStates={loadingStates}
          onSwitch={onSwitch}
          onPull={onPull}
          onReset={onReset}
        />

        {infoMsg && (
          <div role="status" aria-live="polite" className="rounded-lg ring-1 ring-sky-300/60 bg-sky-50 text-sky-800 px-4 py-3">
            {infoMsg}
          </div>
        )}

        <UserTable users={users} />
      </main>
    </div>
  )
}
