import { useEffect } from 'react'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { useAppState } from './hooks/useAppState'
import { Header } from './components/Header'
import { QuickGuide } from './components/QuickGuide'
import { Controls } from './components/Controls'
import { UserTable } from './components/UserTable'

export default function App() {
  const { state, actions } = useAppState()
  const {
    active,
    users,
    videoId,
    intervalSec,
    lastUpdated,
    lastFetchTime,
    errorMsg,
    infoMsg,
    loadingStates
  } = state

  useEffect(() => { 
    actions.refresh() 
  }, [actions.refresh])

  useAutoRefresh(intervalSec, actions.refresh)

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
          setVideoId={actions.setVideoId}
          intervalSec={intervalSec}
          setIntervalSec={actions.setIntervalSec}
          lastFetchTime={lastFetchTime}
          loadingStates={loadingStates}
          onSwitch={actions.onSwitch}
          onPull={actions.onPull}
          onReset={actions.onReset}
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