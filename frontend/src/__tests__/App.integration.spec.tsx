import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App.jsx'

type User = {
  channelId: string
  displayName: string
  joinedAt: string
}

describe('App Integration (MSW)', () => {
  test('切替成功で ACTIVE 表示になり、pull で人数が増える', async () => {
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

    // 初期は WAITING / 0 人
    const waitingEls = await screen.findAllByText('WAITING')
    expect(waitingEls[0]).toBeInTheDocument()
    expect(screen.getByTestId('counter')).toHaveTextContent('0')

    // videoId 未入力でエラー
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    expect(await screen.findByRole('alert')).toBeInTheDocument()

    // 入力して切替
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    // 今すぐ取得 → 人数 1
    fireEvent.click(screen.getByRole('button', { name: '今すぐ取得' }))
    await waitFor(() => expect(screen.getByTestId('counter')).toHaveTextContent('1'))
  })

  test('参加時間が正しく表示される', async () => {
    const mockDate = new Date('2024-01-01T10:30:00Z')
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `TestUser-${users.length + 1}`,
          joinedAt: mockDate.toISOString()
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期状態で参加時間ヘッダーが表示されている
    expect(screen.getByText('参加時間')).toBeInTheDocument()

    // ユーザーがいない時は「ユーザーがいません。」が表示される
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()

    // 今すぐ取得でユーザーを追加
    fireEvent.click(screen.getByRole('button', { name: '今すぐ取得' }))

    // 参加時間が正しく表示される
    await waitFor(() => {
      expect(screen.getByText('TestUser-1')).toBeInTheDocument()
      // 日本時間で表示されているかチェック (10:30 UTC → 19:30 JST)
      expect(screen.getByText('19:30')).toBeInTheDocument()
    })
  })

  test('複数ユーザーの参加時間表示', async () => {
    const user1Date = new Date('2024-01-01T09:00:00Z')
    const user2Date = new Date('2024-01-01T09:15:00Z')
    let users: User[] = [
      {
        channelId: 'UC1',
        displayName: 'FirstUser',
        joinedAt: user1Date.toISOString()
      },
      {
        channelId: 'UC2',
        displayName: 'SecondUser',
        joinedAt: user2Date.toISOString()
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
    )

    render(<App />)

    // 両方のユーザーと参加時間が表示される
    await waitFor(() => {
      expect(screen.getByText('FirstUser')).toBeInTheDocument()
      expect(screen.getByText('SecondUser')).toBeInTheDocument()
      // JST時間で表示される
      expect(screen.getByText('18:00')).toBeInTheDocument() // 09:00 UTC → 18:00 JST
      expect(screen.getByText('18:15')).toBeInTheDocument() // 09:15 UTC → 18:15 JST
    })

    // 参加者数が正しく表示される
    expect(screen.getByTestId('counter')).toHaveTextContent('2')
  })
})
