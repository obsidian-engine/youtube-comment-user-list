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
import { Tabs } from './components/Tabs'
import type { TabType } from './components/Tabs'
import { CommentControls } from './components/CommentTab/CommentControls'
import { CommentList } from './components/CommentTab/CommentList'
import { LogPanel } from './components/LogPanel'
import { HelpPanel } from './components/HelpPanel'
import { HistoryTab } from './components/HistoryTab'
import { PollControls } from './components/PollTab/PollControls'
import { PollResults } from './components/PollTab/PollResults'
import { usePollCount, POLL_INTERVAL_SEC } from './hooks/usePollCount'
import { ErrorBanner } from './components/ErrorBanner'

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
    commentSearch.clearComments()
    pollCount.clearResults()
    await actions.onSwitch()
  }

  // フィルター適用（非表示コメントを除外）
  const visibleComments = (commentSearch.comments || []).filter((c) => !hiddenState.isHidden(c.id))

  // リセット: 表示中の全コメントを非表示
  const handleReset = () => {
    const visibleIds = visibleComments.map((c) => c.id)
    if (visibleIds.length === 0) return
    if (
      !window.confirm(
        `表示中の ${visibleIds.length} 件のコメントを全て非表示にします。よろしいですか?`,
      )
    )
      return
    hiddenState.hideAll(visibleIds)
  }

  const {
    active,
    users,
    videoId,
    intervalSec,
    lastUpdated,
    lastSnapshotAt,
    errorMsg,
    infoMsg,
    snapshotRestoreMsg,
    loadingStates,
  } = state

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
    <div className="min-h-dvh" style={{ background: 'var(--c-bg)', color: 'var(--c-ink)' }}>
      <main className="relative z-10 mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        <div className="flex items-center">
          <Tabs activeTab={activeTab} onTabChange={handleTabChange} />
        </div>

        {errorMsg && <ErrorBanner message={errorMsg} />}

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
              lastSnapshotAt={lastSnapshotAt}
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
            {commentSearch.errorMsg && <ErrorBanner message={commentSearch.errorMsg} />}

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
              lastUpdated={commentSearch.lastUpdated ?? '--:--:--'}
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
            {pollCount.errorMsg && <ErrorBanner message={pollCount.errorMsg} />}
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

        {activeTab === 'history' && <HistoryTab />}

        {activeTab === 'help' && <HelpPanel />}
      </main>
    </div>
  )
}
