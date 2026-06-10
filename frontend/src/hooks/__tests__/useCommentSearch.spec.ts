import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useCommentSearch } from '../useCommentSearch'
import * as api from '../../utils/api'

vi.mock('../../utils/api', async () => {
  const actual = await vi.importActual<typeof api>('../../utils/api')
  return {
    ...actual,
    searchComments: vi.fn(),
  }
})

describe('useCommentSearch - clearComments', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('clearComments 呼出後 comments が空配列になる', async () => {
    const mockedSearch = vi.mocked(api.searchComments)
    mockedSearch.mockResolvedValue([
      {
        id: 'c1',
        channelId: 'u1',
        displayName: 'User1',
        message: 'hello',
        publishedAt: '2024-01-01T00:00:01Z',
      },
    ])

    const { result } = renderHook(() => useCommentSearch())

    // キーワードを追加して検索を実行
    act(() => result.current.addKeyword('hello'))
    await act(async () => {
      await result.current.search()
    })
    expect(result.current.comments).toHaveLength(1)

    // clearComments で空になる
    act(() => result.current.clearComments())
    expect(result.current.comments).toEqual([])
  })

  it('clearComments 呼出後 errorMsg が null になる', async () => {
    const mockedSearch = vi.mocked(api.searchComments)
    mockedSearch.mockRejectedValue(new TypeError('Failed to fetch'))

    const { result } = renderHook(() => useCommentSearch())
    act(() => result.current.addKeyword('hello'))
    await act(async () => {
      await result.current.search()
    })
    expect(result.current.errorMsg).not.toBe('')

    act(() => result.current.clearComments())
    expect(result.current.errorMsg).toBeNull()
  })

  it('clearComments 呼出後 lastUpdated が null になる', async () => {
    const mockedSearch = vi.mocked(api.searchComments)
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => useCommentSearch())
    act(() => result.current.addKeyword('hello'))
    await act(async () => {
      await result.current.search()
    })
    expect(result.current.lastUpdated).not.toBe('--:--:--')

    act(() => result.current.clearComments())
    expect(result.current.lastUpdated).toBeNull()
  })

  it('clearComments 呼出後 keywords と intervalSec は保持される', () => {
    const { result } = renderHook(() => useCommentSearch())
    const initialKeywords = [...result.current.keywords]
    act(() => result.current.addKeyword('uniqueword123'))
    act(() => result.current.setIntervalSec(30))

    const keywordsBeforeClear = [...result.current.keywords]
    act(() => result.current.clearComments())

    expect(result.current.keywords).toEqual(keywordsBeforeClear)
    expect(result.current.keywords).toContain('uniqueword123')
    expect(result.current.intervalSec).toBe(30)
    // keywords の件数はデフォルト + 追加分
    expect(result.current.keywords).toHaveLength(initialKeywords.length + 1)
  })

  it('clearComments が公開されている', () => {
    const { result } = renderHook(() => useCommentSearch())
    expect(typeof result.current.clearComments).toBe('function')
  })
})
