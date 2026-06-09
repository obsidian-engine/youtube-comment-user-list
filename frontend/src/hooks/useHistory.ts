import { useState, useCallback, useRef } from 'react'
import { getHistorySnapshots, getHistorySnapshot } from '../utils/api'
import type { HistorySummary, HistorySnapshot } from '../utils/api'
import { logger } from '../utils/logger'

export interface UseHistoryReturn {
  snapshots: HistorySummary[]
  selected: HistorySnapshot | null
  loading: boolean
  error: string
  loadList: () => Promise<void>
  select: (videoId: string) => Promise<void>
  clearSelected: () => void
}

export function useHistory(): UseHistoryReturn {
  const [snapshots, setSnapshots] = useState<HistorySummary[]>([])
  const [selected, setSelected] = useState<HistorySnapshot | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const listControllerRef = useRef<AbortController | null>(null)
  const detailControllerRef = useRef<AbortController | null>(null)

  const loadList = useCallback(async () => {
    if (listControllerRef.current) listControllerRef.current.abort()
    const controller = new AbortController()
    listControllerRef.current = controller
    setLoading(true)
    setError('')
    try {
      const items = await getHistorySnapshots(controller.signal)
      setSnapshots(items)
    } catch (e) {
      if (e instanceof Error && e.name === 'AbortError') return
      logger.error('useHistory.loadList failed:', e)
      setError('履歴の取得に失敗しました')
    } finally {
      setLoading(false)
    }
  }, [])

  const select = useCallback(async (videoId: string) => {
    if (detailControllerRef.current) detailControllerRef.current.abort()
    const controller = new AbortController()
    detailControllerRef.current = controller
    setLoading(true)
    setError('')
    try {
      const snap = await getHistorySnapshot(videoId, controller.signal)
      setSelected(snap)
    } catch (e) {
      if (e instanceof Error && e.name === 'AbortError') return
      logger.error('useHistory.select failed:', e)
      setError('スナップショットの取得に失敗しました')
      setSelected(null)
    } finally {
      setLoading(false)
    }
  }, [])

  const clearSelected = useCallback(() => {
    setSelected(null)
  }, [])

  return { snapshots, selected, loading, error, loadList, select, clearSelected }
}
