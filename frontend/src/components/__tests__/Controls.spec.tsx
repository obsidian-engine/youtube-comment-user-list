import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { Controls } from '../Controls'

describe('Controls コンポーネント', () => {
  const mockProps = {
    videoId: '',
    setVideoId: vi.fn(),
    loadingStates: {
      switching: false,
      pulling: false,
      resetting: false,
      refreshing: false
    },
    onSwitch: vi.fn(),
    onPull: vi.fn(),
    onReset: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('videoId入力欄が正しく表示される', () => {
    render(<Controls {...mockProps} />)

    const input = screen.getByLabelText('videoId')
    expect(input).toBeInTheDocument()
    expect(input).toHaveAttribute('placeholder', 'videoId を入力')
    expect(input).toHaveValue('')
  })

  test('videoIdの値が正しく反映される', () => {
    render(<Controls {...mockProps} videoId="test-video-id" />)

    const input = screen.getByLabelText('videoId')
    expect(input).toHaveValue('test-video-id')
  })

  test('videoId入力時にsetVideoIdが呼ばれる', () => {
    render(<Controls {...mockProps} />)

    const input = screen.getByLabelText('videoId')
    fireEvent.change(input, { target: { value: 'new-video-id' } })

    expect(mockProps.setVideoId).toHaveBeenCalledWith('new-video-id')
  })

  test('操作ボタンが正しく表示される', () => {
    render(<Controls {...mockProps} />)

    expect(screen.getByLabelText('切替')).toBeInTheDocument()
    expect(screen.getByLabelText('今すぐ取得')).toBeInTheDocument()
    expect(screen.getByLabelText('リセット')).toBeInTheDocument()
  })

  test('切替ボタンクリック時にonSwitchが呼ばれる', async () => {
    render(<Controls {...mockProps} videoId="test-video" />)

    const switchButton = screen.getByLabelText('切替')
    fireEvent.click(switchButton)

    await waitFor(() => {
      expect(mockProps.onSwitch).toHaveBeenCalled()
    })
  })

  test('切替ボタンクリック時にonSwitchが呼ばれる', async () => {
    const mockOnSwitch = vi.fn().mockResolvedValue(undefined)
    render(<Controls {...mockProps} videoId="test-video-id" onSwitch={mockOnSwitch} />)

    const switchButton = screen.getByLabelText('切替')
    fireEvent.click(switchButton)

    await waitFor(() => {
      expect(mockOnSwitch).toHaveBeenCalled()
    })
  })

  test('今すぐ取得ボタンクリック時にonPullが呼ばれる', async () => {
    render(<Controls {...mockProps} />)

    const pullButton = screen.getByLabelText('今すぐ取得')
    fireEvent.click(pullButton)

    await waitFor(() => {
      expect(mockProps.onPull).toHaveBeenCalled()
    })
  })

  test('リセットボタンクリック時にonResetが呼ばれる', async () => {
    render(<Controls {...mockProps} />)

    const resetButton = screen.getByLabelText('リセット')
    fireEvent.click(resetButton)

    await waitFor(() => {
      expect(mockProps.onReset).toHaveBeenCalled()
    })
  })

  test('ローディング状態でボタンが無効化される', () => {
    const loadingStates = {
      switching: true,
      pulling: false,
      resetting: false,
      refreshing: false
    }
    render(<Controls {...mockProps} loadingStates={loadingStates} />)

    const input = screen.getByLabelText('videoId')
    const switchButton = screen.getByRole('button', { name: /切替/ })

    expect(input).toBeDisabled()
    expect(switchButton).toHaveAttribute('aria-busy', 'true')
  })

  test('ローディング中の表示が正しい', () => {
    const loadingStates = {
      switching: false,
      pulling: true,
      resetting: false,
      refreshing: false
    }
    render(<Controls {...mockProps} loadingStates={loadingStates} />)

    expect(screen.getByText('取得中…')).toBeInTheDocument()
  })
})