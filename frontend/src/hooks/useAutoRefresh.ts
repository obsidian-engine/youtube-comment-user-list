import { useEffect } from 'react'

export function useAutoRefresh(intervalSec: number, refresh: () => Promise<void> | void) {
  useEffect(() => {
    if (!intervalSec) {
      console.log('🚫 Auto refresh stopped (interval set to 0)')
      return
    }
    console.log(`⏰ Auto refresh timer set to ${intervalSec} seconds`)
    const id = setInterval(() => {
      console.log(`⏰ Auto refresh interval triggered (${intervalSec}s)`)
      Promise.resolve(refresh()).catch(() => {})
    }, intervalSec * 1000)
    return () => {
      console.log('🗑️ Auto refresh timer cleared')
      clearInterval(id)
    }
  }, [intervalSec, refresh])
}

