import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from '../App.jsx'
import { __mock } from '../mocks/handlers'

describe('取得日時表示機能', () => {
  beforeEach(() => {
    // 初期状態をリセット
    __mock.state = 'WAITING'
    __mock.users = []
    // モックの時刻を固定
    jest.spyOn(Date, 'now').mockReturnValue(new Date('2024-01-01T12:00:00Z').getTime())
  })

  afterEach(() => {
    jest.restoreAllMocks()
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
    const user = userEvent.setup()
    __mock.state = 'ACTIVE'

    render(<App />)

    const pullButton = screen.getByText('今すぐ取得')
    await user.click(pullButton)

    await waitFor(() => {
      const fetchTimeElement = screen.getByTestId('last-fetch-time')
      expect(fetchTimeElement).toHaveTextContent(/最終取得: \d{2}:\d{2}:\d{2}/)
    })
  })

  test('取得日時の形式が正しい', async () => {
    const user = userEvent.setup()
    __mock.state = 'ACTIVE'

    // 特定の時刻にモック
    jest.spyOn(Date, 'now').mockReturnValue(new Date('2024-01-01T15:30:45Z').getTime())

    render(<App />)

    const pullButton = screen.getByText('今すぐ取得')
    await user.click(pullButton)

    await waitFor(() => {
      const fetchTimeElement = screen.getByTestId('last-fetch-time')
      expect(fetchTimeElement).toHaveTextContent('最終取得: 15:30:45')
    })
  })
})