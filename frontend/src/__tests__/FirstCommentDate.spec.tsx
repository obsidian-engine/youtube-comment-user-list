import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import App from '../App.jsx'
import { __mock } from '../mocks/handlers'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'

describe('初回コメント日時保持・表示機能', () => {
  beforeEach(() => {
    // 初期状態をリセット
    __mock.state = 'WAITING'
    __mock.users = []
  })

  test('テーブルヘッダーに「初回コメント」列が表示される', async () => {
    render(<App />)
    expect(screen.getByText('初回コメント')).toBeInTheDocument()
  })

  test('初回コメント日時がない場合は「--:--」と表示される', async () => {
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

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('--:--')
  })

  test('初回コメント日時が設定されている場合は正しく表示される', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: '2024-01-01T12:05:30Z'
      },
      {
        channelId: 'UC2',
        displayName: 'TestUser2',
        joinedAt: '2024-01-01T12:30:00Z',
        firstCommentedAt: '2024-01-01T12:35:15Z'
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

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('21:05')
    expect(screen.getByTestId('first-comment-1')).toHaveTextContent('21:35')
  })

  test('初回コメント日時の時刻フォーマットが正しい', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T09:05:00Z',
        firstCommentedAt: '2024-01-01T09:08:45Z'
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

    // 時刻が18:08として表示されることを確認（日本時間フォーマット）
    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('18:08')
  })

  test('firstCommentedAtが空文字列の場合は「--:--」が表示される', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      {
        channelId: 'UC1',
        displayName: 'TestUser1',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: ''  // 空文字列
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

    expect(await screen.findByTestId('first-comment-0')).toHaveTextContent('--:--')
  })
})