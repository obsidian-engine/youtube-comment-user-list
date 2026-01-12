import { useState } from 'react'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { useAppState } from './hooks/useAppState'
import { useCommentSearch } from './hooks/useCommentSearch'
import { useCheckState } from './hooks/useCheckState'
import { StatsCard } from './components/StatsCard'
import { QuickGuide } from './components/QuickGuide'
import { Controls } from './components/Controls'
import { UserTable } from './components/UserTable'
import { Toast } from './components/Toast'
import { ThemeToggle } from './components/ThemeToggle'
import { Tabs } from './components/Tabs'
import { CommentControls } from './components/CommentTab/CommentControls'
import { CommentList } from './components/CommentTab/CommentList'

type TabType = 'users' | 'comments'

export default function App() {
  const { state, actions } = useAppState()
  const commentSearch = useCommentSearch()
  const checkState = useCheckState()
  const [showCommentTime, setShowCommentTime] = useState(true)
  const [activeTab, setActiveTab] = useState<TabType>(() => {
    const saved = localStorage.getItem('activeTab')
    return (saved as TabType) || 'users'
  })

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab)
    localStorage.setItem('activeTab', tab)
  }
  const { active, users, videoId, intervalSec, lastUpdated, errorMsg, infoMsg, loadingStates } =
    state

  // デバッグログはテスト環境では無効化

  // 名前読み上げタブの自動更新
  useAutoRefresh(intervalSec, actions.onPullSilent)

  // コメント検索タブの自動更新
  useAutoRefresh(commentSearch.intervalSec, commentSearch.search)

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <div className="fixed inset-0 -z-10 bg-field" />
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-4">
            <Tabs activeTab={activeTab} onTabChange={handleTabChange} />
            {activeTab === 'users' && <QuickGuide />}
          </div>
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

        {activeTab === 'users' ? (
          <>
            <Controls
              videoId={videoId}
              setVideoId={actions.setVideoId}
              loadingStates={loadingStates}
              onSwitch={actions.onSwitch}
              onPull={actions.onPull}
              onReset={actions.onReset}
            />

            {infoMsg && <Toast message={infoMsg} type="success" onClose={actions.clearInfoMsg} />}

            <StatsCard
              users={users}
              active={active}
              startTime={state.startTime}
              lastUpdated={lastUpdated}
            />

            <UserTable
              users={users}
              intervalSec={intervalSec}
              setIntervalSec={actions.setIntervalSec}
              isRefreshing={loadingStates.refreshing}
              showCommentTime={showCommentTime}
              onToggleCommentTime={() => setShowCommentTime(!showCommentTime)}
            />
          </>
        ) : (
          <>
            {commentSearch.errorMsg && (
              <div
                role="alert"
                aria-live="assertive"
                className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3"
              >
                {commentSearch.errorMsg}
              </div>
            )}

            <CommentControls
              keywords={commentSearch.keywords}
              onAddKeyword={commentSearch.addKeyword}
              onRemoveKeyword={commentSearch.removeKeyword}
              onSearch={commentSearch.search}
              onClearChecked={checkState.clear}
              isLoading={commentSearch.isLoading}
              intervalSec={commentSearch.intervalSec}
              setIntervalSec={commentSearch.setIntervalSec}
              commentsCount={commentSearch.comments.length}
              checkedCount={checkState.checkedCount}
              lastUpdated={commentSearch.lastUpdated}
            />

            <CommentList
              comments={commentSearch.comments}
              isChecked={checkState.isChecked}
              onToggle={checkState.toggle}
              isLoading={commentSearch.isLoading}
            />
          </>
        )}
      </main>
    </div>
  )
}
