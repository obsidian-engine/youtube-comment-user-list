import { useEffect } from 'react'

export function useAutoRefresh(intervalSec: number, refresh: () => Promise<void> | void) {
  useEffect(() => {
    if (!intervalSec) return
    const id = setInterval(() => {
      Promise.resolve(refresh()).catch(() => {})
    }, intervalSec * 1000)
    return () => clearInterval(id)
  }, [intervalSec, refresh])
}

