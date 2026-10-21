import {
  MIN_PASSPHRASE_LENGTH,
  MIN_SESSION_SECRET_LENGTH,
  SESSION_COOKIE,
  checkPassphrase,
  issueSession,
  jsonResponse,
  misconfigured,
  readCookie,
  sessionCookie,
  unauthorized,
  verifySession,
} from './_session.js'

// GET: session が有効かを返す。POST: 合言葉を確かめて session cookie を発行する。
export default {
  async fetch(req: Request): Promise<Response> {
    const secret = process.env.SESSION_SECRET ?? ''
    const passphrase = process.env.APP_PASSPHRASE ?? ''
    if (secret.length < MIN_SESSION_SECRET_LENGTH || passphrase.length < MIN_PASSPHRASE_LENGTH) {
      return misconfigured()
    }

    if (req.method === 'GET') {
      return verifySession(readCookie(req, SESSION_COOKIE), secret)
        ? new Response(null, { status: 204 })
        : unauthorized()
    }
    if (req.method !== 'POST') {
      return jsonResponse({ error: 'method not allowed', httpCode: 405 }, 405)
    }

    let body: unknown
    try {
      body = await req.json()
    } catch {
      return jsonResponse({ error: 'invalid json', httpCode: 400 }, 400)
    }
    const input = (body as { passphrase?: unknown } | null)?.passphrase
    if (typeof input !== 'string' || !checkPassphrase(input, passphrase)) {
      return unauthorized()
    }

    return new Response(null, {
      status: 204,
      headers: { 'Set-Cookie': sessionCookie(issueSession(secret)) },
    })
  },
}
