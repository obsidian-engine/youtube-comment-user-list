import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import App from '../App'

// API mock
vi.mock('../utils/api', () => ({
  getStatus: vi.fn().mockResolvedValue({ status: 'ACTIVE', count: 0 }),
  getUsers: vi.fn().mockResolvedValue([]),
  postSwitchVideo: vi.fn().mockResolvedValue({}),
  postPull: vi
    .fn()
    .mockResolvedValue({
      addedCount: 0,
      skippedCount: 0,
      autoReset: false,
      pollingIntervalMillis: 15000,
    }),
  postReset: vi.fn().mockResolvedValue({}),
  searchComments: vi.fn().mockResolvedValue([]),
  HttpError: class HttpError extends Error {
    status: number
    constructor(status: number) {
      super(`HTTP ${status}`)
      this.status = status
    }
  },
  BackendError: class BackendError extends Error {},
}))

// useHistory mock
vi.mock('../hooks/useHistory', () => ({
  useHistory: () => ({
    snapshots: [],
    selected: null,
    loading: false,
    error: '',
    loadList: vi.fn().mockResolvedValue(undefined),
    select: vi.fn(),
    clearSelected: vi.fn(),
  }),
}))

const clearCommentsSpy = vi.fn()
const clearResultsSpy = vi.fn()

vi.mock('../hooks/useCommentSearch', () => ({
  useCommentSearch: () => ({
    keywords: [],
    comments: [],
    isLoading: false,
    errorMsg: null,
    lastUpdated: null,
    intervalSec: 0,
    addKeyword: vi.fn(),
    removeKeyword: vi.fn(),
    setIntervalSec: vi.fn(),
    search: vi.fn(),
    clearComments: clearCommentsSpy,
  }),
}))

vi.mock('../hooks/usePollCount', () => ({
  POLL_INTERVAL_SEC: 15,
  usePollCount: () => ({
    keywords: [],
    counts: {},
    voters: {},
    totalVotes: 0,
    isLoading: false,
    errorMsg: '',
    lastUpdated: '--:--:--',
    addKeyword: vi.fn(),
    removeKeyword: vi.fn(),
    clearKeywords: vi.fn(),
    clearResults: clearResultsSpy,
    recount: vi.fn(),
  }),
}))

describe('App - handleSwitch', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    localStorage.setItem('activeTab', 'users')
  })

  it('切替ボタン押下時に clearComments が呼ばれる', async () => {
    render(<App />)

    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID_NEW' } })
    fireEvent.click(screen.getByRole('button', { name: '開始' }))

    await waitFor(() => {
      expect(clearCommentsSpy).toHaveBeenCalledTimes(1)
    })
  })

  it('切替ボタン押下時に clearResults が呼ばれる', async () => {
    render(<App />)

    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID_NEW' } })
    fireEvent.click(screen.getByRole('button', { name: '開始' }))

    await waitFor(() => {
      expect(clearResultsSpy).toHaveBeenCalledTimes(1)
    })
  })

  it('切替ボタン押下時に clearComments と clearResults が両方呼ばれる', async () => {
    render(<App />)

    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID_NEW' } })
    fireEvent.click(screen.getByRole('button', { name: '開始' }))

    await waitFor(() => {
      expect(clearCommentsSpy).toHaveBeenCalledTimes(1)
      expect(clearResultsSpy).toHaveBeenCalledTimes(1)
    })
  })
})
