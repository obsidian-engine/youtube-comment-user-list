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

const txtFile = (content: string): File =>
  new File([content], 'keywords.txt', { type: 'text/plain' })

const c = (channelId: string, message: string, publishedAt: string): Comment => ({
  id: `${channelId}-${publishedAt}`,
  channelId,
  displayName: channelId,
  message,
  publishedAt,
})

describe('usePollCount - 初期状態', () => {
  beforeEach(() => vi.clearAllMocks())

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
    expect(typeof result.current.loadKeywordsFromFile).toBe('function')
    expect(typeof result.current.clearKeywords).toBe('function')
    expect(typeof result.current.recount).toBe('function')
  })
})

describe('usePollCount - loadKeywordsFromFile', () => {
  beforeEach(() => vi.clearAllMocks())

  it('有効な txt 読込で keywords セット + counts 初期化', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge\nfuga\n'))
    })
    expect(result.current.keywords).toEqual(['hoge', 'fuga'])
    expect(result.current.counts).toEqual({ hoge: 0, fuga: 0 })
    expect(result.current.errorMsg).toBe('')
  })

  it('空 txt はエラー表示、keywords は更新されない', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('\n\n  \n'))
    })
    expect(result.current.keywords).toEqual([])
    expect(result.current.errorMsg).toContain('含まれていません')
  })

  it(', 含む行は除外し警告表示（残りキーワードはロード）', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge\n1,000円\nfuga'))
    })
    expect(result.current.keywords).toEqual(['hoge', 'fuga'])
    expect(result.current.errorMsg).toContain('1,000円')
  })

  it('全行がカンマ含むなら EMPTY_FILE エラー', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('a,b\nc,d'))
    })
    expect(result.current.keywords).toEqual([])
    expect(result.current.errorMsg).toContain('含まれていません')
  })

  it('再ロードで keywords を上書き', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('a\nb'))
    })
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('x\ny\nz'))
    })
    expect(result.current.keywords).toEqual(['x', 'y', 'z'])
    expect(result.current.counts).toEqual({ x: 0, y: 0, z: 0 })
  })

  it('再ロードでエラーメッセージはクリアされる（正常時）', async () => {
    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile(''))
    })
    expect(result.current.errorMsg).not.toBe('')
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    expect(result.current.errorMsg).toBe('')
  })

  it('file.text() が throw した場合 READ_FAILED エラー', async () => {
    const brokenFile = {
      text: () => Promise.reject(new Error('read error')),
    } as unknown as File

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(brokenFile)
    })
    expect(result.current.errorMsg).toContain('読み込みに失敗')
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge\nfuga'))
    })
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge\nfuga'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.counts).toEqual({ hoge: 0, fuga: 0 })
    expect(result.current.totalVotes).toBe(0)
  })

  it('totalVotes は counts 合計に追従', async () => {
    mockedSearch.mockResolvedValue([
      c('u1', 'a', '2024-01-01T00:00:01Z'),
      c('u2', 'a', '2024-01-01T00:00:02Z'),
      c('u3', 'b', '2024-01-01T00:00:03Z'),
    ])

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('a\nb\nc'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.totalVotes).toBe(3)
  })

  it('recount は searchComments に keywords を渡す', async () => {
    mockedSearch.mockResolvedValue([])

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('alpha\nbeta'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(mockedSearch).toHaveBeenCalledWith(['alpha', 'beta'], expect.any(AbortSignal))
  })

  it('recount 連続呼び出し: 最終的に1回分の結果が反映', async () => {
    mockedSearch.mockResolvedValue([c('u1', 'hoge', '2024-01-01T00:00:01Z')])

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
      await result.current.recount()
    })

    expect(result.current.counts).toEqual({ hoge: 1 })
  })

  it('searchComments が null を返した場合も全 0 で正常終了', async () => {
    mockedSearch.mockResolvedValue(null)

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge\nfuga'))
    })
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('接続できません')
    expect(result.current.isLoading).toBe(false)
  })

  it('HttpError(500) で SERVER_ERROR', async () => {
    mockedSearch.mockRejectedValue(new api.HttpError(500))

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('サーバーエラー')
  })

  it('HttpError(503) も SERVER_ERROR（>=500）', async () => {
    mockedSearch.mockRejectedValue(new api.HttpError(503))

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('サーバーエラー')
  })

  it('HttpError(400) は GENERIC（<500 かつ ≠404）', async () => {
    mockedSearch.mockRejectedValue(new api.HttpError(400))

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('失敗しました')
  })

  it('TypeError(Failed to fetch) で NETWORK 表示', async () => {
    mockedSearch.mockRejectedValue(new TypeError('Failed to fetch'))

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toBe('')
  })

  it('不明エラー（generic Error）は GENERIC', async () => {
    mockedSearch.mockRejectedValue(new Error('unknown'))

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      await result.current.recount()
    })

    expect(result.current.errorMsg).toContain('失敗しました')
  })

  it('エラー後のリトライで成功すると errorMsg がクリア', async () => {
    mockedSearch.mockRejectedValueOnce(new api.HttpError(500))
    mockedSearch.mockResolvedValueOnce([c('u1', 'hoge', '2024-01-01T00:00:01Z')])

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    act(() => {
      result.current.recount()
    })
    await waitFor(() => expect(capturedSignal).toBeDefined())

    act(() => {
      result.current.clearKeywords()
    })
    expect(capturedSignal?.aborted).toBe(true)
  })

  it('loadKeywordsFromFile が進行中 recount を abort', async () => {
    let capturedSignal: AbortSignal | undefined
    mockedSearch.mockImplementation((_keywords, signal) => {
      capturedSignal = signal
      return new Promise(() => {})
    })

    const { result } = renderHook(() => usePollCount())
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('a'))
    })
    act(() => {
      result.current.recount()
    })
    await waitFor(() => expect(capturedSignal).toBeDefined())

    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('b\nc'))
    })
    expect(capturedSignal?.aborted).toBe(true)
    expect(result.current.keywords).toEqual(['b', 'c'])
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
    await act(async () => {
      await result.current.loadKeywordsFromFile(txtFile('hoge'))
    })
    await act(async () => {
      void result.current.recount()
      await result.current.recount()
    })

    expect(signals[0].aborted).toBe(true)
    expect(signals[1].aborted).toBe(false)
  })
})
