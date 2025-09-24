import { render, screen } from '@testing-library/react'
import { StatsCard } from '../StatsCard'
import type { User } from '../../utils/api'

// テスト用のモックデータ
const mockUsers: User[] = [
  {
    channelId: 'channel1',
    displayName: 'User1',
    joinedAt: '2024-01-01T10:00:00Z',
    firstCommentedAt: '2024-01-01T10:05:00Z',
    commentCount: 5
  },
  {
    channelId: 'channel2', 
    displayName: 'User2',
    joinedAt: '2024-01-01T10:01:00Z',
    firstCommentedAt: '2024-01-01T10:10:00Z',
    commentCount: 3
  },
  {
    channelId: 'channel3',
    displayName: 'User3', 
    joinedAt: '2024-01-01T10:02:00Z',
    // コメントなし（firstCommentedAt, commentCount なし）
  }
]

const mockEmptyUsers: User[] = []

describe('StatsCard', () => {
  beforeEach(() => {
    // 現在時刻を固定 (2024-01-01T10:30:00Z)
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-01-01T10:30:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('総ユーザー数表示', () => {
    test('ユーザーが3人の場合、総ユーザー数が3人と表示される', () => {
      render(<StatsCard users={mockUsers} active={true} />)
      
      expect(screen.getByText('3')).toBeInTheDocument()
      expect(screen.getByText('人')).toBeInTheDocument()
    })

    test('ユーザーが0人の場合、総ユーザー数が0人と表示される', () => {
      render(<StatsCard users={mockEmptyUsers} active={true} />)
      
      expect(screen.getByText('0')).toBeInTheDocument()
    })
  })

  describe('監視時間表示', () => {
    test('activeがtrueでstartTimeが提供された場合、経過時間を表示', () => {
      const startTime = '2024-01-01T10:00:00Z' // 30分前に開始
      render(<StatsCard users={mockUsers} active={true} startTime={startTime} />)
      
      expect(screen.getByText('30分')).toBeInTheDocument()
    })

    test('activeがtrueでstartTimeが1時間15分前の場合、時間分で表示', () => {
      const startTime = '2024-01-01T09:15:00Z' // 1時間15分前
      render(<StatsCard users={mockUsers} active={true} startTime={startTime} />)
      
      expect(screen.getByText('1時間15分')).toBeInTheDocument()
    })

    test('activeがfalseの場合、停止中と表示', () => {
      render(<StatsCard users={mockUsers} active={false} />)
      
      // 監視時間セクションとステータスインジケーター両方に停止中が表示される
      const stopElements = screen.getAllByText('停止中')
      expect(stopElements.length).toBe(2)
    })

    test('startTimeが未提供の場合、停止中と表示', () => {
      render(<StatsCard users={mockUsers} active={true} />)
      
      // 監視時間セクションに停止中が表示される（ステータスは監視中）
      expect(screen.getByText('監視中')).toBeInTheDocument()
      const stopElements = screen.getAllByText('停止中')
      expect(stopElements.length).toBe(1)
    })
  })

  describe('最新コメント時間表示', () => {
    test('最も新しいコメント時間からの経過時間を表示', () => {
      // User2の方が新しい(10:10) > User1(10:05)
      render(<StatsCard users={mockUsers} active={true} />)
      
      // 現在時刻10:30から10:10のコメントまで20分前
      expect(screen.getByText('20分前')).toBeInTheDocument()
    })

    test('コメントがない場合、なしと表示', () => {
      const usersWithNoComments: User[] = [
        {
          channelId: 'channel1',
          displayName: 'User1',
          joinedAt: '2024-01-01T10:00:00Z'
        }
      ]
      
      render(<StatsCard users={usersWithNoComments} active={true} />)
      
      expect(screen.getByText('なし')).toBeInTheDocument()
    })

    test('1分未満の場合、1分未満前と表示', () => {
      const recentUsers: User[] = [
        {
          channelId: 'channel1',
          displayName: 'User1',
          joinedAt: '2024-01-01T10:29:30Z',
          firstCommentedAt: '2024-01-01T10:29:30Z', // 30秒前
          commentCount: 1
        }
      ]
      
      render(<StatsCard users={recentUsers} active={true} />)
      
      expect(screen.getByText('1分未満前')).toBeInTheDocument()
    })
  })

  describe('ステータスインジケーター', () => {
    test('activeがtrueの場合、監視中と表示', () => {
      render(<StatsCard users={mockUsers} active={true} />)
      
      expect(screen.getByText('監視中')).toBeInTheDocument()
    })

    test('activeがfalseの場合、停止中と表示', () => {
      render(<StatsCard users={mockUsers} active={false} />)
      
      // 監視時間とステータスインジケーター両方に停止中が表示される
      const stopElements = screen.getAllByText('停止中')
      expect(stopElements.length).toBe(2)
    })
  })

  describe('エッジケース', () => {
    test('firstCommentedAtが空文字列の場合、コメントなしとして扱う', () => {
      const usersWithEmptyComment: User[] = [
        {
          channelId: 'channel1',
          displayName: 'User1',
          joinedAt: '2024-01-01T10:00:00Z',
          firstCommentedAt: '', // 空文字列
          commentCount: 0
        }
      ]
      
      render(<StatsCard users={usersWithEmptyComment} active={true} />)
      
      expect(screen.getByText('なし')).toBeInTheDocument()
    })

    test('不正なfirstCommentedAtの場合、エラーにならずなしと表示', () => {
      const usersWithInvalidDate: User[] = [
        {
          channelId: 'channel1',
          displayName: 'User1',
          joinedAt: '2024-01-01T10:00:00Z',
          firstCommentedAt: 'invalid-date',
          commentCount: 1
        }
      ]
      
      render(<StatsCard users={usersWithInvalidDate} active={true} />)
      
      expect(screen.getByText('なし')).toBeInTheDocument()
    })
  })
})