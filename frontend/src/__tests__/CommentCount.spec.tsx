import { render, screen } from '@testing-library/react'
import App from '../App.jsx'
import { __mock } from '../mocks/handlers'

describe.skip('発言数表示機能', () => {
  beforeEach(() => {
    // 初期状態をリセット
    __mock.state = 'WAITING'
    __mock.users = []
  })

  test('テーブルヘッダーに「発言数」列が表示される', async () => {
    render(<App />)
    expect(screen.getByText('発言数')).toBeInTheDocument()
  })

  test('発言数がない場合は「0」と表示される', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z'
      }
    ]

    render(<App />)

    // 発言数セルを検索（テストIDを使用）
    expect(await screen.findByTestId('comment-count-0')).toHaveTextContent('0')
  })

  test('発言数が設定されている場合は正しく表示される', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 5
      },
      {
        channelId: 'UC2',
        displayName: 'TestUser2',
        joinedAt: '2024-01-01T12:30:00Z',
        commentCount: 12
      }
    ]

    render(<App />)

    expect(await screen.findByTestId('comment-count-0')).toHaveTextContent('5')
    expect(screen.getByTestId('comment-count-1')).toHaveTextContent('12')
  })

  test('発言数が大きい数値でも正しく表示される', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'ActiveUser',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 999
      }
    ]

    render(<App />)

    expect(await screen.findByTestId('comment-count-0')).toHaveTextContent('999')
  })
})