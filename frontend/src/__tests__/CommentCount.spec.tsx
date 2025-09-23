import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import App from '../App.jsx'
import { __mock } from '../mocks/handlers'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'

describe('発言数表示機能', () => {
  beforeEach(() => {
    // 初期状態をリセット
    __mock.state = 'WAITING'
    __mock.users = []
  })

  test('テーブルヘッダーに「発言数」列が表示される', async () => {
    render(<App />)
    expect(screen.getByText('発言数')).toBeInTheDocument()
  })

  test('発言数がない場合は「0」と表示される', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z'
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    
    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    // 発言数セルを検索（テストIDを使用）
    expect(await screen.findByTestId('comment-count-0')).toHaveTextContent('0')
  })

  test('発言数が設定されている場合は正しく表示される', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 5
      },
      {
        channelId: 'UC2',
        displayName: 'TestUser2',
        joinedAt: '2024-01-01T12:30:00Z',
        commentCount: 12
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    
    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    expect(await screen.findByTestId('comment-count-0')).toHaveTextContent('5')
    expect(screen.getByTestId('comment-count-1')).toHaveTextContent('12')
  })

  test('発言数が大きい数値でも正しく表示される', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'ActiveUser',
        joinedAt: '2024-01-01T12:00:00Z',
        commentCount: 999
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    
    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    expect(await screen.findByTestId('comment-count-0')).toHaveTextContent('999')
  })
})