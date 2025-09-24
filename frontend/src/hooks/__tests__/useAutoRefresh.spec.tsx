import { renderHook, act } from '@testing-library/react'
import { useAutoRefresh } from '../useAutoRefresh'
import { vi, describe, test, expect, beforeEach, afterEach } from 'vitest'
import { logger } from '../../utils/logger'

describe('useAutoRefresh', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.spyOn(logger, 'log').mockImplementation(() => {})
    vi.spyOn(logger, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  test('intervalSec が 0 の場合、タイマーを設定しない', () => {
    const mockRefresh = vi.fn()
    
    renderHook(() => useAutoRefresh(0, mockRefresh))
    
    // 時間を進めてもrefreshが呼ばれないことを確認
    act(() => {
      vi.advanceTimersByTime(10000)
    })
    
    expect(mockRefresh).not.toHaveBeenCalled()
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('Auto refresh stopped'))
  })

  test('intervalSec が正の値の場合、指定間隔でrefreshを実行', () => {
    const mockRefresh = vi.fn()
    
    renderHook(() => useAutoRefresh(5, mockRefresh))
    
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('Auto refresh timer set to 5 seconds'))
    
    // 5秒経過
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockRefresh).toHaveBeenCalledTimes(1)
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('Auto refresh timer set to 5 seconds'))
    
    // さらに5秒経過
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockRefresh).toHaveBeenCalledTimes(2)
  })

  test('actions.onPullが渡された場合、タイマーでonPullが呼ばれる', () => {
    const mockOnPull = vi.fn().mockName('onPull')
    
    renderHook(() => useAutoRefresh(5, mockOnPull))
    
    // 5秒経過でonPullが呼ばれる
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockOnPull).toHaveBeenCalledTimes(1)
    expect(mockOnPull).toHaveBeenCalledWith()
  })

  test('intervalSec が変更されると古いタイマーをクリアして新しいタイマーを設定', () => {
    const mockRefresh = vi.fn()
    
    const { rerender } = renderHook(
      ({ interval }) => useAutoRefresh(interval, mockRefresh),
      { initialProps: { interval: 10 } }
    )
    
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('Auto refresh timer set to 10 seconds'))
    
    // intervalを変更
    rerender({ interval: 5 })
    
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('Clearing previous auto refresh timer'))
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('Auto refresh timer set to 5 seconds'))
    
    // 新しい間隔で動作することを確認
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockRefresh).toHaveBeenCalledTimes(1)
  })

  test('refresh関数が変更されると新しい関数でタイマーを再設定', () => {
    const mockRefresh1 = vi.fn()
    const mockRefresh2 = vi.fn()
    
    const { rerender } = renderHook(
      ({ refresh }) => useAutoRefresh(5, refresh),
      { initialProps: { refresh: mockRefresh1 } }
    )
    
    // 最初の関数で実行
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockRefresh1).toHaveBeenCalledTimes(1)
    expect(mockRefresh2).not.toHaveBeenCalled()
    
    // refresh関数を変更
    rerender({ refresh: mockRefresh2 })
    
    // 新しい関数で実行
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockRefresh1).toHaveBeenCalledTimes(1)
    expect(mockRefresh2).toHaveBeenCalledTimes(1)
  })

  test('refresh関数でエラーが発生してもタイマーは継続', async () => {
    const mockRefreshError = vi.fn().mockRejectedValue(new Error('Refresh failed'))
    
    renderHook(() => useAutoRefresh(5, mockRefreshError))
    
    // エラーが発生してもタイマーは継続
    await act(async () => {
      vi.advanceTimersByTime(5000)
      await vi.runAllTimersAsync()
    })
    
    expect(mockRefreshError).toHaveBeenCalledTimes(1)
    
    // 次の間隔でも実行される
    await act(async () => {
      vi.advanceTimersByTime(5000)
      await vi.runAllTimersAsync()
    })
    
    expect(mockRefreshError).toHaveBeenCalledTimes(2)
  })

  test('コンポーネントアンマウント時にタイマーをクリア', () => {
    const mockRefresh = vi.fn()
    
    const { unmount } = renderHook(() => useAutoRefresh(5, mockRefresh))
    
    unmount()
    
    expect(logger.log).toHaveBeenCalledWith(expect.stringContaining('🗑️ Auto refresh timer cleared on cleanup'))
    
    // アンマウント後は実行されない
    act(() => {
      vi.advanceTimersByTime(10000)
    })
    
    expect(mockRefresh).not.toHaveBeenCalled()
  })

  test('refresh関数がPromiseを返す場合の処理', async () => {
    const mockRefresh = vi.fn().mockResolvedValue(undefined)
    
    renderHook(() => useAutoRefresh(5, mockRefresh))
    
    act(() => {
      vi.advanceTimersByTime(5000)
    })
    
    expect(mockRefresh).toHaveBeenCalledTimes(1)
  })
})