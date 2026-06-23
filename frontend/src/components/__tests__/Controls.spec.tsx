import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { Controls } from '../Controls'

describe('Controls コンポーネント', () => {
  const mockProps = {
    videoId: '',
    setVideoId: vi.fn(),
    status: 'WAITING',

    loadingStates: {
      switching: false,
      pulling: false,
      resetting: false,
      refreshing: false,
    },
    onSwitch: vi.fn(),
    onPull: vi.fn(),
    onReset: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('videoId入力欄が正しく表示される', () => {
    render(<Controls {...mockProps} />)

    const input = screen.getByRole('textbox', { name: 'videoId' })
    expect(input).toBeInTheDocument()
    expect(input).toHaveAttribute('placeholder', 'videoId を入力')
    expect(input).toHaveValue('')
  })

  test('videoIdの値が正しく反映される', () => {
    render(<Controls {...mockProps} videoId="test-video-id" />)

    const input = screen.getByRole('textbox', { name: 'videoId' })
    expect(input).toHaveValue('test-video-id')
  })

  test('videoId入力時にsetVideoIdが呼ばれる', () => {
    render(<Controls {...mockProps} />)

    const input = screen.getByRole('textbox', { name: 'videoId' })
    fireEvent.change(input, { target: { value: 'new-video-id' } })

    expect(mockProps.setVideoId).toHaveBeenCalledWith('new-video-id')
  })

  test('WAITING 状態では「開始」ボタンとリセットボタンが表示される', () => {
    render(<Controls {...mockProps} />)

    expect(screen.getByRole('button', { name: '開始' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '今すぐ取得' })).not.toBeInTheDocument()
  })

  test('ACTIVE 状態かつ入力 videoId が現行と一致なら「今すぐ取得」になる', () => {
    render(<Controls {...mockProps} status="ACTIVE" videoId="vid001" currentVideoId="vid001" />)

    expect(screen.getByRole('button', { name: '今すぐ取得' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '開始' })).not.toBeInTheDocument()
  })

  test('ACTIVE 状態でも入力が空なら「今すぐ取得」', () => {
    render(<Controls {...mockProps} status="ACTIVE" videoId="" currentVideoId="vid001" />)

    expect(screen.getByRole('button', { name: '今すぐ取得' })).toBeInTheDocument()
  })

  test('ACTIVE 状態で別 videoId 入力中は「開始」(切替) になる', () => {
    render(<Controls {...mockProps} status="ACTIVE" videoId="vid_new" currentVideoId="vid_old" />)

    expect(screen.getByRole('button', { name: '開始' })).toBeInTheDocument()
  })

  test('RESERVED 状態では「開始」ボタンが disabled になる', () => {
    render(<Controls {...mockProps} status="RESERVED" videoId="vid001" />)

    const button = screen.getByRole('button', { name: '開始' })
    expect(button).toBeDisabled()
    const input = screen.getByRole('textbox', { name: 'videoId' })
    expect(input).toHaveAttribute('placeholder', '予約中 (キャンセルは curl)')
  })

  test('「開始」クリック時に onSwitch が呼ばれる', async () => {
    const mockOnSwitch = vi.fn().mockResolvedValue(undefined)
    render(<Controls {...mockProps} videoId="vid001" onSwitch={mockOnSwitch} />)

    const button = screen.getByRole('button', { name: '開始' })
    fireEvent.click(button)

    await waitFor(() => {
      expect(mockOnSwitch).toHaveBeenCalled()
    })
  })

  test('「今すぐ取得」クリック時に onPull が呼ばれる', async () => {
    const mockOnPull = vi.fn().mockResolvedValue(undefined)
    render(
      <Controls
        {...mockProps}
        status="ACTIVE"
        videoId="vid001"
        currentVideoId="vid001"
        onPull={mockOnPull}
      />,
    )

    const button = screen.getByRole('button', { name: '今すぐ取得' })
    fireEvent.click(button)

    await waitFor(() => {
      expect(mockOnPull).toHaveBeenCalled()
    })
  })

  test('リセットボタンクリック時に confirm を通過すると onReset が呼ばれる', async () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    render(<Controls {...mockProps} />)

    const resetButton = screen.getByRole('button', { name: 'リセット' })
    fireEvent.click(resetButton)

    await waitFor(() => {
      expect(mockProps.onReset).toHaveBeenCalled()
    })
    confirmSpy.mockRestore()
  })

  test('リセットボタンクリック時に confirm をキャンセルすると onReset は呼ばれない', () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    render(<Controls {...mockProps} />)

    const resetButton = screen.getByRole('button', { name: 'リセット' })
    fireEvent.click(resetButton)

    expect(mockProps.onReset).not.toHaveBeenCalled()
    confirmSpy.mockRestore()
  })

  test('switching 中はボタンと入力欄が無効化される', () => {
    const loadingStates = {
      switching: true,
      pulling: false,
      resetting: false,
      refreshing: false,
    }
    render(<Controls {...mockProps} loadingStates={loadingStates} />)

    const input = screen.getByRole('textbox', { name: 'videoId' })
    expect(input).toBeDisabled()
    const button = screen.getByRole('button', { name: /開始/ })
    expect(button).toHaveAttribute('aria-busy', 'true')
  })

  test('pulling 中は「取得中…」が表示される', () => {
    const loadingStates = {
      switching: false,
      pulling: true,
      resetting: false,
      refreshing: false,
    }
    render(
      <Controls
        {...mockProps}
        status="ACTIVE"
        videoId="vid001"
        currentVideoId="vid001"
        loadingStates={loadingStates}
      />,
    )

    expect(screen.getByRole('button', { name: '取得中…' })).toBeInTheDocument()
  })
})
