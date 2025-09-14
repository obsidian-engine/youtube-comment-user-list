import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../mocks/setup'
import App from '../App.jsx'

describe('App Integration', () => {
  test('切替成功で ACTIVE 表示になり、pull で人数が増える', async () => {
    // テスト専用の状態を持つ
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    let users: string[] = []

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
        users = [...users, `User-${users.length + 1}`]
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期は WAITING / 0 人
    expect(await screen.findByText('WAITING')).toBeInTheDocument()
    expect(screen.getByText(/人数: 0/)).toBeInTheDocument()

    // 切替
    const input = screen.getByPlaceholderText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID123' } })
    fireEvent.click(screen.getByText('切替'))

    await waitFor(() => expect(screen.getByText('ACTIVE')).toBeInTheDocument())

    // 今すぐ取得 → 人数 1
    fireEvent.click(screen.getByText('今すぐ取得'))
    await waitFor(() => expect(screen.getByText(/人数: 1/)).toBeInTheDocument())
  })
})
