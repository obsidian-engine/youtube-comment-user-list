import { useEffect, useRef, useCallback } from 'react'
import { logger } from '../utils/logger'

export function useAutoRefresh(intervalSec: number, refresh: () => Promise<void> | void) {
  const isRefreshingRef = useRef(false)
  const intervalIdRef = useRef<number | null>(null)

  // デバッグ：初期化時にログ出力
  logger.log(`🔧 useAutoRefresh initialized with intervalSec: ${intervalSec}`)

  // 安全なrefresh関数：重複実行を防ぐ
  const safeRefresh = useCallback(async () => {
    logger.log(`🚀 safeRefresh called - isRefreshing: ${isRefreshingRef.current}`)
    
    if (isRefreshingRef.current) {
      logger.log('⏭️ Auto refresh skipped (already running)')
      return
    }

    isRefreshingRef.current = true
    try {
      logger.log(`⏰ Auto refresh executing (${intervalSec}s interval)`)
      const result = await Promise.resolve(refresh())
      logger.log(`✅ Auto refresh completed successfully`)
      return result
    } catch (error) {
      logger.error('❌ Auto refresh error:', error)
    } finally {
      isRefreshingRef.current = false
      logger.log(`🏁 safeRefresh finished - isRefreshing reset to false`)
    }
  }, [refresh, intervalSec])

  useEffect(() => {
    logger.log(`🔄 useAutoRefresh useEffect triggered - intervalSec: ${intervalSec}, currentTimer: ${intervalIdRef.current}`)
    
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
    logger.log(`📋 Timer will call safeRefresh every ${intervalSec * 1000}ms`)
    
    // 新しいタイマーを作成
    intervalIdRef.current = window.setInterval(() => {
      logger.log(`⏰ Timer fired! Calling safeRefresh...`)
      safeRefresh()
    }, intervalSec * 1000)
    
    logger.log(`✅ Timer created with ID: ${intervalIdRef.current}`)

    return () => {
      if (intervalIdRef.current) {
        logger.log(`🗑️ Auto refresh timer cleared on cleanup (ID: ${intervalIdRef.current})`)
        window.clearInterval(intervalIdRef.current)
        intervalIdRef.current = null
      }
    }
  }, [intervalSec, safeRefresh])
}

