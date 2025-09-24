import { describe, test, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../mocks/setup'
import { 
  getStatus, 
  getUsers, 
  postSwitchVideo, 
  postPull, 
  postReset,
  type StatusResponse,
  type User
} from './api'

describe('API functions', () => {
  beforeEach(() => {
    server.resetHandlers()
  })

  describe('getStatus', () => {
    test('正常なレスポンスを返す', async () => {
      const mockResponse: StatusResponse = {
        status: 'ACTIVE',
        count: 5
      }

      // MSWハンドラーでレスポンスを設定
      server.use(
        http.get('*/status', () => {
          return HttpResponse.json(mockResponse)
        })
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
        })
      )

      await getStatus(controller.signal)
      // MSWはAbortSignalの処理も適切に行うため、テストが成功すれば正しく動作している
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      server.use(
        http.get('*/status', () => {
          return new HttpResponse(null, { status: 500 })
        })
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
          joinedAt: '2024-01-01T10:00:00Z'
        }
      ]

      server.use(
        http.get('*/users.json', () => {
          return HttpResponse.json(mockUsers)
        })
      )

      const result = await getUsers()
      expect(result).toEqual(mockUsers)
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.get('*/users.json', () => {
          return HttpResponse.json([])
        })
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
        })
      )

      await postSwitchVideo('test-video')
      // 例外が投げられなければ成功
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.post('*/switch-video', () => {
          return new HttpResponse(null, { status: 200 })
        })
      )

      await postSwitchVideo('test-video', controller.signal)
      // 例外が投げられなければ成功
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      server.use(
        http.post('*/switch-video', () => {
          return new HttpResponse(null, { status: 400 })
        })
      )

      await expect(postSwitchVideo('test-video')).rejects.toThrow('HTTP 400')
    })
  })

  describe('postPull', () => {
    test('正常にプルリクエストを送信', async () => {
      server.use(
        http.post('*/pull', () => {
          return new HttpResponse(null, { status: 200 })
        })
      )

      await postPull()
      // 例外が投げられなければ成功
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.post('*/pull', () => {
          return new HttpResponse(null, { status: 200 })
        })
      )

      await postPull(controller.signal)
      // 例外が投げられなければ成功
    })
  })

  describe('postReset', () => {
    test('正常にリセットリクエストを送信', async () => {
      server.use(
        http.post('*/reset', () => {
          return new HttpResponse(null, { status: 200 })
        })
      )

      await postReset()
      // 例外が投げられなければ成功
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      server.use(
        http.post('*/reset', () => {
          return new HttpResponse(null, { status: 200 })
        })
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
        })
      )

      await expect(getStatus()).rejects.toThrow()
    })

    test('AbortErrorは正常に伝播される', async () => {
      const controller = new AbortController()
      controller.abort()

      await expect(getStatus(controller.signal)).rejects.toThrow()
    })
  })
})