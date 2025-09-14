import { useCallback, useEffect, useMemo, useState } from 'react'
import { getStatus, getUsers, postPull, postReset, postSwitchVideo } from './utils/api'
import { useAutoRefresh } from './hooks/useAutoRefresh'

const fmtTime = (d) => (d ? new Date(d).toLocaleString() : '-')

export default function App() {
  const [appState, setAppState] = useState('WAITING') // 'WAITING' | 'ACTIVE'
  const [users, setUsers] = useState([])
  const [mode, setMode] = useState('chip') // 'chip' | 'table'
  const [filter, setFilter] = useState('')
  const [intervalSec, setIntervalSec] = useState(30)
  const [lastUpdated, setLastUpdated] = useState('')
  const [videoId, setVideoId] = useState(() => localStorage.getItem('videoId') || '')

  const filtered = useMemo(() => {
    const q = filter.trim().toLowerCase()
    if (!q) return users
    return users.filter((n) => n.toLowerCase().includes(q))
  }, [users, filter])

  const refresh = useCallback(async () => {
    const ac = new AbortController()
    const [st, us] = await Promise.all([
      getStatus(ac.signal),
      getUsers(ac.signal),
    ])
    setAppState(st.status || st.Status || 'WAITING')
    setUsers(us || [])
    setLastUpdated(new Date().toLocaleString())
  }, [])

  useEffect(() => {
    refresh().catch(() => {})
  }, [refresh])

  useAutoRefresh(intervalSec, refresh)

  const onSwitch = async () => {
    if (!videoId) return
    await postSwitchVideo(videoId)
    localStorage.setItem('videoId', videoId)
    await refresh()
  }
  const onPull = async () => {
    await postPull()
    await refresh()
  }
  const onReset = async () => {
    await postReset()
    await refresh()
  }

  return (
    <div className="max-w-5xl mx-auto p-4 space-y-4">
      <header className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">YouTube コメント参加ユーザー</h1>
        <div className="flex items-center gap-3 text-sm">
          <span className={
            `px-2 py-1 rounded font-medium ${appState==='ACTIVE' ? 'bg-green-100 text-green-700' : 'bg-gray-200 text-gray-700'}`
          }>
            {appState}
          </span>
          <span>人数: {users.length}</span>
          <span>最終更新: {lastUpdated || '-'}</span>
        </div>
      </header>

      <section className="flex flex-wrap gap-2 items-center">
        <input
          value={videoId}
          onChange={(e) => setVideoId(e.target.value)}
          placeholder="videoId"
          className="border rounded px-2 py-1 min-w-[260px]"
        />
        <button onClick={onSwitch} className="px-3 py-1 rounded bg-blue-600 text-white">切替</button>

        <select value={intervalSec} onChange={(e)=>setIntervalSec(Number(e.target.value))} className="border rounded px-2 py-1">
          {[0,3,5,10,30,60].map(s=> <option key={s} value={s}>{s===0?'停止':`${s}s`}</option>)}
        </select>

        <button onClick={onPull} className="px-3 py-1 rounded bg-emerald-600 text-white">今すぐ取得</button>
        <button onClick={onReset} className="px-3 py-1 rounded bg-rose-600 text-white">リセット</button>

        <input
          value={filter}
          onChange={(e)=>setFilter(e.target.value)}
          placeholder="フィルタ（部分一致）"
          className="border rounded px-2 py-1 ml-auto"
        />

        <select value={mode} onChange={(e)=>setMode(e.target.value)} className="border rounded px-2 py-1">
          <option value="chip">チップ</option>
          <option value="table">表</option>
        </select>
      </section>

      {mode === 'chip' ? (
        <ul className="grid grid-cols-chips gap-2">
          {filtered.map((n, i) => (
            <li key={`${n}-${i}`} className="border rounded px-3 py-2 bg-white">
              <span className="truncate-1" title={n}>{n}</span>
            </li>
          ))}
        </ul>
      ) : (
        <table className="w-full bg-white border rounded">
          <thead>
            <tr className="text-left bg-gray-100">
              <th className="px-3 py-2 w-16">#</th>
              <th className="px-3 py-2">名前</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((n, i) => (
              <tr key={`${n}-${i}`} className="border-t">
                <td className="px-3 py-2">{i+1}</td>
                <td className="px-3 py-2"><span className="truncate-1" title={n}>{n}</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {appState === 'WAITING' && users.length === 0 && (
        <div className="p-3 rounded bg-amber-50 text-amber-800">配信が終了しました。次の videoId を入力して「切替」してください。</div>
      )}
    </div>
  )
}
