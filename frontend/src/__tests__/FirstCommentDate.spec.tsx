import { render, screen, fireEvent, act } from '@testing-library/react'
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
        firstCommentedAt: '2024-01-01T12:05:30Z'
      },
      {
        channelId: 'UC2',
        displayName: 'TestUser2',
        joinedAt: '2024-01-01T12:30:00Z',
        firstCommentedAt: '2024-01-01T12:35:15Z'
      }
    ]

    render(<App />)

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('21:05')
    expect(screen.getByTestId('first-comment-1')).toHaveTextContent('21:35')
  })

  test('初回コメント日時の時刻フォーマットが正しい', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T09:05:00Z',
        firstCommentedAt: '2024-01-01T09:08:45Z'
      }
    ]

    render(<App />)

    // 時刻が18:08として表示されることを確認（日本時間フォーマット）
    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('18:08')
  })

  test('/pullエンドポイントで追加されたユーザーはfirstCommentAtが設定される', async () => {
    // 初期状態はユーザー無し
    __mock.users = []
    render(<App />)

    // 今すぐ取得ボタンをクリック（/pullエンドポイントを呼び出す）
    fireEvent.click(screen.getByRole('button', { name: /今すぐ取得/ }))

    // 追加されたユーザーにfirstCommentAtが設定されていることを確認
    // 修正後は時刻が表示されるはず
    expect(await screen.findByTestId('first-comment-0')).not.toHaveTextContent('--:--')
  })

  test('firstCommentedAtが空文字列の場合は「--:--」が表示される', async () => {
    __mock.users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: ''  // 空文字列
      }
    ]

    render(<App />)

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('--:--')
  })
})