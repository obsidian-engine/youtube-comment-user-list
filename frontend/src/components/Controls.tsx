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
  loadingStates: LoadingStates
  onSwitch: () => Promise<void>
  onPull: () => Promise<void>
  onReset: () => Promise<void>
}

export function Controls({
  videoId,
  setVideoId,
  loadingStates,
  onSwitch,
  onPull,
  onReset,
}: ControlsProps) {
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
              placeholder="videoId を入力"
              disabled={loadingStates.switching}
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
              ariaLabel="切替"
              isLoading={loadingStates.switching}
              loadingText="切替中…"
              onClick={onSwitch}
            >
              切替
            </LoadingButton>
          </div>
          <div className="md:col-span-4 flex gap-2 justify-start md:justify-end">
            <LoadingButton
              ariaLabel="今すぐ取得"
              isLoading={loadingStates.pulling}
              loadingText="取得中…"
              onClick={onPull}
            >
              今すぐ取得
            </LoadingButton>
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
