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
    <section
      aria-label="操作"
      style={{
        background: 'var(--c-bg-2)',
        border: '1px solid var(--c-line-strong)',
        padding: '20px 24px',
      }}
    >
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
            style={{
              flex: 1,
              padding: '9px 12px',
              background: 'var(--c-bg)',
              border: '1px solid var(--c-line-strong)',
              color: 'var(--c-ink)',
              fontFamily: 'var(--f-mono)',
              fontSize: '13px',
              outline: 'none',
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = 'var(--c-accent-2)'
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = 'var(--c-line-strong)'
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
            onClick={onReset}
          >
            リセット
          </LoadingButton>
        </div>
      </div>
    </section>
  )
}
