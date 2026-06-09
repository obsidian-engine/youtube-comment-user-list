import { renderHook, act } from '@testing-library/react'
import { useHistory } from '../useHistory'
import { vi, describe, test, expect, beforeEach, afterEach } from 'vitest'
import { BackendError } from '../../utils/api'
import type { HistorySummary, HistorySnapshot } from '../../utils/api'

const mockGetHistorySnapshots = vi.fn()
const mockGetHistorySnapshot = vi.fn()

vi.mock('../../utils/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../utils/api')>()
  return {
    ...actual,
    getHistorySnapshots: (...args: unknown[]) => mockGetHistorySnapshots(...args),
    getHistorySnapshot: (...args: unknown[]) => mockGetHistorySnapshot(...args),
  }
})

const mockSummaries: HistorySummary[] = [
  { videoId: 'vid-1', savedAt: '2024-06-01T10:00:00Z', userCount: 10, commentCount: 50 },
  { videoId: 'vid-2', savedAt: '2024-06-02T11:00:00Z', userCount: 5, commentCount: 20 },
]

const mockSnapshot: HistorySnapshot = {
  videoId: 'vid-1',
  savedAt: '2024-06-01T10:00:00Z',
  users: [{ channelId: 'UC1', displayName: 'User1', joinedAt: '2024-06-01T09:00:00Z' }],
  comments: [],
}

describe('useHistory', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  test('初期状態: snapshots=[], selected=null, loading=false, error=""', () => {
    const { result } = renderHook(() => useHistory())

    expect(result.current.snapshots).toEqual([])
    expect(result.current.selected).toBeNull()
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe('')
  })

  test('loadList() 成功: loading=true → snapshots に値, loading=false, error=""', async () => {
    mockGetHistorySnapshots.mockResolvedValue(mockSummaries)

    const { result } = renderHook(() => useHistory())

    await act(async () => {
      await result.current.loadList()
    })

    expect(mockGetHistorySnapshots).toHaveBeenCalled()
    expect(result.current.snapshots).toEqual(mockSummaries)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe('')
  })

  test('loadList() 失敗: error メッセージ set, loading=false', async () => {
    mockGetHistorySnapshots.mockRejectedValue(new Error('network error'))

    const { result } = renderHook(() => useHistory())

    await act(async () => {
      await result.current.loadList()
    })

    expect(result.current.error).toBe('履歴の取得に失敗しました')
    expect(result.current.loading).toBe(false)
    expect(result.current.snapshots).toEqual([])
  })

  test('select(videoId) 成功: loading=true → detail fetch → selected に snapshot', async () => {
    mockGetHistorySnapshot.mockResolvedValue(mockSnapshot)

    const { result } = renderHook(() => useHistory())

    await act(async () => {
      await result.current.select('vid-1')
    })

    expect(mockGetHistorySnapshot).toHaveBeenCalledWith('vid-1', expect.any(AbortSignal))
    expect(result.current.selected).toEqual(mockSnapshot)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe('')
  })

  test('select(unknown) 404 (BackendError): error set, selected=null, loading=false', async () => {
    const backendErr = new BackendError('not found', { httpCode: 404, logs: [] })
    mockGetHistorySnapshot.mockRejectedValue(backendErr)

    const { result } = renderHook(() => useHistory())

    await act(async () => {
      await result.current.select('unknown-vid')
    })

    expect(result.current.error).toBe('スナップショットの取得に失敗しました')
    expect(result.current.selected).toBeNull()
    expect(result.current.loading).toBe(false)
  })

  test('clearSelected(): selected=null になる', async () => {
    mockGetHistorySnapshot.mockResolvedValue(mockSnapshot)

    const { result } = renderHook(() => useHistory())

    await act(async () => {
      await result.current.select('vid-1')
    })
    expect(result.current.selected).toEqual(mockSnapshot)

    act(() => {
      result.current.clearSelected()
    })

    expect(result.current.selected).toBeNull()
  })
})
