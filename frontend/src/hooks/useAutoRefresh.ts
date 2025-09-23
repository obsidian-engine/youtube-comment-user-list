import { useEffect, useRef, useCallback } from 'react'
import { logger } from '../utils/logger'

export function useAutoRefresh(intervalSec: number, refresh: () => Promise<void> | void) {
  const isRefreshingRef = useRef(false)
  const intervalIdRef = useRef<number | null>(null)

  // 安全なrefresh関数：重複実行を防ぐ
  const safeRefresh = useCallback(async () => {
    if (isRefreshingRef.current) {
      logger.log('⏭️ Auto refresh skipped (already running)')
      return
    }

    isRefreshingRef.current = true
    try {
      logger.log(`⏰ Auto refresh executing (${intervalSec}s interval)`)
      await Promise.resolve(refresh())
    } catch (error) {
      logger.error('❌ Auto refresh error:', error)
    } finally {
      isRefreshingRef.current = false
    }
  }, [refresh, intervalSec])

  useEffect(() => {
    // 既存のタイマーをクリア
    if (intervalIdRef.current) {
      logger.log('🗑️ Clearing previous auto refresh timer')
      window.clearInterval(intervalIdRef.current)
      intervalIdRef.current = null
    }

    if (!intervalSec) {
      logger.log('🚫 Auto refresh stopped (interval set to 0)')
      return
    }

    logger.log(`⏰ Auto refresh timer set to ${intervalSec} seconds`)
    
    // 新しいタイマーを作成
    intervalIdRef.current = window.setInterval(safeRefresh, intervalSec * 1000)

    return () => {
      if (intervalIdRef.current) {
        logger.log('🗑️ Auto refresh timer cleared on cleanup')
        window.clearInterval(intervalIdRef.current)
        intervalIdRef.current = null
      }
    }
  }, [intervalSec, safeRefresh])
}

