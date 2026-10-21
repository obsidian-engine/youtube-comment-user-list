import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import login from '../../api/login'
import proxy from '../../api/proxy'
import { SESSION_COOKIE, issueSession, verifySession } from '../../api/_session'

const SECRET = 's'.repeat(32)
const PASSPHRASE = 'correct horse battery staple'

function sessionHeader(token = issueSession(SECRET)): Record<string, string> {
  return { cookie: `${SESSION_COOKIE}=${token}` }
}

describe('session', () => {
  it('rejects a token signed with another secret', () => {
    expect(verifySession(issueSession('x'.repeat(32)), SECRET)).toBe(false)
  })

  it('rejects an expired token', () => {
    const token = issueSession(SECRET, Date.now() - 8 * 24 * 60 * 60 * 1000)
    expect(verifySession(token, SECRET)).toBe(false)
  })

  it('rejects every token when the secret is empty', () => {
    expect(verifySession(issueSession(''), '')).toBe(false)
  })
})

describe('api handlers', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    vi.stubEnv('SESSION_SECRET', SECRET)
    vi.stubEnv('APP_PASSPHRASE', PASSPHRASE)
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

  it('login issues a session cookie only for the correct passphrase', async () => {
    const wrong = await login.fetch(
      new Request('https://app.example.com/api/login', {
        method: 'POST',
        body: JSON.stringify({ passphrase: 'wrong' }),
      }),
    )
    expect(wrong.status).toBe(401)
    expect(wrong.headers.get('set-cookie')).toBeNull()

    const ok = await login.fetch(
      new Request('https://app.example.com/api/login', {
        method: 'POST',
        body: JSON.stringify({ passphrase: PASSPHRASE }),
      }),
    )
    expect(ok.status).toBe(204)
    expect(ok.headers.get('set-cookie')).toContain('HttpOnly')
  })

  it('login fails closed when the passphrase is not configured', async () => {
    vi.stubEnv('APP_PASSPHRASE', '')
    const res = await login.fetch(
      new Request('https://app.example.com/api/login', {
        method: 'POST',
        body: JSON.stringify({ passphrase: '' }),
      }),
    )
    expect(res.status).toBe(500)
  })

  it('proxy rejects requests without a session', async () => {
    const res = await proxy.fetch(new Request('https://app.example.com/api/proxy?path=status'))
    expect(res.status).toBe(401)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('proxy rejects paths outside the allow-list', async () => {
    const res = await proxy.fetch(
      new Request('https://app.example.com/api/proxy?path=../admin', { headers: sessionHeader() }),
    )
    expect(res.status).toBe(404)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('proxy forwards the query and adds the backend key', async () => {
    const res = await proxy.fetch(
      new Request('https://app.example.com/api/proxy?path=comments&keywords=a%2Cb', {
        headers: sessionHeader(),
      }),
    )
    expect(res.status).toBe(200)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('https://backend.example.com/comments?keywords=a%2Cb')
    expect(init.headers['X-API-Key']).toBe('backend-key')
  })
})
