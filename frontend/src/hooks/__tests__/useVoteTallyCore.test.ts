import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useVoteTallyCore } from '../useVoteTallyCore'
import type { Comment } from '../../utils/api'

const c = (channelId: string, message: string): Comment => ({
  id: channelId,
  channelId,
  displayName: channelId,
  message,
  publishedAt: `2024-01-01T00:00:00Z`,
})

const comments: Comment[] = [c('u1', 'A'), c('u2', 'B'), c('u3', 'A')]

afterEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
})

describe('useVoteTallyCore - matchMode 初期値', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('localStorage 未設定なら exact', () => {
    const store: Record<string, string> = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })
    const { result } = renderHook(() => useVoteTallyCore({ keywords: [], comments }))
    expect(result.current.matchMode).toBe('exact')
  })

  it('localStorage に partial が保存されていれば partial で復元', () => {
    const store: Record<string, string> = { pollMatchMode: 'partial' }
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })
    const { result } = renderHook(() => useVoteTallyCore({ keywords: [], comments }))
    expect(result.current.matchMode).toBe('partial')
  })
})

describe('useVoteTallyCore - counts/voters/totalVotes 算出', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('keywords 空なら counts={} voters={} totalVotes=0', () => {
    const { result } = renderHook(() => useVoteTallyCore({ keywords: [], comments }))
    expect(result.current.counts).toEqual({})
    expect(result.current.voters).toEqual({})
    expect(result.current.totalVotes).toBe(0)
  })

  it('keywords に A を渡すと exact で 2 票カウント', () => {
    const { result } = renderHook(() => useVoteTallyCore({ keywords: ['A'], comments }))
    expect(result.current.counts['A']).toBe(2)
    expect(result.current.totalVotes).toBe(2)
  })

  it('keywords に A,B を渡すと合計 3 票', () => {
    const { result } = renderHook(() => useVoteTallyCore({ keywords: ['A', 'B'], comments }))
    expect(result.current.counts['A']).toBe(2)
    expect(result.current.counts['B']).toBe(1)
    expect(result.current.totalVotes).toBe(3)
  })

  it('matchMode=partial で部分一致集計', () => {
    const partialComments: Comment[] = [c('u1', '賛成'), c('u2', '賛成です')]
    const store: Record<string, string> = { pollMatchMode: 'partial' }
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() =>
      useVoteTallyCore({ keywords: ['賛成'], comments: partialComments }),
    )
    expect(result.current.counts['賛成']).toBe(2)
  })
})

describe('useVoteTallyCore - setMatchMode', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('setMatchMode で matchMode が更新される', () => {
    const { result } = renderHook(() => useVoteTallyCore({ keywords: ['A'], comments }))
    act(() => result.current.setMatchMode('partial'))
    expect(result.current.matchMode).toBe('partial')
  })

  it('setMatchMode で localStorage に保存される', () => {
    const store: Record<string, string> = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() => useVoteTallyCore({ keywords: ['A'], comments }))
    act(() => result.current.setMatchMode('partial'))
    expect(store['pollMatchMode']).toBe('partial')
  })

  it('setMatchMode で counts が再算出される', () => {
    const partialComments: Comment[] = [c('u1', '賛成'), c('u2', '賛成です')]
    const store: Record<string, string> = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    const { result } = renderHook(() =>
      useVoteTallyCore({ keywords: ['賛成'], comments: partialComments }),
    )
    // exact: '賛成' のみ 1 票
    expect(result.current.counts['賛成']).toBe(1)

    act(() => result.current.setMatchMode('partial'))
    // partial: '賛成' '賛成です' → 2 票
    expect(result.current.counts['賛成']).toBe(2)
  })
})

describe('useVoteTallyCore - comments/keywords 変化で再算出', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('keywords が変わると再集計される', () => {
    const { result, rerender } = renderHook(
      ({ keywords }: { keywords: string[] }) => useVoteTallyCore({ keywords, comments }),
      { initialProps: { keywords: ['A'] } },
    )
    expect(result.current.counts['A']).toBe(2)

    rerender({ keywords: ['B'] })
    expect(result.current.counts['B']).toBe(1)
    expect(result.current.counts['A']).toBeUndefined()
  })

  it('comments が変わると再集計される', () => {
    const { result, rerender } = renderHook(
      ({ coms }: { coms: Comment[] }) => useVoteTallyCore({ keywords: ['A'], comments: coms }),
      { initialProps: { coms: comments } },
    )
    expect(result.current.counts['A']).toBe(2)

    rerender({ coms: [c('u4', 'A')] })
    expect(result.current.counts['A']).toBe(1)
  })
})
