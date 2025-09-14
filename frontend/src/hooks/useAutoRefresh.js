import { useEffect } from 'react'

export function useAutoRefresh(intervalSec, refresh) {
  useEffect(() => {
    if (!intervalSec) return
    const id = setInterval(() => {
      refresh().catch(() => {})
    }, intervalSec * 1000)
    return () => clearInterval(id)
  }, [intervalSec, refresh])
}

