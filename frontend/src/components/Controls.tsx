import { LoadingButton } from './LoadingButton'

interface LoadingStates {
  switching: boolean
  pulling: boolean
  resetting: boolean
  refreshing: boolean
}

interface ControlsProps {
  videoId: string
  setVideoId: (value: string) => void
  status: string
  currentVideoId?: string
  loadingStates: LoadingStates
  onSwitch: () => Promise<void>
  onPull: () => Promise<void>
  onReset: () => Promise<void>
}

export function Controls({
  videoId,
  setVideoId,
  status,
  currentVideoId,
  loadingStates,
  onSwitch,
  onPull,
  onReset,
}: ControlsProps) {
  const isReserved = status === 'RESERVED'
  const isActive = status === 'ACTIVE'
  // ACTIVE 中: 入力 videoId が空 or 現行と一致 → Pull、別 video 入力中 → 開始 (= 切替)
  // WAITING / RESERVED → 開始 (SwitchVideo dispatcher が reserve/switch を自動判別)
  const sameVideo = !videoId || videoId === currentVideoId
  const action: 'pull' | 'start' = isActive && sameVideo ? 'pull' : 'start'
  const actionLabel = action === 'pull' ? '今すぐ取得' : '開始'
  const actionLoading = action === 'pull' ? loadingStates.pulling : loadingStates.switching
  const actionLoadingText = action === 'pull' ? '取得中…' : '開始中…'
  const actionHandler = action === 'pull' ? onPull : onSwitch
  const actionDisabled = isReserved
  return (
    <section aria-label="操作" className="card-editorial">
      <div className="eyebrow">
        OPERATIONS
        <div className="eyebrow__rule" />
      </div>

      <div style={{ padding: '16px 20px 20px' }}>
        <div className="grid gap-3 md:grid-cols-12 items-center">
          <div className="md:col-span-8 flex gap-2">
            <label htmlFor="videoId" className="sr-only">
              videoId
            </label>
            <input
              id="videoId"
              aria-label="videoId"
              value={videoId}
              onChange={(e) => setVideoId(e.target.value)}
              placeholder={isReserved ? '予約中 (キャンセルは curl)' : 'videoId を入力'}
              disabled={actionLoading || actionDisabled}
              autoComplete="off"
              spellCheck={false}
              className="input-rule"
              style={{ flex: 1 }}
              onFocus={(e) => {
                e.currentTarget.style.borderBottomColor = 'var(--c-accent)'
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderBottomColor = 'var(--c-line-strong)'
              }}
            />
            <LoadingButton
              ariaLabel={actionLabel}
              isLoading={actionLoading}
              loadingText={actionLoadingText}
              onClick={actionHandler}
              disabled={actionDisabled}
            >
              {actionLabel}
            </LoadingButton>
          </div>
          <div className="md:col-span-4 flex gap-2 justify-start md:justify-end">
            <LoadingButton
              variant="outline"
              ariaLabel="リセット"
              isLoading={loadingStates.resetting}
              loadingText="リセット中…"
              onClick={() => {
                if (
                  window.confirm(
                    '監視中のユーザーリストをリセットします。この操作は元に戻せません。よろしいですか?',
                  )
                ) {
                  void onReset()
                }
              }}
            >
              リセット
            </LoadingButton>
          </div>
        </div>
      </div>
    </section>
  )
}
