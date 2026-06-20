import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { usePollCount } from '../usePollCount'
import * as api from '../../utils/api'
import type { Comment } from '../../utils/api'

vi.mock('../../utils/api', async () => {
  const actual = await vi.importActual<typeof api>('../../utils/api')
  return {
    ...actual,
    searchComments: vi.fn(),
  }
})

const mockedSearch = vi.mocked(api.searchComments)

const c = (channelId: string, message: string, publishedAt: string): Comment => ({
  id: `${channelId}-${publishedAt}`,
  channelId,
  displayName: channelId,
  message,
  publishedAt,
})

describe('usePollCount - 初期状態', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('空キーワード / counts={} / lastUpdated=--:--:--', () => {
    const { result } = renderHook(() => usePollCount())
    expect(result.current.keywords).toEqual([])
    expect(result.current.counts).toEqual({})
    expect(result.current.totalVotes).toBe(0)
    expect(result.current.lastUpdated).toBe('--:--:--')
    expect(result.current.errorMsg).toBe('')
    expect(result.current.isLoading).toBe(false)
  })

  it('公開関数が揃っている', () => {
    const { result } = renderHook(() => usePollCount())
    expect(typeof result.current.addKeyword).toBe('function')
    expect(typeof result.current.removeKeyword).toBe('function')
    expect(typeof result.current.clearKeywords).toBe('function')
    expect(typeof result.current.recount).toBe('function')
  })
})

describe('usePollCount - addKeyword / removeKeyword', () => {
  beforeEach(() => vi.clearAllMocks())

  it('追加するとキーワードが増え counts も 0 初期化', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    expect(result.current.keywords).toEqual(['hoge'])
    expect(result.current.counts).toEqual({ hoge: 0 })
  })

  it('複数追加で順序保持', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('a'))
    act(() => result.current.addKeyword('b'))
    act(() => result.current.addKeyword('c'))
    expect(result.current.keywords).toEqual(['a', 'b', 'c'])
  })

  it('重複追加は無視', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => result.current.addKeyword('hoge'))
    expect(result.current.keywords).toEqual(['hoge'])
  })

  it('空文字 / 空白のみ追加は無視', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword(''))
    act(() => result.current.addKeyword('   '))
    expect(result.current.keywords).toEqual([])
  })

  it('前後空白は trim される', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('  hoge  '))
    expect(result.current.keywords).toEqual(['hoge'])
  })

  it('removeKeyword で削除', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('a'))
    act(() => result.current.addKeyword('b'))
    act(() => result.current.removeKeyword('a'))
    expect(result.current.keywords).toEqual(['b'])
    expect(result.current.counts).toEqual({ b: 0 })
  })
})

describe('usePollCount - recount 正常系', () => {
  beforeEach(() => vi.clearAllMocks())

  it('成功で counts/lastUpdated 更新', async () => {
    mockedSearch.mockResolvedValue([
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'fuga', '2024-01-01T00:00:02Z'),
      c('u3', 'hoge', '2024-01-01T00:00:03Z'),
    ])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => result.current.addKeyword('fuga'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.counts).toEqual({ hoge: 2, fuga: 1 })
    expect(result.current.totalVotes).toBe(3)
    expect(result.current.lastUpdated).not.toBe('--:--:--')
    expect(result.current.isLoading).toBe(false)
  })

  it('comments 0 件でも counts 全 0 で正常終了', async () => {
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => result.current.addKeyword('fuga'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.counts).toEqual({ hoge: 0, fuga: 0 })
    expect(result.current.totalVotes).toBe(0)
  })

  it('recount は searchComments に keywords を渡す', async () => {
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('alpha'))
    act(() => result.current.addKeyword('beta'))
    await act(async () => {
      await result.current.recount()
    })

    expect(mockedSearch).toHaveBeenCalledWith(['alpha', 'beta'], expect.any(AbortSignal))
  })

  it('searchComments が null を返した場合も全 0 で正常終了', async () => {
    mockedSearch.mockResolvedValue(null)

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => result.current.addKeyword('fuga'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.counts).toEqual({ hoge: 0, fuga: 0 })
    expect(result.current.totalVotes).toBe(0)
  })
})

