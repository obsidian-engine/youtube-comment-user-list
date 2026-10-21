import { useEffect, useState, type FormEvent, type ReactNode } from 'react'
import { checkSession, login } from '../utils/api'
import { ErrorBanner } from './ErrorBanner'

interface Props {
  children: ReactNode
}

type GateState = 'checking' | 'locked' | 'unlocked'

// 合言葉で session を作るまで子要素を表示しない。
// npm run dev では vite の proxy が API key を header に追加するため、確認を省く。
export function LoginGate({ children }: Props) {
  const [state, setState] = useState<GateState>(import.meta.env.DEV ? 'unlocked' : 'checking')
  const [passphrase, setPassphrase] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (state !== 'checking') return
    const controller = new AbortController()
    checkSession(controller.signal)
      .then((ok) => setState(ok ? 'unlocked' : 'locked'))
      .catch((e: unknown) => {
        if (e instanceof Error && e.name === 'AbortError') return
        setError('ログイン状態を確認できませんでした')
        setState('locked')
      })
    return () => controller.abort()
  }, [state])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      if (await login(passphrase)) {
        setPassphrase('')
        setState('unlocked')
      } else {
        setError('合言葉が違います')
      }
    } catch {
      setError('ログインに失敗しました')
    } finally {
      setSubmitting(false)
    }
  }

  if (state === 'unlocked') return <>{children}</>
  if (state === 'checking') return null

  return (
    <main className="mx-auto mt-24 max-w-sm px-4">
      <form onSubmit={handleSubmit} className="space-y-3">
        <label htmlFor="passphrase" className="block text-sm">
          合言葉
        </label>
        <input
          id="passphrase"
          type="password"
          autoComplete="current-password"
          value={passphrase}
          onChange={(e) => setPassphrase(e.target.value)}
          className="w-full border px-3 py-2"
          required
        />
        <button type="submit" disabled={submitting} className="w-full border px-3 py-2">
          ログイン
        </button>
        {error && <ErrorBanner message={error} />}
      </form>
    </main>
  )
}
