import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App.jsx'
import type { User } from '../utils/api'

describe('App Integration (MSW)', () => {
  test('切替成功で監視中表示になり、pull で人数が増える', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async ({ request }) => {
        try {
          const body = (await request.json()) as { videoId?: string }
          if (!body?.videoId) return new HttpResponse('bad request', { status: 400 })
        } catch {
          return new HttpResponse('bad request', { status: 400 })
        }
        currentState = 'ACTIVE'
        users = []
        return new HttpResponse(null, { status: 200 })
      }),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `User-${users.length + 1}`,
          joinedAt: new Date().toISOString()
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期は 停止中 / 0 人
    const stopEls = await screen.findAllByText('停止中')
    expect(stopEls[0]).toBeInTheDocument()
    // ユーザー数をテキストベースで取得 (0人の表示)
    expect(screen.getByText('0')).toBeInTheDocument()

    // videoId 未入力でエラー
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    expect(await screen.findByRole('alert')).toBeInTheDocument()

    // 入力して切替
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('監視中')
      expect(activeEls[0]).toBeInTheDocument()
    })

    // 今すぐ取得 → 人数 1
    fireEvent.click(screen.getByRole('button', { name: '今すぐ取得' }))
    await waitFor(() => expect(screen.getByText('1')).toBeInTheDocument())
  })

  test('初回コメント時間が正しく表示される', async () => {
    const mockDate = new Date('2024-01-01T10:30:00Z')
    // eslint-disable-next-line prefer-const
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `TestUser-${users.length + 1}`,
          joinedAt: mockDate.toISOString(),
          commentCount: 1,
          firstCommentedAt: mockDate.toISOString()
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期状態で初回コメントヘッダーが表示されている
    expect(screen.getByText('初回コメント')).toBeInTheDocument()

    // ユーザーがいない時は「ユーザーがいません。」が表示される
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()

    // 今すぐ取得でユーザーを追加
    fireEvent.click(screen.getByRole('button', { name: '今すぐ取得' }))

    // 初回コメント時間が正しく表示される
    await waitFor(() => {
      expect(screen.getByText('TestUser-1')).toBeInTheDocument()
      // 環境に依存しないように時刻パターンをチェック (HH:mm形式)
      expect(screen.getByText(/^\d{2}:\d{2}$/)).toBeInTheDocument()
    })
  })

})
