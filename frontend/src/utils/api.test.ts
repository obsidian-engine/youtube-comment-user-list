import { describe, test, expect, beforeEach, vi, afterEach } from 'vitest'
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
global.fetch = mockFetch

describe('API functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // 環境変数のモック
    vi.stubEnv('VITE_BACKEND_URL', 'http://localhost:8080')
  })

  afterEach(() => {
    vi.unstubAllEnvs()
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
      const signal = controller.signal

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ status: 'WAITING' })
      })

      await getStatus(signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/status', { signal })
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
          channelId: 'UC1',
          displayName: 'User1',
          joinedAt: '2024-01-01T09:00:00Z',
          firstCommentedAt: '2024-01-01T09:05:00Z',
          commentCount: 3
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
      const signal = controller.signal

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([])
      })

      await getUsers(signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/users.json', { signal })
    })
  })

  describe('postSwitchVideo', () => {
    test('正常にビデオ切替リクエストを送信', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await postSwitchVideo('test-video-id')

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/switch-video', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ videoId: 'test-video-id' }),
        signal: undefined
      })
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()
      const signal = controller.signal

      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await postSwitchVideo('test-video-id', signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/switch-video', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ videoId: 'test-video-id' }),
        signal
      })
    })

    test('HTTPエラーの場合例外を投げる', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400
      })

      await expect(postSwitchVideo('test-video-id')).rejects.toThrow('HTTP 400')
    })
  })

  describe('postPull', () => {
    test('正常にプルリクエストを送信', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await postPull()

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/pull', {
        method: 'POST',
        signal: undefined
      })
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()
      const signal = controller.signal

      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await postPull(signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/pull', {
        method: 'POST',
        signal
      })
    })
  })

  describe('postReset', () => {
    test('正常にリセットリクエストを送信', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await postReset()

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/reset', {
        method: 'POST',
        signal: undefined
      })
    })

    test('AbortSignalを正しく渡す', async () => {
      const controller = new AbortController()
      const signal = controller.signal

      mockFetch.mockResolvedValueOnce({
        ok: true
      })

      await postReset(signal)

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/reset', {
        method: 'POST',
        signal
      })
    })
  })

  describe('network error handling', () => {
    test('ネットワークエラーの場合例外を投げる', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      await expect(getStatus()).rejects.toThrow('Network error')
    })

    test('AbortErrorは正常に伝播される', async () => {
      const controller = new AbortController()
      const signal = controller.signal
      
      mockFetch.mockRejectedValueOnce(new DOMException('The operation was aborted', 'AbortError'))

      await expect(getStatus(signal)).rejects.toThrow('AbortError')
    })
  })
})