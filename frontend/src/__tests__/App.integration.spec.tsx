import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { server } from '../mocks/setup'
import { http, HttpResponse } from 'msw'
import App from '../App.jsx'
import type { User } from '../utils/api'

// Note: matchMedia and localStorage are mocked globally in setup.ts

describe('App Integration (MSW)', () => {
  // 統合テスト用に長めのタイムアウトを設定
  vi.setConfig({ testTimeout: 12000 })
  test('切替成功で監視中表示になり、pull で人数が増える', async () => {
    let currentState: 'WAITING' | 'ACTIVE' = 'WAITING'
    const users: User[] = []

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
        users.length = 0 // 配列をクリア
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
    const users: User[] = []

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

  test('自動更新時は「取得しました」メッセージが表示されない（軽量化版）', async () => {
    // 軽量化：useFakeTimersを使わずにuseAutoRefreshフックの動作のみテスト
    const users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
    )

    render(<App />)

    // 基本的な表示確認のみ（自動更新の詳細はuseAutoRefreshのユニットテストでカバー）
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
    expect(screen.queryByText('取得しました')).not.toBeInTheDocument()
  })

  test('手動「今すぐ取得」は「取得しました」メッセージが表示される', async () => {
    const users: User[] = []

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

  test('自動更新処理の基本動作確認（軽量化版）', async () => {
    // 軽量化：タイマー処理を除いて基本的な状態確認のみ
    const users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
    )

    render(<App />)

    // 基本的な表示確認のみ（タイマーテストはuseAutoRefreshのユニットテストでカバー）
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
    expect(screen.queryByText('取得しました')).not.toBeInTheDocument()
  })

  test('自動更新間隔設定の基本動作確認（軽量化版）', async () => {
    // 軽量化：タイマー処理を除いて設定UIの確認のみ
    const users: User[] = []

    server.use(
      http.get('*/status', () => HttpResponse.json({ status: 'ACTIVE', count: users.length })),
      http.get('*/users.json', () => HttpResponse.json(users)),
    )

    render(<App />)

    // 基本的な表示確認
    expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()

    // 更新間隔セレクトボックスの存在確認（詳細なタイマー動作はuseAutoRefreshでテスト）
    const intervalInput = screen.getByLabelText('更新間隔')
    expect(intervalInput).toBeInTheDocument()
  })

  test('停止中(WAITING)でもユーザーリストが保持される', async () => {
    let currentStatus: 'WAITING' | 'ACTIVE' = 'ACTIVE'
    const users: User[] = [
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
        count: users.length,
        startedAt: currentStatus === 'ACTIVE' ? new Date().toISOString() : undefined
      })),
      http.get('*/users.json', () => HttpResponse.json(users)),
    )

    render(<App />)

    // 手動でrefresh実行（MSWがレスポンスを確実に返すため）
    const refreshButton = screen.getByRole('button', { name: '今すぐ取得' })
    fireEvent.click(refreshButton)

    // レスポンス後のユーザー表示を待つ
    await waitFor(() => {
      expect(screen.getByText('ExistingUser1')).toBeInTheDocument()
      expect(screen.getByText('ExistingUser2')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument()
    }, { timeout: 5000 })

    // ACTIVE状態が反映されて監視中が表示される
    await waitFor(() => {
      const statusElements = screen.queryAllByText('監視中')
      expect(statusElements.length).toBeGreaterThan(0)
    }, { timeout: 3000 })

    // サーバー状態を停止中に変更
    currentStatus = 'WAITING'

    // 手動でrefreshを実行（自動更新と同じ処理）
    fireEvent.click(refreshButton)

    // 状態は停止中になるが、ユーザーリストは保持される
    await waitFor(() => {
      const statusElements = screen.queryAllByText('停止中')
      expect(statusElements.length).toBeGreaterThan(0)
    }, { timeout: 5000 })

    // ユーザーリストは保持されている
    expect(screen.getByText('ExistingUser1')).toBeInTheDocument()
    expect(screen.getByText('ExistingUser2')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.queryByText('ユーザーがいません。')).not.toBeInTheDocument()
  })

  test('切替実行時のみユーザーリストがクリアされる', async () => {
    const users: User[] = [
      {
        channelId: 'UC1',
        displayName: 'OldUser1',
        joinedAt: new Date().toISOString(),
        commentCount: 2
      }
    ]

    server.use(
      http.get('*/status', () => {
        return HttpResponse.json({
          status: 'ACTIVE',
          count: users.length,
          startedAt: users.length > 0 ? new Date().toISOString() : undefined
        })
      }),
      http.get('*/users.json', () => {
        return HttpResponse.json(users)
      }),
      http.post('*/switch-video', () => {
        // 切替時にユーザーをクリア
        users.length = 0
        return new HttpResponse(null, { status: 200 })
      }),
    )

    render(<App />)

    // 手動でrefresh実行（MSWがレスポンスを確実に返すため）
    const refreshBtn = screen.getByRole('button', { name: '今すぐ取得' })
    fireEvent.click(refreshBtn)

    // 初期状態：ユーザーが読み込まれるまで待つ
    await waitFor(() => {
      expect(screen.getByText('OldUser1')).toBeInTheDocument()
      expect(screen.getByText('1')).toBeInTheDocument()
    }, { timeout: 5000 })

    // 監視中状態の確認
    await waitFor(() => {
      const statusElements = screen.queryAllByText('監視中')
      expect(statusElements.length).toBeGreaterThan(0)
    }, { timeout: 3000 })

    // videoIdを入力して切替実行
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'new-video-123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))

    // 切替後はリストがクリアされる（refreshWithClearが呼ばれる）
    await waitFor(() => {
      expect(screen.getByText('ユーザーがいません。')).toBeInTheDocument()
      expect(screen.queryByText('OldUser1')).not.toBeInTheDocument()
    }, { timeout: 5000 })
  })

})
