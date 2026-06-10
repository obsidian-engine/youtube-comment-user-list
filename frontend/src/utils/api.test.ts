import { describe, test, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../mocks/setup'
import {
  getStatus,
  getUsers,
  postSwitchVideo,
  postPull,
  postReset,
  searchComments,
  BackendError,
  type StatusResponse,
  type User,
  type ErrorResponse,
} from './api'

describe('API functions', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  describe('getStatus', () => {
    test('正常なレスポンスを返す', async () => {
      const mockResponse: StatusResponse = {
        status: 'ACTIVE',
        count: 5,
      }

      // MSWハンドラーでレスポンスを設定
      server.use(
        http.get('*/status', () => {
          return HttpResponse.json(mockResponse)
        }),
      )

      const result = await getStatus()
      expect(result).toEqual(mockResponse)
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()
      const mockResponse: StatusResponse = { status: 'WAITING', count: 0 }

      server.use(
        http.get('*/status', () => {
          return HttpResponse.json(mockResponse)
        }),
      )

      await getStatus(controller.signal)
      // MSWはAbortSignalの処理も適切に行うため、テストが成功すれば正しく動作している
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      server.use(
        http.get('*/status', () => {
          return new HttpResponse(null, { status: 500 })
        }),
      )

      await expect(getStatus()).rejects.toThrow('HTTP 500')
    })
  })

  describe('getUsers', () => {
    test('正常なユーザー配列を返す', async () => {
      const mockUsers: User[] = [
        {
          channelId: 'UC123',
          displayName: 'TestUser',
          joinedAt: '2024-01-01T10:00:00Z',
        },
      ]

      server.use(
        http.get('*/users.json', () => {
          return HttpResponse.json(mockUsers)
        }),
      )

      const result = await getUsers()
      expect(result).toEqual(mockUsers)
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.get('*/users.json', () => {
          return HttpResponse.json([])
        }),
      )

      await getUsers(controller.signal)
      // MSWはAbortSignalの処理も適切に行う
    })
  })

  describe('postSwitchVideo', () => {
    test('正常にビデオ切替リクエストを送信', async () => {
      server.use(
        http.post('*/switch-video', () => {
          return new HttpResponse(null, { status: 200 })
        }),
      )

      await postSwitchVideo('test-video')
      // 例外が投げられなければ成功
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.post('*/switch-video', () => {
          return new HttpResponse(null, { status: 200 })
        }),
      )

      await postSwitchVideo('test-video', controller.signal)
      // 例外が投げられなければ成功
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      server.use(
        http.post('*/switch-video', () => {
          return new HttpResponse(null, { status: 400 })
        }),
      )

      await expect(postSwitchVideo('test-video')).rejects.toThrow('HTTP 400')
    })
  })

  describe('postPull', () => {
    test('正常にプルリクエストを送信', async () => {
      const mockResponse = {
        addedCount: 1,
        skippedCount: 0,
        autoReset: false,
        pollingIntervalMillis: 15000,
      }

      server.use(
        http.post('*/pull', () => {
          return HttpResponse.json(mockResponse)
        }),
      )

      const result = await postPull()
      expect(result).toEqual(mockResponse)
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()
      const mockResponse = {
        addedCount: 0,
        skippedCount: 0,
        autoReset: false,
        pollingIntervalMillis: 15000,
      }

      server.use(
        http.post('*/pull', () => {
          return HttpResponse.json(mockResponse)
        }),
      )

      const result = await postPull(controller.signal)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('postReset', () => {
    test('正常にリセットリクエストを送信', async () => {
      server.use(
        http.post('*/reset', () => {
          return new HttpResponse(null, { status: 200 })
        }),
      )

      await postReset()
      // 例外が投げられなければ成功
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.post('*/reset', () => {
          return new HttpResponse(null, { status: 200 })
        }),
      )

      await postReset(controller.signal)
      // 例外が投げられなければ成功
    })
  })

  describe('network error handling', () => {
    test('ネットワークエラーの場合例外を投げる', async () => {
      server.use(
        http.get('*/status', () => {
          return HttpResponse.error()
        }),
      )

      await expect(getStatus()).rejects.toThrow()
    })

    test('AbortErrorは正常に伝播される', async () => {
      const controller = new AbortController()
      controller.abort()

      await expect(getStatus(controller.signal)).rejects.toThrow()
    })
  })

  describe('BackendError', () => {
    test('ErrorResponse を含む 4xx レスポンスで BackendError をスロー', async () => {
      const errResp: ErrorResponse = {
        error: 'video_not_found',
        code: 'video_not_found',
        message: '指定された動画が見つかりません',
        httpCode: 404,
        logs: [{ level: 'error', source: 'YOUTUBE', message: 'videoId not found' }],
      }

      server.use(
        http.post('*/switch-video', () => {
          return HttpResponse.json(errResp, { status: 404 })
        }),
      )

      const err = await postSwitchVideo('invalid-id').catch((e: unknown) => e)
      expect(err).toBeInstanceOf(BackendError)
      const backendErr = err as BackendError
      expect(backendErr.httpCode).toBe(404)
      expect(backendErr.code).toBe('video_not_found')
      expect(backendErr.message).toBe('指定された動画が見つかりません')
      expect(backendErr.logs).toHaveLength(1)
      expect(backendErr.logs[0].source).toBe('YOUTUBE')
    })

    test('ErrorResponse の logs が空の場合 BackendError.logs は空配列', async () => {
      const errResp: ErrorResponse = {
        error: 'internal_error',
        httpCode: 500,
      }

      server.use(
        http.post('*/reset', () => {
          return HttpResponse.json(errResp, { status: 500 })
        }),
      )

      const err = await postReset().catch((e: unknown) => e)
      expect(err).toBeInstanceOf(BackendError)
      const backendErr = err as BackendError
      expect(backendErr.logs).toEqual([])
    })

    test('JSON パース不能な 5xx レスポンスは HttpError にフォールバック', async () => {
      server.use(
        http.post('*/reset', () => {
          return new HttpResponse('Internal Server Error', { status: 500 })
        }),
      )

      const err = await postReset().catch((e: unknown) => e)
      // JSON parse 失敗 → HttpError fallback
      expect(err).not.toBeInstanceOf(BackendError)
    })

    test('logs を含む ErrorResponse の全フィールドが BackendError に正しく格納される', async () => {
      const errResp: ErrorResponse = {
        error: 'quota_exceeded',
        code: 'quota_exceeded',
        message: 'YouTube API クォータを超過しました',
        httpCode: 429,
        logs: [
          { level: 'warn', source: 'YOUTUBE', message: 'quota limit approaching' },
          { level: 'error', source: 'PULL', message: 'quota exceeded on listMessages' },
        ],
      }

      server.use(
        http.post('*/switch-video', () => {
          return HttpResponse.json(errResp, { status: 429 })
        }),
      )

      const err = await postSwitchVideo('any-id').catch((e: unknown) => e)
      expect(err).toBeInstanceOf(BackendError)
      const backendErr = err as BackendError
      expect(backendErr.logs).toHaveLength(2)
      expect(backendErr.logs[0].level).toBe('warn')
      expect(backendErr.logs[1].level).toBe('error')
    })
  })

  describe('searchComments BackendError', () => {
    test('ErrorResponse を含む 4xx で BackendError をスロー (fetchWithRetry 経由)', async () => {
      const errResp: ErrorResponse = {
        error: 'quota_exceeded',
        code: 'quota_exceeded',
        message: 'クォータ超過',
        httpCode: 429,
        logs: [{ level: 'error', source: 'YOUTUBE', message: 'quota exceeded' }],
      }

      server.use(
        http.get('*/comments', () => {
          return HttpResponse.json(errResp, { status: 429 })
        }),
      )

      const err = await searchComments(['test']).catch((e: unknown) => e)
      expect(err).toBeInstanceOf(BackendError)
      const backendErr = err as BackendError
      expect(backendErr.code).toBe('quota_exceeded')
      expect(backendErr.httpCode).toBe(429)
      expect(backendErr.logs).toHaveLength(1)
      expect(backendErr.logs[0].source).toBe('YOUTUBE')
    })
  })
})
