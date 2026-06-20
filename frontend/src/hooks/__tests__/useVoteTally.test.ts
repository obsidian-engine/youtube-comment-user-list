import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useVoteTally } from '../useVoteTally'
import type { Comment } from '../../utils/api'

const makeComment = (id: string, channelId: string, message: string): Comment => ({
  id,
  channelId,
  displayName: `User-${channelId}`,
  message,
  publishedAt: `2024-01-01T00:00:0${id}Z`,
})

const baseComments: Comment[] = [
  makeComment('1', 'ch1', 'A'),
  makeComment('2', 'ch2', 'B'),
  makeComment('3', 'ch3', 'A'),
]

// テスト間 localStorage 汚染を防ぐため全テストで afterEach クリア
afterEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
})

describe('useVoteTally(snapshot) - 初期状態', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('初期値が正しい', () => {
    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))
    expect(result.current.keywordsInput).toBe('')
    expect(result.current.parsedKeywords).toEqual([])
    expect(result.current.counts).toEqual({})
    expect(result.current.voters).toEqual({})
    expect(result.current.totalVotes).toBe(0)
  })

  it('matchMode は localStorage から復元される', () => {
    const store: Record<string, string> = { pollMatchMode: 'partial' }
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))
    expect(result.current.matchMode).toBe('partial')
  })

  it('localStorage が未設定なら matchMode は exact', () => {
    const store: Record<string, string> = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))
    expect(result.current.matchMode).toBe('exact')
  })
})

describe('useVoteTally(snapshot) - キーワード入力 / 集計', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('setKeywordsInput で parsedKeywords と counts が更新される', () => {
    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))

    act(() => result.current.setKeywordsInput('A\nB'))

    expect(result.current.parsedKeywords).toEqual(['A', 'B'])
    expect(result.current.counts['A']).toBe(2) // ch1, ch3 が 'A'
    expect(result.current.counts['B']).toBe(1) // ch2 が 'B'
    expect(result.current.totalVotes).toBe(3)
  })

  it('API fetch は発生しない (snapshot モード)', () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch')

    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))
    act(() => result.current.setKeywordsInput('A'))

    expect(fetchSpy).not.toHaveBeenCalled()
  })

  it('重複キーワードは正規化される', () => {
    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))

    act(() => result.current.setKeywordsInput('A\nA'))

    expect(result.current.parsedKeywords).toEqual(['A'])
  })
})

describe('useVoteTally(snapshot) - matchMode 切替', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('partial に切替えると部分一致で集計が変わる', () => {
    const partialComments: Comment[] = [
      makeComment('1', 'ch1', '賛成'),
      makeComment('2', 'ch2', '賛成です'),
    ]
    const store: Record<string, string> = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() =>
      useVoteTally({ mode: 'snapshot', comments: partialComments }),
    )

    act(() => result.current.setKeywordsInput('賛成'))

    // exact では '賛成' のみ 1 票
    expect(result.current.counts['賛成']).toBe(1)

    act(() => result.current.setMatchMode('partial'))

    // partial では '賛成' / '賛成です' 両方でマッチ → 2 票
    expect(result.current.counts['賛成']).toBe(2)
  })

  it('matchMode 切替で localStorage に保存される', () => {
    const store: Record<string, string> = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() => useVoteTally({ mode: 'snapshot', comments: baseComments }))

    act(() => result.current.setMatchMode('partial'))

    expect(store['pollMatchMode']).toBe('partial')
  })
})

describe('useVoteTally(snapshot) - comments 再レンダー', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('comments が変わると再集計される', () => {
    const { result, rerender } = renderHook(
      ({ comments }: { comments: Comment[] }) => useVoteTally({ mode: 'snapshot', comments }),
      { initialProps: { comments: baseComments } },
    )

    act(() => result.current.setKeywordsInput('A'))
    expect(result.current.counts['A']).toBe(2)

    const newComments: Comment[] = [makeComment('4', 'ch4', 'A')]
    rerender({ comments: newComments })

    expect(result.current.counts['A']).toBe(1)
  })
})
