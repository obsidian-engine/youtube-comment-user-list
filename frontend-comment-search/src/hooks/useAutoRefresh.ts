import { useEffect, useRef, useCallback } from 'react'

export function useAutoRefresh(intervalSec: number, refresh: () => Promise<void> | void) {
  const isRefreshingRef = useRef(false)
  const intervalIdRef = useRef<number | null>(null)

  const safeRefresh = useCallback(async () => {
    if (isRefreshingRef.current) return

    isRefreshingRef.current = true
    try {
      await Promise.resolve(refresh())
    } finally {
      isRefreshingRef.current = false
    }
  }, [refresh])

  useEffect(() => {
    if (intervalIdRef.current) {
      window.clearInterval(intervalIdRef.current)
      intervalIdRef.current = null
    }

    if (!intervalSec) return

    intervalIdRef.current = window.setInterval(safeRefresh, intervalSec * 1000)

    return () => {
      if (intervalIdRef.current) {
        window.clearInterval(intervalIdRef.current)
        intervalIdRef.current = null
      }
    }
  }, [intervalSec, safeRefresh])
}
