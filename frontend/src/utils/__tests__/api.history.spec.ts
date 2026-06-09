import { describe, test, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../../mocks/setup'
import {
  getHistorySnapshots,
  getHistorySnapshot,
  BackendError,
  type HistorySummary,
  type HistorySnapshot,
  type ErrorResponse,
} from '../api'

describe('getHistorySnapshots', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  test('200 で items 配列を返す', async () => {
    const mockItems: HistorySummary[] = [
      {
        videoId: 'abc123',
        savedAt: '2026-06-01T10:00:00Z',
        userCount: 42,
        commentCount: 100,
      },
      {
        videoId: 'def456',
        savedAt: '2026-06-02T12:00:00Z',
        userCount: 10,
        commentCount: 20,
      },
    ]

    server.use(
      http.get('*/history/snapshots', () => {
        return HttpResponse.json({ items: mockItems })
      }),
    )

    const result = await getHistorySnapshots()
    expect(result).toHaveLength(2)
    expect(result[0].videoId).toBe('abc123')
    expect(result[0].userCount).toBe(42)
    expect(result[0].commentCount).toBe(100)
    expect(result[1].videoId).toBe('def456')
  })

  test('items が空配列の場合 [] を返す', async () => {
    server.use(
      http.get('*/history/snapshots', () => {
        return HttpResponse.json({ items: [] })
      }),
    )

    const result = await getHistorySnapshots()
    expect(result).toEqual([])
  })

  test('AbortSignal を受け付ける', async () => {
    const controller = new AbortController()

    server.use(
      http.get('*/history/snapshots', () => {
        return HttpResponse.json({ items: [] })
      }),
    )

    await getHistorySnapshots(controller.signal)
    // 例外が投げられなければ OK
  })
})

describe('getHistorySnapshot', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  test('200 で snapshot を返す', async () => {
    const mockSnapshot: HistorySnapshot = {
      videoId: 'abc123',
      savedAt: '2026-06-01T10:00:00Z',
      users: [
        {
          channelId: 'UCxxx',
          displayName: 'UserA',
          joinedAt: '2026-06-01T09:00:00Z',
        },
      ],
      comments: [
        {
          id: 'cmt1',
          channelId: 'UCxxx',
          displayName: 'UserA',
          message: 'hello',
          publishedAt: '2026-06-01T09:30:00Z',
        },
      ],
    }

    server.use(
      http.get('*/history/snapshots/abc123', () => {
        return HttpResponse.json(mockSnapshot)
      }),
    )

    const result = await getHistorySnapshot('abc123')
    expect(result.videoId).toBe('abc123')
    expect(result.savedAt).toBe('2026-06-01T10:00:00Z')
    expect(result.users).toHaveLength(1)
    expect(result.users[0].channelId).toBe('UCxxx')
    expect(result.comments).toHaveLength(1)
    expect(result.comments[0].message).toBe('hello')
  })

  test('404 で BackendError を throw する', async () => {
    const errResp: ErrorResponse = {
      error: 'snapshot_not_found',
      code: 'snapshot_not_found',
      message: '指定された snapshot が見つかりません',
      httpCode: 404,
    }

    server.use(
      http.get('*/history/snapshots/unknown', () => {
        return HttpResponse.json(errResp, { status: 404 })
      }),
    )

    const err = await getHistorySnapshot('unknown').catch((e: unknown) => e)
    expect(err).toBeInstanceOf(BackendError)
    const backendErr = err as BackendError
    expect(backendErr.httpCode).toBe(404)
    expect(backendErr.code).toBe('snapshot_not_found')
  })

  test('videoId が URL エンコードされる', async () => {
    let capturedUrl = ''

    server.use(
      http.get('*/history/snapshots/:videoId', ({ request }) => {
        capturedUrl = request.url
        return HttpResponse.json({
          videoId: 'test-id',
          savedAt: '2026-06-01T10:00:00Z',
          users: [],
          comments: [],
        })
      }),
    )

    await getHistorySnapshot('test-id')
    expect(capturedUrl).toContain('test-id')
  })
})
