import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App.jsx'

describe('App Integration (MSW)', () => {
  test('切替成功で ACTIVE 表示になり、pull で人数が増える', async () => {
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
})
