import { vi, describe, test, expect, beforeEach, afterEach } from 'vitest'

// 環境変数をテスト用に設定
vi.stubGlobal('import.meta', { env: { VITE_BACKEND_URL: 'http://localhost:8080' } })

import { 
  getStatus, 
  getUsers, 
  postSwitchVideo, 
  postPull, 
  postReset,
  type StatusResponse,
  type User
} from './api'

// fetchのモック
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

describe('API functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getStatus', () => {
    test('正常なレスポンスを返す', async () => {
      const mockResponse: StatusResponse = {
        status: 'ACTIVE',
        count: 5,
        videoId: 'test-video',
        startedAt: '2024-01-01T09:00:00Z'
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse)
      })

      const result = await getStatus()

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/status', { signal: undefined })
      expect(result).toEqual(mockResponse)
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()
      const mockResponse: StatusResponse = { status: 'WAITING', count: 0 }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse)
      })

      await getStatus(controller.signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/status', { signal: controller.signal })
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500
      })

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

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockUsers)
      })

      const result = await getUsers()

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/users.json', { signal: undefined })
      expect(result).toEqual(mockUsers)
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      await getUsers(controller.signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/users.json', { signal: controller.signal })
    })
  })

  describe('postSwitchVideo', () => {
    test('正常にビデオ切替リクエストを送信', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await postSwitchVideo('test-video')

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/switch-video', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ videoId: 'test-video' }),
        signal: undefined
      })
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await postSwitchVideo('test-video', controller.signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/switch-video', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ videoId: 'test-video' }),
        signal: controller.signal
      })
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400
      })

      await expect(postSwitchVideo('test-video')).rejects.toThrow('HTTP 400')
    })
  })

  describe('postPull', () => {
    test('正常にプルリクエストを送信', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await postPull()

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/pull', { method: 'POST', signal: undefined })
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await postPull(controller.signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/pull', { method: 'POST', signal: controller.signal })
    })
  })

  describe('postReset', () => {
    test('正常にリセットリクエストを送信', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await postReset()

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/reset', { method: 'POST', signal: undefined })
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await postReset(controller.signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/reset', { method: 'POST', signal: controller.signal })
    })
  })

  describe('network error handling', () => {
    test('ネットワークエラーの場合例外を投げる', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      await expect(getStatus()).rejects.toThrow('Network error')
    })

    test('AbortErrorは正常に伝播される', async () => {
      const abortError = new Error('The user aborted a request')
      abortError.name = 'AbortError'
      mockFetch.mockRejectedValueOnce(abortError)

      await expect(getStatus()).rejects.toThrow('The user aborted a request')
    })
  })
})