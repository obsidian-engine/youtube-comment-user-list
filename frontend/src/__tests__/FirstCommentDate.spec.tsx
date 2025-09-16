import { render, screen } from '@testing-library/react'
import App from '../App.jsx'
import { __mock } from '../mocks/handlers'

describe('初回コメント日時保持・表示機能', () => {
  beforeEach(() => {
    // 初期状態をリセット
    __mock.state = 'WAITING'
    __mock.users = []
  })

  test('テーブルヘッダーに「初回コメント」列が表示される', async () => {
    render(<App />)
    expect(screen.getByText('初回コメント')).toBeInTheDocument()
  })

  test('初回コメント日時がない場合は「--:--」と表示される', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z'
      }
    ]

    render(<App />)

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('--:--')
  })

  test('初回コメント日時が設定されている場合は正しく表示される', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentAt: '2024-01-01T12:05:30Z'
      },
      {
        channelId: 'UC2',
        displayName: 'TestUser2',
        joinedAt: '2024-01-01T12:30:00Z',
        firstCommentAt: '2024-01-01T12:35:15Z'
      }
    ]

    render(<App />)

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('12:05')
    expect(screen.getByTestId('first-comment-1')).toHaveTextContent('12:35')
  })

  test('初回コメント日時の時刻フォーマットが正しい', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T09:05:00Z',
        firstCommentAt: '2024-01-01T09:08:45Z'
      }
    ]

    render(<App />)

    // 時刻が09:08として表示されることを確認（日本時間フォーマット）
    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('18:08')
  })
})