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

  describe('監視開始時間表示', () => {
    test('activeがtrueでstartTimeが提供された場合、開始時間を表示', () => {
      const startTime = '2024-01-01T10:00:00Z'
      render(<StatsCard users={mockUsers} active={true} startTime={startTime} />)
      
      expect(screen.getByText('2024/01/01 19:00')).toBeInTheDocument()
    })

    test('activeがfalseの場合、未開始と表示', () => {
      render(<StatsCard users={mockUsers} active={false} />)
      
      // 監視開始時間セクションに未開始が表示される
      expect(screen.getByText('未開始')).toBeInTheDocument()
      expect(screen.getByText('停止中')).toBeInTheDocument()
    })

    test('startTimeが未提供の場合、未開始と表示', () => {
      render(<StatsCard users={mockUsers} active={true} />)
      
      expect(screen.getByText('監視中')).toBeInTheDocument()
      expect(screen.getByText('未開始')).toBeInTheDocument()
    })
  })

  describe('ステータスインジケーター', () => {
    test('activeがtrueの場合、監視中と表示', () => {
      render(<StatsCard users={mockUsers} active={true} />)
      
      expect(screen.getByText('監視中')).toBeInTheDocument()
    })

    test('activeがfalseの場合、停止中と表示', () => {
      render(<StatsCard users={mockUsers} active={false} />)
      
      expect(screen.getByText('停止中')).toBeInTheDocument()
    })
  })

  describe('画面最終更新表示', () => {
    test('lastUpdatedが提供された場合、その値を表示', () => {
      render(<StatsCard users={mockUsers} active={true} lastUpdated="12:34:56" />)
      
      expect(screen.getByText('12:34:56')).toBeInTheDocument()
    })

    test('lastUpdatedが未提供の場合、デフォルト値を表示', () => {
      render(<StatsCard users={mockUsers} active={true} />)
      
      expect(screen.getByText('--:--:--')).toBeInTheDocument()
    })
  })

  describe('エッジケース', () => {
    test('不正なstartTimeの場合、エラーにならず未開始と表示', () => {
      render(<StatsCard users={mockUsers} active={true} startTime="invalid-date" />)
      
      expect(screen.getByText('未開始')).toBeInTheDocument()
    })
  })
})