import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import proxy from '../../api/proxy'

describe('api proxy', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    vi.stubEnv('BACKEND_URL', 'https://backend.example.com/')
    vi.stubEnv('API_KEY', 'backend-key')
    fetchMock.mockResolvedValue(new Response('{}', { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
    fetchMock.mockReset()
  })

  it('fails closed when the backend key is not configured', async () => {
    vi.stubEnv('API_KEY', '')
    const res = await proxy.fetch(new Request('https://app.example.com/api/proxy?path=status'))
    expect(res.status).toBe(500)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('rejects paths outside the allow-list', async () => {
    const res = await proxy.fetch(new Request('https://app.example.com/api/proxy?path=../admin'))
    expect(res.status).toBe(404)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('forwards the query and adds the backend key', async () => {
    const res = await proxy.fetch(
      new Request('https://app.example.com/api/proxy?path=comments&keywords=a%2Cb'),
    )
    expect(res.status).toBe(200)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('https://backend.example.com/comments?keywords=a%2Cb')
    expect(init.headers['X-API-Key']).toBe('backend-key')
  })
})
