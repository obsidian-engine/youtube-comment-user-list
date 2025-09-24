import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { vi } from 'vitest'
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

  test('自動更新時は「取得しました」メッセージが表示されない', async () => {
    vi.useFakeTimers()
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `AutoUser-${users.length + 1}`,
          joinedAt: new Date().toISOString(),
          commentCount: 1
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期状態：ユーザー0人、メッセージなし
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
    expect(screen.queryByText('取得しました')).not.toBeInTheDocument()

    // 15秒経過させて自動更新をトリガー
    act(() => {
      vi.advanceTimersByTime(15000)
    })

    // ユーザーが追加されるが「取得しました」メッセージは表示されない
    await waitFor(() => {
      expect(screen.getByText('AutoUser-1')).toBeInTheDocument()
    })
    
    // 自動更新時は成功メッセージが表示されないことを確認
    expect(screen.queryByText('取得しました')).not.toBeInTheDocument()

    vi.useRealTimers()
  })

  test('手動「今すぐ取得」は「取得しました」メッセージが表示される', async () => {
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `ManualUser-${users.length + 1}`,
          joinedAt: new Date().toISOString(),
          commentCount: 1
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期状態
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()

    // 手動で「今すぐ取得」ボタンをクリック
    fireEvent.click(screen.getByRole('button', { name: '今すぐ取得' }))

    // ユーザーが追加される
    await waitFor(() => {
      expect(screen.getByText('ManualUser-1')).toBeInTheDocument()
    })

    // 手動取得時は成功メッセージが表示される
    await waitFor(() => {
      expect(screen.getByText('取得しました')).toBeInTheDocument()
    })
  })

  test('自動更新がonPullSilent処理を使用して新しいユーザーを取得する', async () => {
    vi.useFakeTimers()
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `AutoUser-${users.length + 1}`,
          joinedAt: new Date().toISOString(),
          commentCount: 1
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期状態：ユーザー0人
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()

    // 更新間隔を15秒に設定（デフォルト）
    const intervalSelect = screen.getByLabelText('更新間隔') as HTMLSelectElement
    expect(intervalSelect.value).toBe('15')

    // 15秒経過させて自動更新をトリガー
    act(() => {
      vi.advanceTimersByTime(15000)
    })

    // onPull処理によってユーザーが追加されることを確認
    await waitFor(() => {
      expect(screen.getByText('AutoUser-1')).toBeInTheDocument()
      expect(screen.getByText('1')).toBeInTheDocument() // ユーザー数表示
    })

    // さらに15秒経過で2人目追加
    act(() => {
      vi.advanceTimersByTime(15000)
    })

    await waitFor(() => {
      expect(screen.getByText('AutoUser-2')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument() // ユーザー数表示
    })

    vi.useRealTimers()
  })

  test('自動更新間隔を0に設定すると自動更新が停止する', async () => {
    vi.useFakeTimers()
    let users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/pull', () => {
        users.push({
          channelId: `UC${users.length + 1}`,
          displayName: `AutoUser-${users.length + 1}`,
          joinedAt: new Date().toISOString()
        })
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 更新間隔を0（停止）に設定
    const intervalSelect = screen.getByLabelText('更新間隔') as HTMLSelectElement
    fireEvent.change(intervalSelect, { target: { value: '0' } })

    // 15秒経過してもユーザーは追加されない
    act(() => {
      vi.advanceTimersByTime(15000)
    })

    await waitFor(() => {
      expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
    })

    vi.useRealTimers()
  })

  test('停止中(WAITING)でもユーザーリストが保持される', async () => {
    let currentStatus: 'WAITING' | 'ACTIVE' = 'ACTIVE'
    let users: User[] = [
      {
        channelId: 'UC1',
        displayName: 'ExistingUser1',
        joinedAt: new Date().toISOString(),
        commentCount: 5
      },
      {
        channelId: 'UC2', 
        displayName: 'ExistingUser2',
        joinedAt: new Date().toISOString(),
        commentCount: 3
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ 
        status: currentStatus, 
        count: users.length 
      })),
      http.get('*/users.json', () => HttpResponse.json(users)),
    )

    render(<App />)

    // 初期状態は停止中、その後自動更新でサーバー状態を取得し監視中になる
    // まず初期レンダリングを待つ
    expect(screen.getByText('停止中')).toBeInTheDocument()
    
    // 自動更新でサーバー状態（ACTIVE + ユーザー）を取得後、監視中になる
    await waitFor(() => {
      expect(screen.getAllByText('監視中')[0]).toBeInTheDocument()
      expect(screen.getByText('ExistingUser1')).toBeInTheDocument()
      expect(screen.getByText('ExistingUser2')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument()
    }, { timeout: 20000 })

    // サーバー状態を停止中に変更
    currentStatus = 'WAITING'

    // 手動でrefreshを実行（自動更新と同じ処理）
    const refreshButton = screen.getByRole('button', { name: '今すぐ取得' })
    fireEvent.click(refreshButton)

    // 状態は停止中になるが、ユーザーリストは保持される
    await waitFor(() => {
      expect(screen.getAllByText('停止中')[0]).toBeInTheDocument()
    })

    // ユーザーリストは保持されている
    expect(screen.getByText('ExistingUser1')).toBeInTheDocument()
    expect(screen.getByText('ExistingUser2')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.queryByText('ユーザーがいません。')).not.toBeInTheDocument()
  })

  test('切替実行時のみユーザーリストがクリアされる', async () => {
    let users: User[] = [
      {
        channelId: 'UC1',
        displayName: 'OldUser1',
        joinedAt: new Date().toISOString(),
        commentCount: 2
      }
    ]

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
      http.post('*/switch-video', () => {
        // 切替時にユーザーをクリア
        users = []
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 初期状態：ユーザーが存在
    await waitFor(() => {
      expect(screen.getByText('OldUser1')).toBeInTheDocument()
      expect(screen.getByText('1')).toBeInTheDocument()
    })

    // videoIdを入力して切替実行
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'new-video-123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))

    // 切替後はリストがクリアされる
    await waitFor(() => {
      expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
      expect(screen.queryByText('OldUser1')).not.toBeInTheDocument()
    })
  })

})
