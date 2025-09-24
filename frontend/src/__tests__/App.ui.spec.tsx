import { render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import App from '../App'

// Mock API functions
vi.mock('../utils/api', () => ({
  getStatus: vi.fn().mockResolvedValue({ count: 0, active: false }),
  getUsers: vi.fn().mockResolvedValue([]),
  postSwitchVideo: vi.fn().mockResolvedValue({}),
  postPull: vi.fn().mockResolvedValue({}),
  postReset: vi.fn().mockResolvedValue({})
}))

describe('App UI', () => {
  test('ヘッダー/カウンタ/テーブルが描画される', async () => {
    render(<App />)
    
    // StatsCard内の要素を確認
    expect(await screen.findByText('総ユーザー数')).toBeInTheDocument()
    expect(screen.getByText('監視開始時間')).toBeInTheDocument()
    expect(screen.getByText('画面最終更新')).toBeInTheDocument()
    
    // UserTable内の要素を確認
    expect(screen.getByText('名前')).toBeInTheDocument()
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
  })
})
