import { render, screen, fireEvent } from '@testing-library/react'
import { describe, test, expect, vi, beforeEach, afterEach } from 'vitest'
import { HistoryVotes } from '../HistoryVotes'
import type { Comment } from '../../../utils/api'

beforeEach(() => {
  Object.assign(navigator, {
    clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
  })
})

afterEach(() => {
  vi.restoreAllMocks()
})

const makeComment = (id: string, channelId: string, message: string): Comment => ({
  id,
  channelId,
  displayName: `User-${channelId}`,
  message,
  publishedAt: `2024-01-01T00:00:0${id}Z`,
})

const comments: Comment[] = [
  makeComment('1', 'ch1', 'A'),
  makeComment('2', 'ch2', 'B'),
  makeComment('3', 'ch3', 'A'),
]

describe('HistoryVotes', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  test('キーワード 0 個のとき PollResults は null レンダー (集計テーブル非表示)', () => {
    render(<HistoryVotes comments={comments} />)
    // PollResults は keywords.length === 0 で null を返す
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(screen.getByText('キーワードを入力すると集計します')).toBeInTheDocument()
  })

  test('キーワード 1 つ入力で counts が正しく集計される', () => {
    render(<HistoryVotes comments={comments} />)

    const textarea = screen.getByPlaceholderText('キーワード（改行またはカンマ区切り）')
    fireEvent.change(textarea, { target: { value: 'A' } })

    // テーブルが表示される
    expect(screen.getByRole('table')).toBeInTheDocument()
    // キーワード "A" に 2 票 (ch1, ch3)
    const rows = screen.getAllByRole('row')
    // header + data row + footer = 3 rows
    const dataRow = rows.find(
      (r) => r.textContent?.includes('A') && !r.textContent?.includes('キーワード'),
    )
    expect(dataRow).toBeTruthy()
    expect(dataRow?.textContent).toContain('2')
  })

  test('改行区切りで複数キーワード入力が動作する', () => {
    render(<HistoryVotes comments={comments} />)

    const textarea = screen.getByPlaceholderText('キーワード（改行またはカンマ区切り）')
    fireEvent.change(textarea, { target: { value: 'A\nB' } })

    expect(screen.getByRole('table')).toBeInTheDocument()
    // キーワード A と B の両方が表示される
    const cells = screen.getAllByRole('cell')
    const texts = cells.map((c) => c.textContent)
    expect(texts.some((t) => t?.includes('A'))).toBe(true)
    expect(texts.some((t) => t?.includes('B'))).toBe(true)
  })

  test('カンマ区切りで複数キーワード入力が動作する', () => {
    render(<HistoryVotes comments={comments} />)

    const textarea = screen.getByPlaceholderText('キーワード（改行またはカンマ区切り）')
    fireEvent.change(textarea, { target: { value: 'A,B' } })

    expect(screen.getByRole('table')).toBeInTheDocument()
    const cells = screen.getAllByRole('cell')
    const texts = cells.map((c) => c.textContent)
    expect(texts.some((t) => t?.includes('A'))).toBe(true)
    expect(texts.some((t) => t?.includes('B'))).toBe(true)
  })

  test('重複キーワード入力で 1 個に正規化される', () => {
    render(<HistoryVotes comments={comments} />)

    const textarea = screen.getByPlaceholderText('キーワード（改行またはカンマ区切り）')
    fireEvent.change(textarea, { target: { value: 'A\nA\nA' } })

    expect(screen.getByRole('table')).toBeInTheDocument()
    // "A" のヘッダー行を除いたキーワード行が 1 つだけ
    const rows = screen.getAllByRole('row')
    // header(1) + data(1) + footer(1) = 3 rows (重複が除去されていれば data は 1 行)
    // 展開なしの場合、tbody の行数は keywords.length
    const tbodyRows = rows.filter((r) => {
      const parent = r.closest('tbody')
      return parent !== null
    })
    expect(tbodyRows.length).toBe(1)
  })

  test('部分一致モードでキーワードを含むコメントを集計する', () => {
    const partialComments: Comment[] = [
      makeComment('1', 'ch1', '賛成'),
      makeComment('2', 'ch2', '賛成です'),
    ]
    render(<HistoryVotes comments={partialComments} />)

    fireEvent.click(screen.getByRole('button', { name: '部分一致' }))
    fireEvent.change(screen.getByPlaceholderText('キーワード（改行またはカンマ区切り）'), {
      target: { value: '賛成' },
    })

    const rows = screen.getAllByRole('row')
    const dataRow = rows.find(
      (r) => r.textContent?.includes('賛成') && !r.textContent?.includes('キーワード'),
    )
    expect(dataRow?.textContent).toContain('2')
  })

  test('完全一致モードでは部分マッチのみのコメントをカウントしない', () => {
    const partialComments: Comment[] = [makeComment('1', 'ch1', '賛成です')]
    render(<HistoryVotes comments={partialComments} />)

    fireEvent.change(screen.getByPlaceholderText('キーワード（改行またはカンマ区切り）'), {
      target: { value: '賛成' },
    })

    const rows = screen.getAllByRole('row')
    const dataRow = rows.find(
      (r) => r.textContent?.includes('賛成') && !r.textContent?.includes('キーワード'),
    )
    expect(dataRow?.textContent).toContain('0')

    fireEvent.click(screen.getByRole('button', { name: '部分一致' }))
    expect(dataRow?.textContent).toContain('1')
  })

  test('matchMode を localStorage から復元する', () => {
    const store: Record<string, string> = { pollMatchMode: 'partial' }
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((k) => store[k] ?? null)
    vi.spyOn(window.localStorage, 'setItem').mockImplementation((k, v) => {
      store[k] = String(v)
    })

    render(<HistoryVotes comments={comments} />)
    expect(screen.getByRole('button', { name: '部分一致' })).toHaveAttribute('aria-pressed', 'true')
  })
})
