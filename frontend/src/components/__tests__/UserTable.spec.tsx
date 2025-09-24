import { render, screen, fireEvent, within } from '@testing-library/react'
import { vi } from 'vitest'
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
    expect(screen.getByText('最新コメント')).toBeInTheDocument()
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

  test('最新コメント時間が正しく表示される', () => {
    const usersWithLatestComment = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 5,
        firstCommentedAt: '2024-01-01T12:05:00Z',
        latestCommentedAt: '2024-01-01T15:30:00Z' // 最新コメント時間
      },
      {
        channelId: 'UC2',
        displayName: 'TestUser2',
        joinedAt: '2024-01-01T12:30:00Z',
        commentCount: 3,
        firstCommentedAt: '2024-01-01T12:32:00Z',
        latestCommentedAt: '2024-01-01T14:15:00Z' // 最新コメント時間
      }
    ]
    render(<UserTable users={usersWithLatestComment} />)

    // 日本時間での表示を確認（UTCから+9時間）
    expect(screen.getByTestId('latest-comment-0')).toHaveTextContent('00:30')
    expect(screen.getByTestId('latest-comment-1')).toHaveTextContent('23:15')
  })

  test('最新コメント時間がない場合--:--が表示される', () => {
    render(<UserTable users={mockUsers} />)

    // latestCommentedAtが未定義の場合
    expect(screen.getByTestId('latest-comment-0')).toHaveTextContent('--:--')
    expect(screen.getByTestId('latest-comment-1')).toHaveTextContent('--:--')
  })

  test('リフレッシュ中のローディング状態が正しく表示される', () => {
    render(<UserTable users={mockUsers} isRefreshing={true} />)
    
    // 大きなスピナーが表示される
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    expect(screen.getByTestId('loading-spinner')).toHaveClass('h-8', 'w-8')
    
    // 「データ更新中...」テキストが表示される
    expect(screen.getByText('データ更新中...')).toBeInTheDocument()
  })

  test('リフレッシュ中でない場合はローディング要素が表示されない', () => {
    render(<UserTable users={mockUsers} isRefreshing={false} />)
    
    // ローディング要素が表示されない
    expect(screen.queryByTestId('loading-spinner')).not.toBeInTheDocument()
    expect(screen.queryByText('データ更新中...')).not.toBeInTheDocument()
  })

  test('リフレッシュ中は更新間隔セレクトボックスが無効化される', () => {
    const mockSetIntervalSec = vi.fn()
    render(
      <UserTable 
        users={mockUsers} 
        intervalSec={30} 
        setIntervalSec={mockSetIntervalSec}
        isRefreshing={true} 
      />
    )
    
    // セレクトボックスが無効化される
    expect(screen.getByLabelText('更新間隔')).toBeDisabled()
  })

  test('リフレッシュ中はデフォルト順ボタンが無効化される', () => {
    render(<UserTable users={mockUsers} isRefreshing={true} />)
    
    // デフォルト順ボタンが無効化される
    expect(screen.getByLabelText('ソートリセット')).toBeDisabled()
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

  test('狭い画面では初回コメントと最新コメントカラムが非表示になる', () => {
    // 画面幅を狭くシミュレート（CSSメディアクエリはJSでテストできないので、クラス名で確認）
    render(<UserTable users={mockUsers} />)

    const table = screen.getByRole('table')
    
    // 初回コメントと最新コメントのヘッダーにレスポンシブクラスが適用されていることを確認
    const firstCommentHeader = screen.getByText('初回コメント').closest('th')
    const latestCommentHeader = screen.getByText('最新コメント').closest('th')
    
    expect(firstCommentHeader).toHaveClass('hidden', 'md:table-cell')
    expect(latestCommentHeader).toHaveClass('hidden', 'md:table-cell')
  })

  test('狭い画面では初回コメントと最新コメントのセルが非表示になる', () => {
    render(<UserTable users={mockUsers} />)

    // 初回コメントと最新コメントのセルにレスポンシブクラスが適用されていることを確認
    const firstCommentCell = screen.getByTestId('first-comment-0')
    const latestCommentCell = screen.getByTestId('latest-comment-0')
    
    expect(firstCommentCell).toHaveClass('hidden', 'md:table-cell')
    expect(latestCommentCell).toHaveClass('hidden', 'md:table-cell')
  })

  describe('コメント時間カラム表示切り替え', () => {
    test('showCommentTimeがtrueの場合、初回コメントと最新コメントカラムが表示される', () => {
      render(<UserTable users={mockUsers} showCommentTime={true} />)

      // ヘッダーが表示される
      expect(screen.getByText('初回コメント')).toBeInTheDocument()
      expect(screen.getByText('最新コメント')).toBeInTheDocument()

      // データセルが表示される
      expect(screen.getByTestId('first-comment-0')).toBeInTheDocument()
      expect(screen.getByTestId('latest-comment-0')).toBeInTheDocument()
    })

    test('showCommentTimeがfalseの場合、初回コメントと最新コメントカラムが非表示になる', () => {
      render(<UserTable users={mockUsers} showCommentTime={false} />)

      // ヘッダーが非表示になる
      expect(screen.queryByText('初回コメント')).not.toBeInTheDocument()
      expect(screen.queryByText('最新コメント')).not.toBeInTheDocument()

      // データセルが非表示になる
      expect(screen.queryByTestId('first-comment-0')).not.toBeInTheDocument()
      expect(screen.queryByTestId('latest-comment-0')).not.toBeInTheDocument()
    })

    test('showCommentTimeが未指定の場合、デフォルトでカラムが表示される', () => {
      render(<UserTable users={mockUsers} />)

      // ヘッダーがデフォルトで表示される
      expect(screen.getByText('初回コメント')).toBeInTheDocument()
      expect(screen.getByText('最新コメント')).toBeInTheDocument()

      // データセルがデフォルトで表示される
      expect(screen.getByTestId('first-comment-0')).toBeInTheDocument()
      expect(screen.getByTestId('latest-comment-0')).toBeInTheDocument()
    })

    test('showCommentTimeがfalseの場合でもレスポンシブクラスは適用されない', () => {
      render(<UserTable users={mockUsers} showCommentTime={false} />)

      // カラムが非表示の場合、レスポンシブクラスによる制御は不要
      // 代わりにカラム自体が条件付きレンダリングで非表示になる
      expect(screen.queryByText('初回コメント')).not.toBeInTheDocument()
      expect(screen.queryByText('最新コメント')).not.toBeInTheDocument()
    })

    test('表示切り替えトグルボタンが表示される', () => {
      render(<UserTable users={mockUsers} showCommentTime={true} onToggleCommentTime={() => {}} />)

      const toggleButton = screen.getByRole('button', { name: /コメント時間表示切り替え/ })
      expect(toggleButton).toBeInTheDocument()
    })

    test('表示切り替えトグルボタンクリックでコールバックが呼ばれる', () => {
      const mockToggle = vi.fn()
      render(<UserTable users={mockUsers} showCommentTime={true} onToggleCommentTime={mockToggle} />)

      const toggleButton = screen.getByRole('button', { name: /コメント時間表示切り替え/ })
      fireEvent.click(toggleButton)

      expect(mockToggle).toHaveBeenCalledTimes(1)
    })

    test('onToggleCommentTimeが未指定の場合、トグルボタンが表示されない', () => {
      render(<UserTable users={mockUsers} showCommentTime={true} />)

      expect(screen.queryByRole('button', { name: /コメント時間表示切り替え/ })).not.toBeInTheDocument()
    })
  })

  describe('ソート機能', () => {
    const sortTestUsers = [
      {
        channelId: 'UC1',
        displayName: 'User1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 5,
        firstCommentedAt: '2024-01-01T12:05:00Z'
      },
      {
        channelId: 'UC2',
        displayName: 'User2',
        joinedAt: '2024-01-01T12:10:00Z',
        commentCount: 2,
        firstCommentedAt: '2024-01-01T12:15:00Z'
      },
      {
        channelId: 'UC3',
        displayName: 'User3',
        joinedAt: '2024-01-01T12:05:00Z',
        commentCount: 8,
        firstCommentedAt: '2024-01-01T12:08:00Z'
      }
    ]

    test('発言数ヘッダーにソートボタンが表示される', () => {
      render(<UserTable users={sortTestUsers} />)

      const commentCountHeader = screen.getByText('発言数').closest('th')
      const sortButton = within(commentCountHeader!).getByRole('button', { name: /発言数でソート/ })
      expect(sortButton).toBeInTheDocument()
    })

    test('初回コメントヘッダーにソートボタンが表示される', () => {
      render(<UserTable users={sortTestUsers} />)

      const firstCommentHeader = screen.getByText('初回コメント').closest('th')
      const sortButton = within(firstCommentHeader!).getByRole('button', { name: /初回コメントでソート/ })
      expect(sortButton).toBeInTheDocument()
    })

    test('発言数ソートボタンクリックで降順にソートされる', () => {
      render(<UserTable users={sortTestUsers} />)

      const commentCountSortButton = within(screen.getByText('発言数').closest('th')!).getByRole('button')
      fireEvent.click(commentCountSortButton)

      const rows = screen.getAllByRole('row')
      // ヘッダー行を除く最初のデータ行をチェック
      expect(rows[1]).toHaveTextContent('User3') // commentCount: 8
      expect(rows[2]).toHaveTextContent('User1') // commentCount: 5
      expect(rows[3]).toHaveTextContent('User2') // commentCount: 2
    })

    test('発言数ソートボタン2回クリックで昇順にソートされる', () => {
      render(<UserTable users={sortTestUsers} />)

      const commentCountSortButton = within(screen.getByText('発言数').closest('th')!).getByRole('button')
      fireEvent.click(commentCountSortButton) // 降順
      fireEvent.click(commentCountSortButton) // 昇順

      const rows = screen.getAllByRole('row')
      expect(rows[1]).toHaveTextContent('User2') // commentCount: 2
      expect(rows[2]).toHaveTextContent('User1') // commentCount: 5
      expect(rows[3]).toHaveTextContent('User3') // commentCount: 8
    })

    test('初回コメントソートボタンクリックで昇順にソートされる', () => {
      render(<UserTable users={sortTestUsers} />)

      const firstCommentSortButton = within(screen.getByText('初回コメント').closest('th')!).getByRole('button')
      fireEvent.click(firstCommentSortButton)

      const rows = screen.getAllByRole('row')
      expect(rows[1]).toHaveTextContent('User1') // 12:05
      expect(rows[2]).toHaveTextContent('User3') // 12:08
      expect(rows[3]).toHaveTextContent('User2') // 12:15
    })

    test('初回コメントソートボタン2回クリックで降順にソートされる', () => {
      render(<UserTable users={sortTestUsers} />)

      const firstCommentSortButton = within(screen.getByText('初回コメント').closest('th')!).getByRole('button')
      fireEvent.click(firstCommentSortButton) // 昇順
      fireEvent.click(firstCommentSortButton) // 降順

      const rows = screen.getAllByRole('row')
      expect(rows[1]).toHaveTextContent('User2') // 12:15
      expect(rows[2]).toHaveTextContent('User3') // 12:08
      expect(rows[3]).toHaveTextContent('User1') // 12:05
    })

    test('ソート後も行番号が正しく表示される', () => {
      render(<UserTable users={sortTestUsers} />)

      const commentCountSortButton = within(screen.getByText('発言数').closest('th')!).getByRole('button')
      fireEvent.click(commentCountSortButton)

      expect(screen.getByText('01')).toBeInTheDocument()
      expect(screen.getByText('02')).toBeInTheDocument()
      expect(screen.getByText('03')).toBeInTheDocument()
    })

    test('ソートリセットボタンが表示される', () => {
      render(<UserTable users={sortTestUsers} />)

      const resetButton = screen.getByRole('button', { name: 'ソートリセット' })
      expect(resetButton).toBeInTheDocument()
    })

    test('ソートリセットボタンクリックで初期表示順に戻る', () => {
      render(<UserTable users={sortTestUsers} />)

      // まず発言数でソート
      const commentCountSortButton = within(screen.getByText('発言数').closest('th')!).getByRole('button')
      fireEvent.click(commentCountSortButton)

      // ソート後の確認
      let rows = screen.getAllByRole('row')
      expect(rows[1]).toHaveTextContent('User3') // commentCount: 8

      // リセットボタンクリック
      const resetButton = screen.getByRole('button', { name: 'ソートリセット' })
      fireEvent.click(resetButton)

      // 初期表示順（props順）に戻ることを確認
      rows = screen.getAllByRole('row')
      expect(rows[1]).toHaveTextContent('User1') // 元の順序
      expect(rows[2]).toHaveTextContent('User2')
      expect(rows[3]).toHaveTextContent('User3')
    })

    test('初期状態ではソートリセットボタンが無効化されている', () => {
      render(<UserTable users={sortTestUsers} />)

      const resetButton = screen.getByRole('button', { name: 'ソートリセット' })
      expect(resetButton).toBeDisabled()
    })

    test('ソート後はソートリセットボタンが有効化される', () => {
      render(<UserTable users={sortTestUsers} />)

      const resetButton = screen.getByRole('button', { name: 'ソートリセット' })
      expect(resetButton).toBeDisabled()

      // 発言数でソート
      const commentCountSortButton = within(screen.getByText('発言数').closest('th')!).getByRole('button')
      fireEvent.click(commentCountSortButton)

      expect(resetButton).not.toBeDisabled()
    })

    test('リセット後は再びソートリセットボタンが無効化される', () => {
      render(<UserTable users={sortTestUsers} />)

      // ソート実行
      const commentCountSortButton = within(screen.getByText('発言数').closest('th')!).getByRole('button')
      fireEvent.click(commentCountSortButton)

      const resetButton = screen.getByRole('button', { name: 'ソートリセット' })
      expect(resetButton).not.toBeDisabled()

      // リセット実行
      fireEvent.click(resetButton)

      expect(resetButton).toBeDisabled()
    })
  })
})