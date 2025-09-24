interface HeaderProps {
  active: boolean
  userCount: number
}

export function Header({ active, userCount }: HeaderProps) {
  return (
    <section className="relative overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur" aria-label="ヘッダー">
      <div className="p-5 md:p-7">
        <div className="grid md:grid-cols-12 gap-6 items-end">
          <div className="md:col-span-7 space-y-1.5">
            <h1 className="text-lg md:text-xl font-semibold tracking-[-0.01em]">
              YouTube Live — <span className="bg-gradient-to-br from-slate-900 to-slate-600 dark:from-white dark:to-slate-300 bg-clip-text text-transparent">参加ユーザー</span>
            </h1>
            <p className="text-xs md:text-sm text-slate-600 dark:text-slate-300/90">
              配信中に参加したユーザーを収集し、終了時点で全員が表示されることを目指します。
            </p>
            <div className="flex items-center gap-4 pt-2">
              <span className={
                `inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-base font-medium ${active ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'border-amber-500/30 bg-amber-400/10 text-amber-800 dark:text-amber-300'}`
              }>
                <span className={`h-2 w-2 rounded-full ${active ? 'bg-emerald-500 shadow-[0_0_0_3px_rgba(16,185,129,.15)]' : 'bg-amber-500 shadow-[0_0_0_3px_rgba(245,158,11,.15)]'}`}></span>
                <span className="tracking-wide">{active ? 'ACTIVE' : 'WAITING'}</span>
              </span>

            </div>
          </div>
          <div className="md:col-span-5">
            <div className="rounded-lg ring-1 ring-black/5 dark:ring-white/10 bg-white/70 dark:bg-white/5 backdrop-blur px-5 py-5 md:px-6 md:py-6 flex items-end justify-between">
              <div className="space-y-0.5">
                <div className="text-[11px] md:text-xs text-slate-500 dark:text-slate-400 tracking-wide">参加者</div>
                <div data-testid="counter" className="text-3xl md:text-4xl font-semibold tabular-nums tracking-tight bg-gradient-to-br from-slate-900 to-slate-700 dark:from-white dark:to-slate-300 bg-clip-text text-transparent">{userCount}</div>
              </div>
              <div className="h-10 w-px bg-slate-300/50 dark:bg-white/10"></div>
              <div className="text-[11px] md:text-xs text-slate-500 dark:text-slate-400">状態: <span className="font-medium text-slate-700 dark:text-slate-200">{active ? 'ACTIVE' : 'WAITING'}</span></div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}