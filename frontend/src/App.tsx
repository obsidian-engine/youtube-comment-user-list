import { useState } from 'react'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { useAppState } from './hooks/useAppState'
import { StatsCard } from './components/StatsCard'
import { QuickGuide } from './components/QuickGuide'
import { Controls } from './components/Controls'
import { UserTable } from './components/UserTable'
import { Toast } from './components/Toast'
import { ThemeToggle } from './components/ThemeToggle'

export default function App() {
  const { state, actions } = useAppState()
  const [showCommentTime, setShowCommentTime] = useState(true)
  const {
    active,
    users,
    videoId,
    intervalSec,
    lastUpdated,
    lastFetchTime,
    errorMsg,
    infoMsg,
    loadingStates,
  } = state

  // デバッグ: Appコンポーネント初期化時のログ
  console.log('🏠 App component rendered:', { 
    intervalSec, 
    active, 
    usersCount: users.length,
    isRefreshing: loadingStates.refreshing 
  })

  useAutoRefresh(intervalSec, actions.onPullSilent)

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <div className="fixed inset-0 -z-10 bg-field" />
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        <div className="flex justify-between items-center">
          <QuickGuide />
          <ThemeToggle />
        </div>
        
        {errorMsg && (
          <div
            role="alert"
            aria-live="assertive"
            className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3"
          >
            {errorMsg}
          </div>
        )}

        <Controls
          videoId={videoId}
          setVideoId={actions.setVideoId}
          loadingStates={loadingStates}
          onSwitch={actions.onSwitch}
          onPull={actions.onPull}
          onReset={actions.onReset}
        />

        {infoMsg && (
          <Toast 
            message={infoMsg}
            type="success"
            onClose={actions.clearInfoMsg}
          />
        )}

        <StatsCard users={users} active={active} startTime={state.startTime} lastUpdated={lastUpdated} />

        <UserTable 
          users={users} 
          intervalSec={intervalSec} 
          setIntervalSec={actions.setIntervalSec} 
          isRefreshing={loadingStates.refreshing}
          showCommentTime={showCommentTime}
          onToggleCommentTime={() => setShowCommentTime(!showCommentTime)}
        />
      </main>
    </div>
  )
}
