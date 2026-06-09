import { renderHook, act } from '@testing-library/react'
import { useAppState } from '../useAppState'
import { vi, describe, test, expect, beforeEach, afterEach } from 'vitest'

// モックAPI関数
const mockGetStatus = vi.fn()
const mockGetUsers = vi.fn()
const mockPostSwitchVideo = vi.fn()
const mockPostPull = vi.fn()
const mockPostReset = vi.fn()

vi.mock('../../utils/api', () => ({
  getStatus: (...args: unknown[]) => mockGetStatus(...args),
  getUsers: (...args: unknown[]) => mockGetUsers(...args),
  postSwitchVideo: (...args: unknown[]) => mockPostSwitchVideo(...args),
  postPull: (...args: unknown[]) => mockPostPull(...args),
  postReset: (...args: unknown[]) => mockPostReset(...args),
}))

describe('useAppState', () => {
  let mockLocalStorage: { [key: string]: string } = {}

  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage = {}

    // 実際に動作するlocalStorageモックを作成
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn((key: string) => mockLocalStorage[key] || null),
        setItem: vi.fn((key: string, value: string) => {
          mockLocalStorage[key] = value
        }),
        removeItem: vi.fn((key: string) => {
          delete mockLocalStorage[key]
        }),
        clear: vi.fn(() => {
          mockLocalStorage = {}
        }),
      },
      writable: true,
    })

    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  test('初期状態が正しく設定される', () => {
    const { result } = renderHook(() => useAppState())

    expect(result.current.state).toEqual({
      active: false,
      users: [],
      videoId: '',
      intervalSec: 60,
      lastUpdated: '--:--:--',
      lastFetchTime: '',
      errorMsg: '',
      infoMsg: '',
      snapshotRestoreMsg: '',
      lastSnapshotAt: '',
      startTime: undefined,
      skippedCount: 0,
      loadingStates: {
        switching: false,
        pulling: false,
        resetting: false,
        refreshing: false,
      },
    })
  })

  test('localStorage からvideoIdを復元する', () => {
    localStorage.setItem('videoId', 'test-video-id')

    const { result } = renderHook(() => useAppState())

    expect(result.current.state.videoId).toBe('test-video-id')
  })

  test('refresh関数がAPIを呼び出して状態を更新する', async () => {
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE', count: 5 })
    mockGetUsers.mockResolvedValue([
      { channelId: 'UC1', displayName: 'User1', joinedAt: '2024-01-01T09:00:00.000Z' },
    ])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.refresh()
    })

    expect(mockGetStatus).toHaveBeenCalled()
    expect(mockGetUsers).toHaveBeenCalled()
    expect(result.current.state.active).toBe(true)
    expect(result.current.state.users).toHaveLength(1)
    expect(result.current.state.errorMsg).toBe('')
  })

  test('refresh関数でエラーが発生した場合、エラーメッセージを設定する', async () => {
    mockGetStatus.mockRejectedValue(new Error('API Error'))

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.refresh()
    })

    expect(result.current.state.errorMsg).toBe(
      '更新に失敗しました。しばらくしてから再試行してください。',
    )
  })

  test('setVideoId関数が状態を更新する', () => {
    const { result } = renderHook(() => useAppState())

    act(() => {
      result.current.actions.setVideoId('new-video-id')
    })

    expect(result.current.state.videoId).toBe('new-video-id')
  })

  test('setIntervalSec関数が状態を更新する', () => {
    const { result } = renderHook(() => useAppState())

    act(() => {
      result.current.actions.setIntervalSec(60)
    })

    expect(result.current.state.intervalSec).toBe(60)
  })

  test('onSwitch関数がvideoIdなしでエラーメッセージを設定する', async () => {
    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.onSwitch()
    })

    expect(result.current.state.errorMsg).toBe('videoId を入力してください。')
    expect(mockPostSwitchVideo).not.toHaveBeenCalled()
  })

  test('onSwitch関数がvideoIdありで成功する', async () => {
    mockPostSwitchVideo.mockResolvedValue(undefined)
    mockPostPull.mockResolvedValue({ addedCount: 0, skippedCount: 0, autoReset: false })
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    act(() => {
      result.current.actions.setVideoId('test-video')
    })

    await act(async () => {
      await result.current.actions.onSwitch()
    })

    expect(mockPostSwitchVideo).toHaveBeenCalledWith('test-video', expect.any(AbortSignal))
    expect(localStorage.getItem('videoId')).toBe('test-video')
    expect(result.current.state.infoMsg).toBe('切替しました')
  })

  test('onPull関数が成功し、lastFetchTimeを更新する', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-01T10:15:30'))

    mockPostPull.mockResolvedValue({
      addedCount: 0,
      skippedCount: 0,
      autoReset: false,
      pollingIntervalMillis: 15000,
    })
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.onPull()
    })

    expect(mockPostPull).toHaveBeenCalled()
    expect(result.current.state.lastFetchTime).toBe('10:15:30')
    expect(result.current.state.infoMsg).toBe('取得しました')

    vi.useRealTimers()
  })

  test('onReset関数が成功する', async () => {
    mockPostReset.mockResolvedValue(undefined)
    mockGetStatus.mockResolvedValue({ status: 'WAITING' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.onReset()
    })

    expect(mockPostReset).toHaveBeenCalled()
    expect(result.current.state.infoMsg).toBe('リセットしました')
  })

  test('ローディング状態が正しく管理される', async () => {
    type ResolveType = (value: {
      addedCount: number
      skippedCount: number
      autoReset: boolean
      pollingIntervalMillis: number
    }) => void
    let resolvePromise: ResolveType
    const delayedPromise = new Promise<{
      addedCount: number
      skippedCount: number
      autoReset: boolean
      pollingIntervalMillis: number
    }>((resolve) => {
      resolvePromise = resolve
    })

    mockPostPull.mockReturnValue(delayedPromise)
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    // onPullを開始（まだ完了しない）
    const pullPromise = result.current.actions.onPull()

    // 少し待ってからローディング状態をチェック
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 0))
    })

    // ローディング中
    expect(result.current.state.loadingStates.pulling).toBe(true)

    // プロミスを解決
    act(() => {
      resolvePromise!({
        addedCount: 0,
        skippedCount: 0,
        autoReset: false,
        pollingIntervalMillis: 15000,
      })
    })

    await act(async () => {
      await pullPromise
    })

    // ローディング完了
    expect(result.current.state.loadingStates.pulling).toBe(false)
  })

  test('API更新間隔を変更すると画面更新間隔も同じ値に設定される', () => {
    const { result } = renderHook(() => useAppState())

    // 初期値の確認
    expect(result.current.state.intervalSec).toBe(60)

    // API更新間隔を60秒に変更
    act(() => {
      result.current.actions.setIntervalSec(60)
    })

    // 画面更新間隔も同じ値になることを確認
    expect(result.current.state.intervalSec).toBe(60)
  })

  test('API更新間隔を0(停止)に設定すると画面更新間隔も停止される', () => {
    const { result } = renderHook(() => useAppState())

    // 最初に値を設定
    act(() => {
      result.current.actions.setIntervalSec(30)
    })
    expect(result.current.state.intervalSec).toBe(30)

    // 0に設定
    act(() => {
      result.current.actions.setIntervalSec(0)
    })

    // 画面更新間隔も0になることを確認
    expect(result.current.state.intervalSec).toBe(0)
  })

  test('onPullSilentは成功メッセージを表示しない', async () => {
    mockPostPull.mockResolvedValue({
      addedCount: 0,
      skippedCount: 0,
      autoReset: false,
      pollingIntervalMillis: 15000,
    })
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.onPullSilent()
    })

    expect(mockPostPull).toHaveBeenCalled()
    expect(result.current.state.infoMsg).toBe('') // メッセージなし
    expect(result.current.state.errorMsg).toBe('')
  })

  test('snapshotSavedAt がある場合 snapshotRestoreMsg をセットする (今日の場合 HH:MM 形式)', async () => {
    vi.useFakeTimers()
    // 今日を 2024-06-09 10:00 (ローカル時刻) に固定
    vi.setSystemTime(new Date(2024, 5, 9, 10, 0, 0))

    // sessionStorage をリセット
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: vi.fn(() => null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    })

    // ローカル時刻 14:23 に相当する ISO 文字列を使う
    const savedAt = new Date(2024, 5, 9, 14, 23, 0)
    mockGetStatus.mockResolvedValue({
      status: 'ACTIVE',
      count: 0,
      snapshotSavedAt: savedAt.toISOString(),
    })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.refresh()
    })

    // HH:MM 形式 (今日のため)
    expect(result.current.state.snapshotRestoreMsg).toContain('14:23')
    expect(result.current.state.snapshotRestoreMsg).toContain('の保存状態を取得しました')

    vi.useRealTimers()
  })

  test('snapshotSavedAt がない場合 snapshotRestoreMsg はセットされない', async () => {
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE', count: 0 })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.refresh()
    })

    expect(result.current.state.snapshotRestoreMsg).toBe('')
  })

  test('onPullは成功メッセージを表示する', async () => {
    mockPostPull.mockResolvedValue({
      addedCount: 0,
      skippedCount: 0,
      autoReset: false,
      pollingIntervalMillis: 15000,
    })
    mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.onPull()
    })

    expect(mockPostPull).toHaveBeenCalled()
    expect(result.current.state.infoMsg).toBe('取得しました') // メッセージあり
    expect(result.current.state.errorMsg).toBe('')
  })

  test('snapshotSavedAt がある場合 lastSnapshotAt が HH:MM 形式で更新される (今日)', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2024, 5, 9, 10, 0, 0))

    const savedAt = new Date(2024, 5, 9, 14, 23, 0)
    mockGetStatus.mockResolvedValue({
      status: 'ACTIVE',
      count: 0,
      snapshotSavedAt: savedAt.toISOString(),
    })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.refresh()
    })

    expect(result.current.state.lastSnapshotAt).toBe('14:23')

    vi.useRealTimers()
  })

  test('snapshotSavedAt がない場合 lastSnapshotAt は前の値を維持する', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2024, 5, 9, 10, 0, 0))

    const savedAt = new Date(2024, 5, 9, 14, 23, 0)
    // 1 回目: snapshotSavedAt あり
    mockGetStatus.mockResolvedValueOnce({
      status: 'ACTIVE',
      count: 0,
      snapshotSavedAt: savedAt.toISOString(),
    })
    mockGetUsers.mockResolvedValueOnce([])
    // 2 回目: snapshotSavedAt なし
    mockGetStatus.mockResolvedValueOnce({ status: 'ACTIVE', count: 0 })
    mockGetUsers.mockResolvedValueOnce([])

    const { result } = renderHook(() => useAppState())

    await act(async () => {
      await result.current.actions.refresh()
    })
    expect(result.current.state.lastSnapshotAt).toBe('14:23')

    await act(async () => {
      await result.current.actions.refresh()
    })
    // snapshotSavedAt がなくても前の値が維持される
    expect(result.current.state.lastSnapshotAt).toBe('14:23')

    vi.useRealTimers()
  })
})