describe('usePollCount - recount エラー系', () => {
  beforeEach(() => vi.clearAllMocks())

  it('keywords 空のとき NO_KEYWORDS エラー、API は呼ばれない', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.errorMsg).toContain('キーワード')
    expect(mockedSearch).not.toHaveBeenCalled()
  })

  it('HttpError(404) で SERVER_UNREACHABLE', async () => {
    mockedSearch.mockRejectedValue(new api.HttpError(404))

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('接続できません')
    expect(result.current.isLoading).toBe(false)
  })

  it('HttpError(500) で SERVER_ERROR', async () => {
    mockedSearch.mockRejectedValue(new api.HttpError(500))

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('サーバーエラー')
  })

  it('TypeError(Failed to fetch) で NETWORK 表示', async () => {
    mockedSearch.mockRejectedValue(new TypeError('Failed to fetch'))

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('ネットワーク')
  })

  it('TimeoutError で TIMEOUT メッセージ表示', async () => {
    const e = new Error('timeout')
    e.name = 'TimeoutError'
    mockedSearch.mockRejectedValue(e)

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('タイムアウト')
    expect(result.current.isLoading).toBe(false)
  })

  it('AbortError は state を変更しない（errorMsg 空のまま）', async () => {
    mockedSearch.mockImplementation(() => {
      const e = new Error('aborted')
      e.name = 'AbortError'
      return Promise.reject(e)
    })

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toBe('')
  })

  it('エラー後のリトライで成功すると errorMsg がクリア', async () => {
    mockedSearch.mockRejectedValueOnce(new api.HttpError(500))
    mockedSearch.mockResolvedValueOnce([c('u1', 'hoge', '2024-01-01T00:00:01Z')])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.errorMsg).not.toBe('')

    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.errorMsg).toBe('')
    expect(result.current.counts).toEqual({ hoge: 1 })
  })
})

describe('usePollCount - clearKeywords / race 防止', () => {
  beforeEach(() => vi.clearAllMocks())

  it('clearKeywords: 状態を初期化', async () => {
    mockedSearch.mockResolvedValue([c('u1', 'hoge', '2024-01-01T00:00:01Z')])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.keywords).toEqual(['hoge'])
    expect(result.current.counts).toEqual({ hoge: 1 })

    act(() => {
      result.current.clearKeywords()
    })
    expect(result.current.keywords).toEqual([])
    expect(result.current.counts).toEqual({})
    expect(result.current.lastUpdated).toBe('--:--:--')
  })

  it('clearKeywords: 進行中リクエストを abort', async () => {
    let capturedSignal: AbortSignal | undefined
    mockedSearch.mockImplementation((_keywords, signal) => {
      capturedSignal = signal
      return new Promise(() => {})
    })

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => {
      void result.current.recount()
    })
    await waitFor(() => expect(capturedSignal).toBeDefined())

    act(() => {
      result.current.clearKeywords()
    })
    expect(capturedSignal?.aborted).toBe(true)
  })

  it('recount を 2 回連投すると先発が abort される', async () => {
    const signals: AbortSignal[] = []
    mockedSearch.mockImplementation((_keywords, signal) => {
      signals.push(signal!)
      return new Promise((resolve) => {
        setTimeout(() => resolve([]), 30)
      })
    })

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      void result.current.recount()
      await result.current.recount()
    })

    expect(signals[0].aborted).toBe(true)
    expect(signals[1].aborted).toBe(false)
  })
})

describe('usePollCount - localStorage 永続化', () => {
  let store: Record<string, string>

  beforeEach(() => {
    vi.clearAllMocks()
    store = {}
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })
    vi.spyOn(window.localStorage, 'clear').mockImplementation(() => {
      store = {}
    })
  })

  it('キーワード追加で localStorage に保存される', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => result.current.addKeyword('fuga'))
    expect(JSON.parse(localStorage.getItem('pollKeywords') ?? '[]')).toEqual(['hoge', 'fuga'])
  })

  it('mount 時に localStorage から復元される', () => {
    localStorage.setItem('pollKeywords', JSON.stringify(['a', 'b', 'c']))
    const { result } = renderHook(() => usePollCount())
    expect(result.current.keywords).toEqual(['a', 'b', 'c'])
    expect(result.current.counts).toEqual({ a: 0, b: 0, c: 0 })
    expect(result.current.voters).toEqual({ a: [], b: [], c: [] })
  })

  it('removeKeyword で localStorage も更新される', () => {
    localStorage.setItem('pollKeywords', JSON.stringify(['a', 'b']))
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.removeKeyword('a'))
    expect(JSON.parse(localStorage.getItem('pollKeywords') ?? '[]')).toEqual(['b'])
  })

  it('clearKeywords で localStorage も空になる', () => {
    localStorage.setItem('pollKeywords', JSON.stringify(['a', 'b']))
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.clearKeywords())
    expect(JSON.parse(localStorage.getItem('pollKeywords') ?? '[]')).toEqual([])
  })

  it('localStorage が壊れていても crash せず空で起動', () => {
    localStorage.setItem('pollKeywords', '{not json')
    const { result } = renderHook(() => usePollCount())
    expect(result.current.keywords).toEqual([])
  })

  it('localStorage が配列でない場合は空で起動', () => {
    localStorage.setItem('pollKeywords', JSON.stringify({ keywords: ['a'] }))
    const { result } = renderHook(() => usePollCount())
    expect(result.current.keywords).toEqual([])
  })
})

