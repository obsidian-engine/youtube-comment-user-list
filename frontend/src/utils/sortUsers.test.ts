import { render, waitFor, fireEvent, screen } from '@testing-library/react'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App'
import { createElement } from 'react'

// TDD: 仕様
// 1) 初回コメント時間 (firstCommentedAt) 昇順
// 2) 初回コメント時間が同一なら channelId 昇順（表示は従来どおり displayName）
// 3) firstCommentedAtがない場合は末尾に寄せる

function namesInTable() {
  const rows = document.querySelectorAll('tbody tr')
  return Array.from(rows).map((tr) => {
    const nameCell = tr.querySelectorAll('td')[1]
    return nameCell?.textContent?.trim() || ''
  })
}

describe('ユーザー並び順', () => {
  test('初回コメント時間→チャンネルIDの優先で安定ソートされる', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      { channelId: 'UC3', displayName: 'Charlie', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T14:00:00Z' },
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC2', displayName: 'Bob', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC6', displayName: 'Frank', joinedAt: '2024-01-01T12:00:00Z' }, // firstCommentedAtなし
      { channelId: 'UC4', displayName: 'David', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T15:00:00Z' },
      { channelId: 'UC5', displayName: 'Eve', joinedAt: '2024-01-01T12:00:00Z' } // firstCommentedAtなし
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(createElement(App))

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    
    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })
    
    // テーブルが描画され、6名が表示されるまで待機
    await waitFor(() => expect(namesInTable().length).toBe(6))

    // 期待順:
    // 1. Alice (UC1, 13:00)
    // 2. Bob (UC2, 13:00)  ← channelId順
    // 3. Charlie (UC3, 14:00)
    // 4. David (UC4, 15:00)
    // 5. Eve (UC5, firstCommentedAtなし)
    // 6. Frank (UC6, firstCommentedAtなし)
    expect(namesInTable()).toEqual(['Alice', 'Bob', 'Charlie', 'David', 'Eve', 'Frank'])
  })

  test('すべてのユーザーがfirstCommentedAtなしの場合、channelId順でソート', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      { channelId: 'UC3', displayName: 'Charlie', joinedAt: '2024-01-01T12:00:00Z' },
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T12:00:00Z' },
      { channelId: 'UC2', displayName: 'Bob', joinedAt: '2024-01-01T12:00:00Z' }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(createElement(App))

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))

    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    await waitFor(() => expect(namesInTable().length).toBe(3))
    expect(namesInTable()).toEqual(['Alice', 'Bob', 'Charlie'])
  })

  test('すべてのユーザーが同じfirstCommentedAtの場合、channelId順でソート', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      { channelId: 'UC3', displayName: 'Charlie', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC2', displayName: 'Bob', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(createElement(App))

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))

    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    await waitFor(() => expect(namesInTable().length).toBe(3))
    expect(namesInTable()).toEqual(['Alice', 'Bob', 'Charlie'])
  })

  test('channelIdが同じ場合、displayName順でソート', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      { channelId: 'UC1', displayName: 'Zoe', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC1', displayName: 'Bob', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(createElement(App))

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))

    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    await waitFor(() => expect(namesInTable().length).toBe(3))
    expect(namesInTable()).toEqual(['Alice', 'Bob', 'Zoe'])
  })

  test('無効なfirstCommentedAt（不正な日付文字列）は末尾に寄せる', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users = [
      { channelId: 'UC3', displayName: 'Charlie', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: 'invalid-date' },
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: '2024-01-01T13:00:00Z' },
      { channelId: 'UC2', displayName: 'Bob', joinedAt: '2024-01-01T12:00:00Z', firstCommentedAt: 'another-invalid' }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: currentState, count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', async () => {
        currentState = 'ACTIVE'
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(createElement(App))

    // 動画切り替えでACTIVE状態にする
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'TEST123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))

    // ACTIVE状態になるまで待機
    await waitFor(async () => {
      const activeEls = await screen.findAllByText('ACTIVE')
      expect(activeEls[0]).toBeInTheDocument()
    })

    await waitFor(() => expect(namesInTable().length).toBe(3))
    // AliceのみfirstCommentedAtが有効、BobとCharlieは無効なのでchannelId順
    expect(namesInTable()).toEqual(['Alice', 'Bob', 'Charlie'])
  })
})