import { useCallback, useEffect, useSyncExternalStore } from 'react'
import { getReplyTargets, putReplyTarget } from '../utils/api'

// 自動返信のフラグの一覧。ユーザー一覧と履歴の一覧で同じ状態を共有するため、module の変数 1 つに格納する。
let targets: ReadonlySet<string> = new Set()
let loaded = false
let error: string | null = null
const listeners = new Set<() => void>()

function emit() {
  listeners.forEach((l) => l())
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

function setTargets(next: ReadonlySet<string>) {
  targets = next
  emit()
}

function setError(next: string | null) {
  error = next
  emit()
}

async function load() {
  try {
    const list = await getReplyTargets()
    loaded = true
    setTargets(new Set(list.map((t) => t.channelId)))
  } catch {
    setError('自動返信の対象を読み込めませんでした')
  }
}

// test から module の状態を初期化する
export function resetReplyTargetsForTest() {
  targets = new Set()
  loaded = false
  error = null
}

export function useReplyTargets() {
  const current = useSyncExternalStore(subscribe, () => targets)
  const currentError = useSyncExternalStore(subscribe, () => error)

  useEffect(() => {
    if (!loaded) void load()
  }, [])

  // 画面を先に切り替え、保存に失敗したら元の状態に戻す
  const toggle = useCallback(async (channelId: string) => {
    const enabled = !targets.has(channelId)
    const before = targets
    const next = new Set(targets)
    if (enabled) next.add(channelId)
    else next.delete(channelId)
    setError(null)
    setTargets(next)
    try {
      await putReplyTarget(channelId, enabled)
    } catch {
      setTargets(before)
      setError('自動返信の設定を保存できませんでした。もう一度試してください')
    }
  }, [])

  return { isTarget: (channelId: string) => current.has(channelId), toggle, error: currentError }
}
