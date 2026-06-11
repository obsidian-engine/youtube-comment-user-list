import { render, screen, fireEvent } from '@testing-library/react'
import { vi, beforeEach, describe, test, expect } from 'vitest'
import { HistoryTab } from '../HistoryTab'

const mockLoadList = vi.fn()
const mockSelect = vi.fn()
const mockClearSelected = vi.fn()

const baseHookReturn = {
  snapshots: [],
  selected: null,
  loading: false,
  error: '',
  loadList: mockLoadList,
  select: mockSelect,
  clearSelected: mockClearSelected,
}

vi.mock('../../hooks/useHistory', () => ({
  useHistory: vi.fn(),
}))

import { useHistory } from '../../hooks/useHistory'
const mockUseHistory = vi.mocked(useHistory)

beforeEach(() => {
  vi.clearAllMocks()
  mockUseHistory.mockReturnValue(baseHookReturn)
})

describe('HistoryTab', () => {
  test('初回 render で loadList を呼び出す', () => {
    render(<HistoryTab />)
    expect(mockLoadList).toHaveBeenCalledTimes(1)
  })

  test('loading=true のとき読み込み中表示が出る', () => {
    mockUseHistory.mockReturnValue({ ...baseHookReturn, loading: true })
    render(<HistoryTab />)
    expect(screen.getByText(/読み込み中/)).toBeInTheDocument()
  })

  test('snapshots 2件のときリスト table に 2行表示される', () => {
    mockUseHistory.mockReturnValue({
      ...baseHookReturn,
      snapshots: [
        {
          videoId: 'vid1',
          savedAt: '2024-06-01T10:00:00Z',
          userCount: 10,
          commentCount: 5,
          videoTitle: '動画タイトル1',
        },
        {
          videoId: 'vid2',
          savedAt: '2024-06-02T12:00:00Z',
          userCount: 20,
          commentCount: 8,
          videoTitle: '動画タイトル2',
        },
      ],
    })
    render(<HistoryTab />)
    expect(screen.getByText('vid1')).toBeInTheDocument()
    expect(screen.getByText('vid2')).toBeInTheDocument()
    // 行数確認: 「表示」ボタンが 2 個
    const showButtons = screen.getAllByRole('button', { name: '表示' })
    expect(showButtons).toHaveLength(2)
  })

  test('row の「表示」クリックで select(videoId) が呼ばれ detail 表示に切り替わる', () => {
    const mockSnapshot = {
      videoId: 'vid1',
      savedAt: '2024-06-01T10:00:00Z',
      users: [
        {
          channelId: 'UC1',
          displayName: 'User1',
          joinedAt: '2024-06-01T09:00:00Z',
          commentCount: 2,
        },
      ],
      comments: [],
    }
    // まず list 表示
    mockUseHistory.mockReturnValue({
      ...baseHookReturn,
      snapshots: [
        { videoId: 'vid1', savedAt: '2024-06-01T10:00:00Z', userCount: 1, commentCount: 0 },
      ],
    })
    const { rerender } = render(<HistoryTab />)
    fireEvent.click(screen.getByRole('button', { name: '表示' }))
    expect(mockSelect).toHaveBeenCalledWith('vid1')

    // selected がセットされた状態を simulate
    mockUseHistory.mockReturnValue({ ...baseHookReturn, selected: mockSnapshot })
    rerender(<HistoryTab />)
    expect(screen.getByText('User1')).toBeInTheDocument()
  })

  test('detail の「戻る」ボタンクリックで clearSelected が呼ばれる', () => {
    const mockSnapshot = {
      videoId: 'vid1',
      savedAt: '2024-06-01T10:00:00Z',
      users: [],
      comments: [],
    }
    mockUseHistory.mockReturnValue({ ...baseHookReturn, selected: mockSnapshot })
    render(<HistoryTab />)
    fireEvent.click(screen.getByRole('button', { name: '戻る' }))
    expect(mockClearSelected).toHaveBeenCalledTimes(1)
  })

  test('error 表示: error メッセージが role=alert で表示される', () => {
    mockUseHistory.mockReturnValue({ ...baseHookReturn, error: '履歴の取得に失敗しました' })
    render(<HistoryTab />)
    expect(screen.getByRole('alert')).toHaveTextContent('履歴の取得に失敗しました')
  })

  test('detail mode で「閲覧モード」badge が表示される', () => {
    const mockSnapshot = {
      videoId: 'vid1',
      savedAt: '2024-06-01T10:00:00Z',
      users: [],
      comments: [],
    }
    mockUseHistory.mockReturnValue({ ...baseHookReturn, selected: mockSnapshot })
    render(<HistoryTab />)
    expect(screen.getByText(/閲覧モード/)).toBeInTheDocument()
  })
})
