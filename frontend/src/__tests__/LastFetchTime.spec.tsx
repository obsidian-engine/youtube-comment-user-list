import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import App from '../App.jsx'
import { __mock } from '../mocks/handlers'

describe('取得日時表示機能', () => {
  beforeEach(() => {
    // 初期状態をリセット
    __mock.state = 'WAITING'
    __mock.users = []
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  test('今すぐ取得ボタンの下に取得日時表示エリアが存在する', async () => {
    render(<App />)

    expect(screen.getByTestId('last-fetch-time')).toBeInTheDocument()
  })

  test('初期状態では取得日時が表示されていない', async () => {
    render(<App />)

    const fetchTimeElement = screen.getByTestId('last-fetch-time')
    expect(fetchTimeElement).toHaveTextContent('')
  })

  test('今すぐ取得ボタンクリック後に取得日時が表示される', async () => {
    __mock.state = 'ACTIVE'

    render(<App />)

    const pullButton = screen.getByText('今すぐ取得')
    fireEvent.click(pullButton)

    await waitFor(() => {
      const fetchTimeElement = screen.getByTestId('last-fetch-time')
      expect(fetchTimeElement).toHaveTextContent(/最終取得: \d{2}:\d{2}:\d{2}/)
    })
  })

  test('取得日時の形式が正しい', async () => {
    __mock.state = 'ACTIVE'

    render(<App />)

    const pullButton = screen.getByText('今すぐ取得')
    fireEvent.click(pullButton)

    await waitFor(() => {
      const fetchTimeElement = screen.getByTestId('last-fetch-time')
      // 正規表現で時刻フォーマット（HH:MM:SS）をチェック
      expect(fetchTimeElement).toHaveTextContent(/最終取得: \d{2}:\d{2}:\d{2}/)
    })
  })
})