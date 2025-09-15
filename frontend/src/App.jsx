import { useEffect, useState, useCallback } from 'react'
import { getStatus, getUsers, postPull, postReset, postSwitchVideo } from './utils/api'
import { useAutoRefresh } from './hooks/useAutoRefresh'
import { LoadingButton } from './components/LoadingButton'


export default function App() {
  const [active, setActive] = useState(false) // ACTIVE / WAITING
  const [users, setUsers] = useState([])
  const [videoId, setVideoId] = useState(() => localStorage.getItem('videoId') || '')
  const [intervalSec, setIntervalSec] = useState(30)
  const [lastUpdated, setLastUpdated] = useState('--:--:--')
  const [errorMsg, setErrorMsg] = useState('')
  const [infoMsg, setInfoMsg] = useState('')
  const [loadingStates, setLoadingStates] = useState({
    switching: false,
    pulling: false,
    resetting: false,
    refreshing: false
  })

  const updateClock = () => {
    const d = new Date();
    const pad = (n) => String(n).padStart(2, '0')
    setLastUpdated(`${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`)
  }

  const refresh = useCallback(async () => {
    try {
      console.log('🔄 Auto refresh starting...', new Date().toLocaleTimeString())
      setLoadingStates(prev => ({ ...prev, refreshing: true }))
      const ac = new AbortController()
      const [st, us] = await Promise.all([
        getStatus(ac.signal),
        getUsers(ac.signal),
      ])
      const status = st.status || st.Status || 'WAITING'
      setActive(status === 'ACTIVE')
      setUsers(Array.isArray(us) ? us : [])
      setErrorMsg('')
      console.log('✅ Auto refresh completed:', { status, userCount: (Array.isArray(us) ? us : []).length })
    } catch (e) {
      console.error('❌ Auto refresh failed:', e)
      setErrorMsg('更新に失敗しました。しばらくしてから再試行してください。')
    } finally {
      updateClock()
      setLoadingStates(prev => ({ ...prev, refreshing: false }))
    }
  }, [])

  // 共通のasyncActionハンドラー
  const handleAsyncAction = useCallback(async (action, loadingKey, successMsg, errorMsgPrefix = '') => {
    try {
      setLoadingStates(prev => ({ ...prev, [loadingKey]: true }))
      await action()
      setErrorMsg('')
      setInfoMsg(successMsg)
      await refresh()
    } catch(e) {
      setErrorMsg(`${errorMsgPrefix}に失敗しました。${loadingKey === 'switching' ? '配信開始後に再度お試しください。' : ''}`)
    } finally {
      setLoadingStates(prev => ({ ...prev, [loadingKey]: false }))
      setTimeout(() => setInfoMsg(''), 2000)
    }
  }, [refresh])

  useEffect(() => { refresh() }, [refresh])

  useAutoRefresh(intervalSec, refresh)

  return (
    <div className="min-h-screen bg-canvas-light dark:bg-canvas-dark text-slate-900 dark:text-slate-100">
      <div className="fixed inset-0 -z-10 bg-field" />
      <main className="mx-auto max-w-4xl px-4 md:px-6 py-6 md:py-10 space-y-6 md:space-y-8">
        {/* Hero */}
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
                  <span className="text-base text-slate-600 dark:text-slate-300/90">最終更新: <span className="font-medium">{lastUpdated}</span></span>
                </div>
              </div>
              <div className="md:col-span-5">
                <div className="rounded-lg ring-1 ring-black/5 dark:ring-white/10 bg-white/70 dark:bg-white/5 backdrop-blur px-5 py-5 md:px-6 md:py-6 flex items-end justify-between">
                  <div className="space-y-0.5">
                    <div className="text-[11px] md:text-xs text-slate-500 dark:text-slate-400 tracking-wide">参加者</div>
                    <div data-testid="counter" className="text-3xl md:text-4xl font-semibold tabular-nums tracking-tight bg-gradient-to-br from-slate-900 to-slate-700 dark:from-white dark:to-slate-300 bg-clip-text text-transparent">{users.length}</div>
                  </div>
                  <div className="h-10 w-px bg-slate-300/50 dark:bg-white/10"></div>
                  <div className="text-[11px] md:text-xs text-slate-500 dark:text-slate-400">状態: <span className="font-medium text-slate-700 dark:text-slate-200">{active ? 'ACTIVE' : 'WAITING'}</span></div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {errorMsg && (
          <div role="alert" aria-live="assertive" className="rounded-lg ring-1 ring-rose-300/60 bg-rose-50 text-rose-800 px-4 py-3">
            {errorMsg}
          </div>
        )}

        {/* Controls */}
        <section className="rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur" aria-label="操作">
          <div className="p-5 md:p-6">
            <div className="grid gap-3 md:grid-cols-12 items-center">
              <div className="md:col-span-8 flex gap-2.5">
                <label htmlFor="videoId" className="sr-only">videoId</label>
                <input
                  id="videoId"
                  aria-label="videoId"
                  value={videoId}
                  onChange={(e)=>setVideoId(e.target.value)}
                  placeholder="videoId を入力"
                  className="flex-1 px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 focus:outline-none focus:ring-2 focus:ring-neutral-400/60 text-[14px]"
                  disabled={loadingStates.switching}
                />
                <LoadingButton
                  ariaLabel="切替"
                  isLoading={loadingStates.switching}
                  loadingText="切替中…"
                  onClick={async ()=>{
                    if(!videoId){ setErrorMsg('videoId を入力してください。'); return }
                    await handleAsyncAction(
                      async () => {
                        await postSwitchVideo(videoId)
                        localStorage.setItem('videoId', videoId)
                      },
                      'switching',
                      '切替しました',
                      '切替'
                    )
                  }}
                >切替</LoadingButton>
              </div>
              <div className="md:col-span-4 flex gap-2.5 justify-start md:justify-end">
                <LoadingButton
                  ariaLabel="今すぐ取得"
                  isLoading={loadingStates.pulling}
                  loadingText="取得中…"
                  onClick={async ()=>{
                    await handleAsyncAction(
                      () => postPull(),
                      'pulling',
                      '取得しました',
                      '取得'
                    )
                  }}
                >今すぐ取得</LoadingButton>
                <LoadingButton
                  variant="outline"
                  ariaLabel="リセット"
                  isLoading={loadingStates.resetting}
                  loadingText="リセット中…"
                  onClick={async ()=>{
                    await handleAsyncAction(
                      () => postReset(),
                      'resetting',
                      'リセットしました',
                      'リセット'
                    )
                  }}
                >リセット</LoadingButton>
              </div>
            </div>

            <div className="mt-4 grid gap-3 md:grid-cols-12">
              <div className="md:col-span-3">
                <label htmlFor="interval" className="text-[11px] text-slate-500 dark:text-slate-400 block mb-1">自動間隔</label>
                <select id="interval" aria-label="自動間隔" value={intervalSec} onChange={(e)=>setIntervalSec(Number(e.target.value))} className="w-full px-3 py-2 rounded-md bg-white/90 dark:bg-white/5 border border-slate-300/80 dark:border-white/10 text-[14px]">
                  <option value="0">停止</option>
                  <option value="10">10s</option>
                  <option value="30">30s</option>
                  <option value="60">60s</option>
                </select>
              </div>
            </div>
          </div>
        </section>

        {infoMsg && (
          <div role="status" aria-live="polite" className="rounded-lg ring-1 ring-sky-300/60 bg-sky-50 text-sky-800 px-4 py-3">
            {infoMsg}
          </div>
        )}

        {/* Table */}
        <section className="overflow-hidden rounded-lg shadow-subtle ring-1 ring-black/5 dark:ring-white/10 bg-white/80 dark:bg-white/5 backdrop-blur">
          <table className="w-full table-auto text-[14px] leading-7">
            <thead className="bg-gradient-to-br from-slate-700 to-slate-800 dark:from-slate-200 dark:to-slate-300 text-white dark:text-slate-900">
              <tr>
                <th className="text-left px-4 py-3.5 w-[72px] font-semibold text-[13px] tracking-wide uppercase">#</th>
                <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">名前</th>
                <th className="text-left px-4 py-3.5 font-semibold text-[13px] tracking-wide uppercase">参加時間</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200/80 dark:divide-white/10">
              {users.map((user,i)=> (
                <tr key={`${user.channelId || user.displayName}-${i}`} className={`transition-colors duration-150 hover:bg-slate-100/60 dark:hover:bg-white/8 ${
                  i % 2 === 0
                    ? 'bg-white/90 dark:bg-white/3'
                    : 'bg-slate-50/80 dark:bg-white/5'
                }`}>
                  <td className="px-4 py-3 tabular-nums text-slate-600 dark:text-slate-300 font-medium">{String(i+1).padStart(2,'0')}</td>
                  <td className="px-4 py-3 truncate-1 text-slate-900 dark:text-slate-100 font-medium" title={user.displayName || user}>{user.displayName || user}</td>
                  <td className="px-4 py-3 text-slate-600 dark:text-slate-300 font-mono text-[13px]">
                    {user.joinedAt ? new Date(user.joinedAt).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' }) : '--:--'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {users.length===0 && (
            <p className="px-4 py-5 text-[13px] text-slate-500 dark:text-slate-400">ユーザーがいません。</p>
          )}
        </section>
      </main>
    </div>
  )
}
