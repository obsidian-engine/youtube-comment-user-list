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
  intervalSec: number
  setIntervalSec: (value: number) => void
  lastFetchTime: string
  loadingStates: LoadingStates
  onSwitch: () => Promise<void>
  onPull: () => Promise<void>
  onReset: () => Promise<void>
}

export function Controls({
  videoId,
  setVideoId,
  intervalSec,
  setIntervalSec,
  lastFetchTime,
  loadingStates,
  onSwitch,
  onPull,
  onReset
}: ControlsProps) {
  return (
    <section className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur" aria-label="操作">
      <div className="p-5 md:p-6">
        <div className="grid gap-3 md:grid-cols-12 items-center">
          <div className="md:col-span-8 flex gap-2.5">
            <label htmlFor="videoId" className="sr-only">videoId</label>
            <input
              id="videoId"
              aria-label="videoId"
              value={videoId}
              onChange={(e) => setVideoId(e.target.value)}
              placeholder="videoId を入力"
              className="flex-1 px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 focus:outline-none focus:ring-2 focus:ring-neutral-400/60 text-[14px]"
              disabled={loadingStates.switching}
            />
            <LoadingButton
              ariaLabel="切替"
              isLoading={loadingStates.switching}
              loadingText="切替中…"
              onClick={onSwitch}
            >切替</LoadingButton>
          </div>
          <div className="md:col-span-4 flex gap-2.5 justify-start md:justify-end">
            <LoadingButton
              ariaLabel="今すぐ取得"
              isLoading={loadingStates.pulling}
              loadingText="取得中…"
              onClick={onPull}
            >今すぐ取得</LoadingButton>
            <LoadingButton
              variant="outline"
              ariaLabel="リセット"
              isLoading={loadingStates.resetting}
              loadingText="リセット中…"
              onClick={onReset}
            >リセット</LoadingButton>
          </div>
        </div>

        <div className="mt-3 text-right">
          <span className="text-sm text-slate-600 dark:text-slate-300" data-testid="last-fetch-time">
            {lastFetchTime}
          </span>
        </div>

        <div className="mt-4 grid gap-3 md:grid-cols-12">
          <div className="md:col-span-3">
            <label htmlFor="interval" className="text-[11px] text-slate-500 dark:text-slate-400 block mb-1">自動間隔</label>
            <select
              id="interval"
              aria-label="自動間隔"
              value={intervalSec}
              onChange={(e) => setIntervalSec(Number(e.target.value))}
              className="w-full px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 text-[14px]"
            >
              <option value="0">停止</option>
              <option value="10">10s</option>
              <option value="15">15s</option>
              <option value="30">30s</option>
              <option value="60">60s</option>
            </select>
          </div>
        </div>
      </div>
    </section>
  )
}