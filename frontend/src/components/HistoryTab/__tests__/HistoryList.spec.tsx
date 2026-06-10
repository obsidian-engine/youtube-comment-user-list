import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, test, expect, beforeEach } from 'vitest'
import { HistoryList } from '../HistoryList'
import type { HistorySummary } from '../../../utils/api'

vi.mock('../../../hooks/useAppState', () => ({
  formatSnapshotSavedAt: (s: string) => s,
}))

const mockSnapshots: HistorySummary[] = [
  { videoId: 'vid-1', savedAt: '2024-06-01T10:00:00Z', userCount: 10, commentCount: 50 },
  { videoId: 'vid-2', savedAt: '2024-06-15T11:00:00Z', userCount: 5, commentCount: 20 },
  { videoId: 'vid-3', savedAt: '2024-07-01T09:00:00Z', userCount: 3, commentCount: 10 },
]

const noop = vi.fn().mockResolvedValue(undefined)

describe('HistoryList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // --- 既存テスト ---

  test('loading 中はスピナーを表示する', () => {
    render(<HistoryList snapshots={[]} loading={true} error="" onSelect={noop} />)
    expect(screen.getByTestId('history-loading-spinner')).toBeInTheDocument()
  })

  test('error があればアラートを表示する', () => {
    render(<HistoryList snapshots={[]} loading={false} error="エラーです" onSelect={noop} />)
    expect(screen.getByRole('alert')).toHaveTextContent('エラーです')
  })

  test('snapshots が空 (filter なし) は「履歴がありません」を表示する', () => {
    render(<HistoryList snapshots={[]} loading={false} error="" onSelect={noop} />)
    expect(screen.getByText('履歴がありません')).toBeInTheDocument()
  })

  test('snapshots がある場合は全件表示する', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)
    expect(screen.getByText('vid-1')).toBeInTheDocument()
    expect(screen.getByText('vid-2')).toBeInTheDocument()
    expect(screen.getByText('vid-3')).toBeInTheDocument()
  })

  // --- filter テスト ---

  test('from のみ指定すると以降の件数に絞られる', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)

    const fromInputs = screen.getAllByDisplayValue('')
    // From input は最初の date input
    fireEvent.change(fromInputs[0], { target: { value: '2024-06-15' } })

    expect(screen.queryByText('vid-1')).not.toBeInTheDocument()
    expect(screen.getByText('vid-2')).toBeInTheDocument()
    expect(screen.getByText('vid-3')).toBeInTheDocument()
  })

  test('to のみ指定すると以前の件数に絞られる', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)

    const labels = screen.getAllByDisplayValue('')
    // To input は2番目の date input
    fireEvent.change(labels[1], { target: { value: '2024-06-15' } })

    expect(screen.getByText('vid-1')).toBeInTheDocument()
    expect(screen.getByText('vid-2')).toBeInTheDocument()
    expect(screen.queryByText('vid-3')).not.toBeInTheDocument()
  })

  test('from と to 両方指定すると範囲内のみ表示される', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)

    const inputs = screen.getAllByDisplayValue('')
    fireEvent.change(inputs[0], { target: { value: '2024-06-10' } })
    fireEvent.change(inputs[1], { target: { value: '2024-06-20' } })

    expect(screen.queryByText('vid-1')).not.toBeInTheDocument()
    expect(screen.getByText('vid-2')).toBeInTheDocument()
    expect(screen.queryByText('vid-3')).not.toBeInTheDocument()
  })

  test('範囲外で 0 件になると「該当する履歴がありません」を表示する', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)

    const inputs = screen.getAllByDisplayValue('')
    fireEvent.change(inputs[0], { target: { value: '2025-01-01' } })

    expect(screen.getByText('該当する履歴がありません')).toBeInTheDocument()
    expect(screen.queryByText('履歴がありません')).not.toBeInTheDocument()
  })

  test('クリアボタンを押すと全件に戻る', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)

    const inputs = screen.getAllByDisplayValue('')
    fireEvent.change(inputs[0], { target: { value: '2025-01-01' } })
    expect(screen.getByText('該当する履歴がありません')).toBeInTheDocument()

    const clearButton = screen.getByRole('button', { name: 'クリア' })
    fireEvent.click(clearButton)

    expect(screen.getByText('vid-1')).toBeInTheDocument()
    expect(screen.getByText('vid-2')).toBeInTheDocument()
    expect(screen.getByText('vid-3')).toBeInTheDocument()
  })

  test('filter が空のときはクリアボタンを表示しない', () => {
    render(<HistoryList snapshots={mockSnapshots} loading={false} error="" onSelect={noop} />)
    expect(screen.queryByRole('button', { name: 'クリア' })).not.toBeInTheDocument()
  })
})
