import { renderHook, act } from '@testing-library/react'
import { useAppState } from '../useAppState'
import { vi, describe, test, expect, beforeEach, afterEach } from 'vitest'
import { BackendError } from '../../utils/api'

// モックAPI関数
const mockGetStatus = vi.fn()
const mockGetUsers = vi.fn()
const mockPostSwitchVideo = vi.fn()
const mockPostPull = vi.fn()
const mockPostReset = vi.fn()

vi.mock('../../utils/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../utils/api')>()
  return {
    ...actual,
    getStatus: (...args: unknown[]) => mockGetStatus(...args),
    getUsers: (...args: unknown[]) => mockGetUsers(...args),
    postSwitchVideo: (...args: unknown[]) => mockPostSwitchVideo(...args),
    postPull: (...args: unknown[]) => mockPostPull(...args),
    postReset: (...args: unknown[]) => mockPostReset(...args),
  }
})

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
      status: 'WAITING',
      active: false,
      reserved: false,
      users: [],
      videoId: '',
      currentVideoId: undefined,
      intervalSec: 60,
      lastUpdated: '--:--:--',
      lastFetchTime: '',
      errorMsg: '',
      infoMsg: '',
      snapshotRestoreMsg: '',
      lastSnapshotAt: '',
      startTime: undefined,
      scheduledStartTime: undefined,
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
    mockPostSwitchVideo.mockResolvedValue({ status: 'ACTIVE' })
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

  test('onSwitch関数がRESERVEDの場合は予約メッセージを表示しpullしない', async () => {
    mockPostSwitchVideo.mockResolvedValue({ status: 'RESERVED' })
    mockGetStatus.mockResolvedValue({ status: 'RESERVED', videoId: 'test-video' })
    mockGetUsers.mockResolvedValue([])

    const { result } = renderHook(() => useAppState())

    act(() => {
      result.current.actions.setVideoId('test-video')
    })

    await act(async () => {
      await result.current.actions.onSwitch()
    })

    expect(mockPostSwitchVideo).toHaveBeenCalledWith('test-video', expect.any(AbortSignal))
    expect(mockPostPull).not.toHaveBeenCalled()
    expect(result.current.state.infoMsg).toBe('予約しました（配信開始を待機中）')
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

  describe('refresh / refreshWithClear 挙動', () => {
    test('refresh() 通常: ACTIVE + server 新 users → 置換 + sortUsersStable 適用 (順序検証)', async () => {
      const userA = {
        channelId: 'UC_A',
        displayName: 'UserA',
        joinedAt: '2024-01-01T09:00:00.000Z',
        firstCommentedAt: '2024-01-01T10:00:00.000Z',
      }
      const userB = {
        channelId: 'UC_B',
        displayName: 'UserB',
        joinedAt: '2024-01-01T08:00:00.000Z',
        firstCommentedAt: '2024-01-01T09:00:00.000Z',
      }
      // server は A, B の順で返すが firstCommentedAt は B が先
      mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
      mockGetUsers.mockResolvedValue([userA, userB])

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      // sortUsersStable により B (firstCommentedAt が早い) が先頭になる
      expect(result.current.state.users).toHaveLength(2)
      expect(result.current.state.users[0].channelId).toBe('UC_B')
      expect(result.current.state.users[1].channelId).toBe('UC_A')
      expect(result.current.state.active).toBe(true)
    })

    test('refresh() 通常: WAITING + server 空 [] + 既存 users あり → 既存保持 (bug 9c0e236 regression guard)', async () => {
      const existingUser = {
        channelId: 'UC_EXIST',
        displayName: 'ExistingUser',
        joinedAt: '2024-01-01T09:00:00.000Z',
        firstCommentedAt: '2024-01-01T10:00:00.000Z',
      }
      // 1回目: users あり
      mockGetStatus.mockResolvedValueOnce({ status: 'ACTIVE' })
      mockGetUsers.mockResolvedValueOnce([existingUser])
      // 2回目: WAITING + 空
      mockGetStatus.mockResolvedValueOnce({ status: 'WAITING' })
      mockGetUsers.mockResolvedValueOnce([])

      const { result } = renderHook(() => useAppState())

      // 1回目 refresh: users をセット
      await act(async () => {
        await result.current.actions.refresh()
      })
      expect(result.current.state.users).toHaveLength(1)

      // 2回目 refresh: WAITING + server 空 → 既存保持
      await act(async () => {
        await result.current.actions.refresh()
      })
      expect(result.current.state.users).toHaveLength(1)
      expect(result.current.state.users[0].channelId).toBe('UC_EXIST')
    })

    test('refresh() 通常: WAITING + server 空 + 既存も空 → 空のまま', async () => {
      mockGetStatus.mockResolvedValue({ status: 'WAITING' })
      mockGetUsers.mockResolvedValue([])

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.users).toHaveLength(0)
    })

    test('refresh() 通常: server 新 users → status 問わず置換', async () => {
      const newUser = {
        channelId: 'UC_NEW',
        displayName: 'NewUser',
        joinedAt: '2024-01-01T09:00:00.000Z',
        firstCommentedAt: '2024-01-01T10:00:00.000Z',
      }
      mockGetStatus.mockResolvedValue({ status: 'WAITING' })
      mockGetUsers.mockResolvedValue([newUser])

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.users).toHaveLength(1)
      expect(result.current.state.users[0].channelId).toBe('UC_NEW')
    })

    test('refreshWithClear (onReset 経由): server 空でも強制 clear', async () => {
      const existingUser = {
        channelId: 'UC_EXIST',
        displayName: 'ExistingUser',
        joinedAt: '2024-01-01T09:00:00.000Z',
        firstCommentedAt: '2024-01-01T10:00:00.000Z',
      }
      // 1回目: users をセット
      mockGetStatus.mockResolvedValueOnce({ status: 'ACTIVE' })
      mockGetUsers.mockResolvedValueOnce([existingUser])
      // onReset: postReset
      mockPostReset.mockResolvedValue(undefined)
      // onReset 後の refresh (clear あり): 空
      mockGetStatus.mockResolvedValueOnce({ status: 'WAITING' })
      mockGetUsers.mockResolvedValueOnce([])

      const { result } = renderHook(() => useAppState())

      // 1回目 refresh: users をセット
      await act(async () => {
        await result.current.actions.refresh()
      })
      expect(result.current.state.users).toHaveLength(1)

      // onReset 経由で refreshWithClear が発火 → server 空でも強制 clear
      await act(async () => {
        await result.current.actions.onReset()
      })
      expect(result.current.state.users).toHaveLength(0)
    })

    test('refreshWithClear (onReset 経由): server 新 users → 置換', async () => {
      const newUser = {
        channelId: 'UC_NEW',
        displayName: 'NewUser',
        joinedAt: '2024-01-01T09:00:00.000Z',
        firstCommentedAt: '2024-01-01T10:00:00.000Z',
      }
      mockPostReset.mockResolvedValue(undefined)
      mockGetStatus.mockResolvedValue({ status: 'ACTIVE' })
      mockGetUsers.mockResolvedValue([newUser])

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.onReset()
      })

      expect(result.current.state.users).toHaveLength(1)
      expect(result.current.state.users[0].channelId).toBe('UC_NEW')
    })

    test('abort 中の AbortError は errorMsg をセットしない (refresh)', async () => {
      const abortError = Object.assign(new Error('The operation was aborted'), {
        name: 'AbortError',
      })
      mockGetStatus.mockRejectedValue(abortError)

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.errorMsg).toBe('')
    })

    test('AbortError 以外の error は エラーメッセージをセットする (refresh)', async () => {
      mockGetStatus.mockRejectedValue(new Error('network error'))

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.errorMsg).toBe(
        '更新に失敗しました。しばらくしてから再試行してください。',
      )
    })

    test('snapshotRestoreMsg が refresh でも onReset 経由でも consumeSnapshotRestoreMsg 経由でセットされる', async () => {
      vi.useFakeTimers()
      vi.setSystemTime(new Date(2024, 5, 9, 10, 0, 0))

      Object.defineProperty(window, 'sessionStorage', {
        value: {
          getItem: vi.fn(() => null),
          setItem: vi.fn(),
          removeItem: vi.fn(),
          clear: vi.fn(),
        },
        writable: true,
      })

      const savedAt = new Date(2024, 5, 9, 14, 23, 0)
      mockGetStatus.mockResolvedValue({
        status: 'ACTIVE',
        snapshotSavedAt: savedAt.toISOString(),
      })
      mockGetUsers.mockResolvedValue([])

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.snapshotRestoreMsg).toContain('14:23')
      expect(result.current.state.snapshotRestoreMsg).toContain('の保存状態を取得しました')

      vi.useRealTimers()
    })

    test('lastSnapshotAt が refresh でも onReset 経由でも formatSnapshotSavedAt 経由で更新される', async () => {
      vi.useFakeTimers()
      vi.setSystemTime(new Date(2024, 5, 9, 10, 0, 0))

      const savedAt = new Date(2024, 5, 9, 15, 45, 0)
      // refresh 用
      mockGetStatus.mockResolvedValueOnce({
        status: 'ACTIVE',
        snapshotSavedAt: savedAt.toISOString(),
      })
      mockGetUsers.mockResolvedValueOnce([])
      // onReset 用
      mockPostReset.mockResolvedValue(undefined)
      mockGetStatus.mockResolvedValueOnce({
        status: 'WAITING',
        snapshotSavedAt: savedAt.toISOString(),
      })
      mockGetUsers.mockResolvedValueOnce([])

      const { result } = renderHook(() => useAppState())

      // refresh 経由
      await act(async () => {
        await result.current.actions.refresh()
      })
      expect(result.current.state.lastSnapshotAt).toBe('15:45')

      // onReset 経由 (refreshWithClear)
      await act(async () => {
        await result.current.actions.onReset()
      })
      expect(result.current.state.lastSnapshotAt).toBe('15:45')

      vi.useRealTimers()
    })
  })

  describe('BackendError logs 流し込み', () => {
    test('onSwitch が BackendError をスローした場合、error.logs が addEntry に流れる', async () => {
      const backendErr = new BackendError('動画が見つかりません', {
        code: 'video_not_found',
        httpCode: 404,
        logs: [
          { level: 'error', source: 'YOUTUBE', message: 'videoId not found' },
          { level: 'warn', source: 'SWITCH', message: 'liveChatId not resolved' },
        ],
      })
      mockPostSwitchVideo.mockRejectedValue(backendErr)

      const addEntry = vi.fn()
      const { result } = renderHook(() => useAppState(addEntry))

      act(() => {
        result.current.actions.setVideoId('bad-video')
      })

      await act(async () => {
        await result.current.actions.onSwitch()
      })

      // error entry (失敗メッセージ) が呼ばれる
      expect(addEntry).toHaveBeenCalledWith('error', '切替に失敗しました')
      // logs が level ごとに addEntry に流れる
      expect(addEntry).toHaveBeenCalledWith('error', '[YOUTUBE] videoId not found')
      expect(addEntry).toHaveBeenCalledWith('warn', '[SWITCH] liveChatId not resolved')
    })

    test('onReset が BackendError をスローした場合、error.logs が addEntry に流れる', async () => {
      const backendErr = new BackendError('リセット失敗', {
        httpCode: 500,
        logs: [{ level: 'error', source: 'RESET', message: 'snapshot flush failed' }],
      })
      mockPostReset.mockRejectedValue(backendErr)

      const addEntry = vi.fn()
      const { result } = renderHook(() => useAppState(addEntry))

      await act(async () => {
        await result.current.actions.onReset()
      })

      expect(addEntry).toHaveBeenCalledWith('error', 'リセットに失敗しました')
      expect(addEntry).toHaveBeenCalledWith('error', '[RESET] snapshot flush failed')
    })

    test('BackendError に logs がない場合は addEntry の logs 呼び出しが発生しない', async () => {
      const backendErr = new BackendError('内部エラー', { httpCode: 500, logs: [] })
      mockPostReset.mockRejectedValue(backendErr)

      const addEntry = vi.fn()
      const { result } = renderHook(() => useAppState(addEntry))

      await act(async () => {
        await result.current.actions.onReset()
      })

      // error entry (失敗メッセージ) のみ
      expect(addEntry).toHaveBeenCalledTimes(1)
      expect(addEntry).toHaveBeenCalledWith('error', 'リセットに失敗しました')
    })

    test('通常の Error スロー時は logs 流し込みが発生しない', async () => {
      mockPostSwitchVideo.mockRejectedValue(new Error('generic error'))

      const addEntry = vi.fn()
      const { result } = renderHook(() => useAppState(addEntry))

      act(() => {
        result.current.actions.setVideoId('any-video')
      })

      await act(async () => {
        await result.current.actions.onSwitch()
      })

      // error entry (失敗メッセージ) のみ、logs 流し込みなし
      expect(addEntry).toHaveBeenCalledTimes(1)
      expect(addEntry).toHaveBeenCalledWith('error', '切替に失敗しました')
    })
  })

  describe('state.status (Step 3)', () => {
    test('state.status の初期値は WAITING', () => {
      const { result } = renderHook(() => useAppState())
      expect(result.current.state.status).toBe('WAITING')
    })

    test('refresh 後 state.status が server response の status と一致する', async () => {
      mockGetStatus.mockResolvedValueOnce({ status: 'ACTIVE' })
      mockGetUsers.mockResolvedValue([])

      const { result } = renderHook(() => useAppState())

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.status).toBe('ACTIVE')

      mockGetStatus.mockResolvedValueOnce({ status: 'WAITING' })
      mockGetUsers.mockResolvedValue([])

      await act(async () => {
        await result.current.actions.refresh()
      })

      expect(result.current.state.status).toBe('WAITING')
    })
  })
})
