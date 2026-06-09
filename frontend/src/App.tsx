import { useState } from 'react'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { useAppState } from './hooks/useAppState'
import { useCommentSearch } from './hooks/useCommentSearch'
import { useCheckState } from './hooks/useCheckState'
import { useHiddenState } from './hooks/useHiddenState'
import { useLogEntries } from './hooks/useLogEntries'
import { StatsCard } from './components/StatsCard'
import { Controls } from './components/Controls'
import { UserTable } from './components/UserTable'
import { Toast } from './components/Toast'
import { ThemeToggle } from './components/ThemeToggle'
import { Tabs } from './components/Tabs'
import type { TabType } from './components/Tabs'
import { CommentControls } from './components/CommentTab/CommentControls'
import { CommentList } from './components/CommentTab/CommentList'
import { LogPanel } from './components/LogPanel'
import { HelpPanel } from './components/HelpPanel'
import { PollControls } from './components/PollTab/PollControls'
import { PollResults } from './components/PollTab/PollResults'
import { usePollCount, POLL_INTERVAL_SEC } from './hooks/usePollCount'

export default function App() {
  const logEntries = useLogEntries()
  const { state, actions } = useAppState(logEntries.addEntry)
  const commentSearch = useCommentSearch()
  const checkState = useCheckState()
  const hiddenState = useHiddenState()
  const pollCount = usePollCount()
  const [showCommentTime, setShowCommentTime] = useState(true)
  const [activeTab, setActiveTab] = useState<TabType>(() => {
    const saved = localStorage.getItem('activeTab')
    return (saved as TabType) || 'users'
  })

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab)
    localStorage.setItem('activeTab', tab)
  }

  const handleSwitch = async () => {
    logEntries.clear()
    await actions.onSwitch()
  }

  // フィルター適用（非表示コメントを除外）
  const visibleComments = (commentSearch.comments || []).filter((c) => !hiddenState.isHidden(c.id))

  // リセット: 表示中の全コメントを非表示
  const handleReset = () => {
    const visibleIds = visibleComments.map((c) => c.id)
    hiddenState.hideAll(visibleIds)
  }

  const {
    active,
    users,
    videoId,
    intervalSec,
    lastUpdated,
    errorMsg,
    infoMsg,
    snapshotRestoreMsg,
    loadingStates,
  } = state

  // デバッグログはテスト環境では無効化

  // 名前読み上げタブの自動更新
  useAutoRefresh(intervalSec, actions.onPullSilent)

  // コメント検索タブの自動更新
  useAutoRefresh(commentSearch.intervalSec, commentSearch.search)

  // 投票タブの自動更新（15秒間隔、votes タブ表示中かつキーワード設定済みの場合のみ）
  useAutoRefresh(
    activeTab === 'votes' && pollCount.keywords.length > 0 ? POLL_INTERVAL_SEC : 0,
    pollCount.recount,
  )

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <div className="fixed inset-0 -z-10 bg-field" />
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-4">
            <Tabs activeTab={activeTab} onTabChange={handleTabChange} />
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

        {activeTab === 'users' && (
          <>
            <Controls
              videoId={videoId}
              setVideoId={actions.setVideoId}
              loadingStates={loadingStates}
              onSwitch={handleSwitch}
              onPull={actions.onPull}
              onReset={actions.onReset}
            />

            {infoMsg && <Toast message={infoMsg} type="success" onClose={actions.clearInfoMsg} />}
            {snapshotRestoreMsg && (
              <Toast
                message={snapshotRestoreMsg}
                type="info"
                duration={5000}
                onClose={actions.clearSnapshotRestoreMsg}
              />
            )}

            <StatsCard
              users={users}
              active={active}
              startTime={state.startTime}
              lastUpdated={lastUpdated}
              skippedCount={state.skippedCount}
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
        )}

        {activeTab === 'comments' && (
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
              onReset={handleReset}
              isLoading={commentSearch.isLoading}
              intervalSec={commentSearch.intervalSec}
              setIntervalSec={commentSearch.setIntervalSec}
              commentsCount={commentSearch.comments?.length ?? 0}
              checkedCount={checkState.checkedCount}
              lastUpdated={commentSearch.lastUpdated}
            />

            <CommentList
              comments={visibleComments}
              isChecked={checkState.isChecked}
              onToggle={checkState.toggle}
              isLoading={commentSearch.isLoading}
            />
          </>
        )}

        {activeTab === 'votes' && (
          <>
            {pollCount.errorMsg && (
              <div
                role="alert"
                aria-live="assertive"
                className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3"
              >
                {pollCount.errorMsg}
              </div>
            )}
            <PollControls
              keywords={pollCount.keywords}
              onAddKeyword={pollCount.addKeyword}
              onRemoveKeyword={pollCount.removeKeyword}
              onClear={pollCount.clearKeywords}
              onRecount={pollCount.recount}
              isLoading={pollCount.isLoading}
              lastUpdated={pollCount.lastUpdated}
            />
            <PollResults
              keywords={pollCount.keywords}
              counts={pollCount.counts}
              voters={pollCount.voters}
              totalVotes={pollCount.totalVotes}
              isLoading={pollCount.isLoading}
            />
          </>
        )}

        {activeTab === 'logs' && (
          <LogPanel entries={logEntries.entries} onClear={logEntries.clear} />
        )}

        {activeTab === 'help' && <HelpPanel />}
      </main>
    </div>
  )
}