describe('usePollCount - clearResults', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('clearResults 呼出後 counts が空オブジェクトになる', async () => {
    mockedSearch.mockResolvedValue([
      c('u1', 'hoge', '2024-01-01T00:00:01Z'),
      c('u2', 'hoge', '2024-01-01T00:00:02Z'),
    ])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.counts).toEqual({ hoge: 2 })

    act(() => result.current.clearResults())
    expect(result.current.counts).toEqual({})
  })

  it('clearResults 呼出後 voters が空オブジェクトになる', async () => {
    mockedSearch.mockResolvedValue([c('u1', 'hoge', '2024-01-01T00:00:01Z')])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.voters).toEqual({ hoge: expect.any(Array) })

    act(() => result.current.clearResults())
    expect(result.current.voters).toEqual({})
  })

  it('clearResults 呼出後 errorMsg が空文字になる', async () => {
    mockedSearch.mockRejectedValue(new api.HttpError(500))

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.errorMsg).not.toBe('')

    act(() => result.current.clearResults())
    expect(result.current.errorMsg).toBe('')
  })

  it('clearResults 呼出後 lastUpdated が --:--:-- になる', async () => {
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.lastUpdated).not.toBe('--:--:--')

    act(() => result.current.clearResults())
    expect(result.current.lastUpdated).toBe('--:--:--')
  })

  it('clearResults 呼出後 keywords は保持される', async () => {
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    act(() => result.current.addKeyword('fuga'))
    await act(async () => {
      await result.current.recount()
    })

    act(() => result.current.clearResults())
    expect(result.current.keywords).toEqual(['hoge', 'fuga'])
  })

  it('clearResults が公開されている', () => {
    const { result } = renderHook(() => usePollCount())
    expect(typeof result.current.clearResults).toBe('function')
  })
})

describe('usePollCount - matchMode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('setMatchMode で matchMode が更新される', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.setMatchMode('partial'))
    expect(result.current.matchMode).toBe('partial')
  })

  it('同一モードを再選択しても recount は呼ばれない', async () => {
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('hoge'))
    await act(async () => {
      await result.current.recount()
    })
    mockedSearch.mockClear()

    act(() => result.current.setMatchMode('exact'))
    expect(mockedSearch).not.toHaveBeenCalled()
  })

  it('キーワード未設定時は matchMode のみ更新し API は呼ばない', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.setMatchMode('partial'))
    expect(result.current.matchMode).toBe('partial')
    expect(mockedSearch).not.toHaveBeenCalled()
  })

  it('matchMode 切替時に自動 recount し、新モードで集計する', async () => {
    mockedSearch.mockResolvedValue([c('u1', '賛成です', '2024-01-01T00:00:01Z')])

    const { result } = renderHook(() => usePollCount())
    act(() => result.current.addKeyword('賛成'))
    await act(async () => {
      await result.current.recount()
    })
    expect(result.current.counts).toEqual({ 賛成: 0 })

    await act(async () => {
      result.current.setMatchMode('partial')
    })

    expect(mockedSearch).toHaveBeenCalledTimes(2)
    expect(result.current.matchMode).toBe('partial')
    expect(result.current.counts).toEqual({ 賛成: 1 })
  })

  it('matchMode を localStorage に保存する', () => {
    const { result } = renderHook(() => usePollCount())
    act(() => result.current.setMatchMode('partial'))
    expect(localStorage.getItem('pollMatchMode')).toBe('partial')
  })
})
