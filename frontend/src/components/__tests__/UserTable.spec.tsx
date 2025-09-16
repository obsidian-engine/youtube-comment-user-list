import { render, screen } from '@testing-library/react'
import { UserTable } from '../UserTable'

describe('UserTable コンポーネント', () => {
  const mockUsers = [
    {
      channelId: 'UC1',
      displayName: 'TestUser1',
      joinedAt: '2024-01-01T12:00:00Z',
      commentCount: 5,
      firstCommentedAt: '2024-01-01T12:05:00Z'
    },
    {
      channelId: 'UC2',
      displayName: 'TestUser2',
      joinedAt: '2024-01-01T12:30:00Z',
      commentCount: 3,
      firstCommentedAt: '2024-01-01T12:32:00Z'
    }
  ]

  test('テーブルヘッダーが正しく表示される', () => {
    render(<UserTable users={mockUsers} />)

    expect(screen.getByText('#')).toBeInTheDocument()
    expect(screen.getByText('名前')).toBeInTheDocument()
    expect(screen.getByText('発言数')).toBeInTheDocument()
    expect(screen.getByText('初回コメント')).toBeInTheDocument()
    expect(screen.getByText('参加時間')).toBeInTheDocument()
  })

  test('ユーザー情報が正しく表示される', () => {
    render(<UserTable users={mockUsers} />)

    // 1番目のユーザー
    expect(screen.getByText('01')).toBeInTheDocument()
    expect(screen.getByText('TestUser1')).toBeInTheDocument()
    expect(screen.getByTestId('comment-count-0')).toHaveTextContent('5')

    // 2番目のユーザー
    expect(screen.getByText('02')).toBeInTheDocument()
    expect(screen.getByText('TestUser2')).toBeInTheDocument()
    expect(screen.getByTestId('comment-count-1')).toHaveTextContent('3')
  })

  test('初回コメント日時が正しく表示される', () => {
    render(<UserTable users={mockUsers} />)

    // 日本時間での表示を確認（UTCから+9時間）
    expect(screen.getByTestId('first-comment-0')).toHaveTextContent('21:05')
    expect(screen.getByTestId('first-comment-1')).toHaveTextContent('21:32')
  })

  test('参加時間が正しく表示される', () => {
    render(<UserTable users={mockUsers} />)

    // テストIDは生成されないが、内容で確認
    const rows = screen.getAllByRole('row')
    expect(rows[1]).toHaveTextContent('21:00') // 1番目のユーザーの参加時間
    expect(rows[2]).toHaveTextContent('21:30') // 2番目のユーザーの参加時間
  })

  test('firstCommentedAtがない場合--:--が表示される', () => {
    const usersWithoutFirstComment = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 0
      }
    ]
    render(<UserTable users={usersWithoutFirstComment} />)

    expect(screen.getByTestId('first-comment-0')).toHaveTextContent('--:--')
  })

  test('firstCommentedAtが空文字列の場合--:--が表示される', () => {
    const usersWithEmptyFirstComment = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 0,
        firstCommentedAt: ''
      }
    ]
    render(<UserTable users={usersWithEmptyFirstComment} />)

    expect(screen.getByTestId('first-comment-0')).toHaveTextContent('--:--')
  })

  test('commentCountがundefinedの場合0が表示される', () => {
    const usersWithoutCommentCount = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z'
      }
    ]
    render(<UserTable users={usersWithoutCommentCount} />)

    expect(screen.getByTestId('comment-count-0')).toHaveTextContent('0')
  })

  test('ユーザーがいない場合メッセージが表示される', () => {
    render(<UserTable users={[]} />)

    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
  })

  test('行のストライプ表示が正しく適用される', () => {
    render(<UserTable users={mockUsers} />)

    const rows = screen.getAllByRole('row')
    // ヘッダーを除く最初のデータ行（偶数行）
    expect(rows[1]).toHaveClass('bg-slate-100/50')
    // 2番目のデータ行（奇数行）
    expect(rows[2]).toHaveClass('bg-slate-200/40')
  })

  test('displayNameがない場合の fallback が動作する', () => {
    const usersWithoutDisplayName = [
      {
        channelId: 'UC1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 1
      }
    ]
    render(<UserTable users={usersWithoutDisplayName} />)

    // displayNameがない場合、channelIdが表示される
    expect(screen.getByText('UC1')).toBeInTheDocument()
  })

  test('行番号が正しくゼロパディングされる', () => {
    const manyUsers = Array.from({ length: 15 }, (_, i) => ({
      channelId: `UC${i + 1}`,
      displayName: `User${i + 1}`,
      joinedAt: '2024-01-01T12:00:00Z',
      commentCount: 1
    }))
    render(<UserTable users={manyUsers} />)

    expect(screen.getByText('01')).toBeInTheDocument()
    expect(screen.getByText('09')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('15')).toBeInTheDocument()
  })
})