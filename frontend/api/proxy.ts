import {
  MIN_SESSION_SECRET_LENGTH,
  SESSION_COOKIE,
  jsonResponse,
  misconfigured,
  readCookie,
  unauthorized,
  verifySession,
} from './_session.js'

// backend の X-API-Key はこの関数の環境変数にだけ置き、ブラウザの JS には含めない。
// vercel.json の rewrite が /api/b/<path> を /api/proxy?path=<path> に変換して呼ぶ。

// 中継してよい backend の path。backend に route が増えたらここにも追加する。
const ALLOWED_PATHS = [
  /^status$/,
  /^users\.json$/,
  /^switch-video$/,
  /^pull$/,
  /^reset$/,
  /^comments$/,
  /^history\/snapshots(\/[A-Za-z0-9_-]+)?$/,
]

export default {
  async fetch(req: Request): Promise<Response> {
    const secret = process.env.SESSION_SECRET ?? ''
    const backendUrl = process.env.BACKEND_URL ?? ''
    const apiKey = process.env.API_KEY ?? ''
    if (secret.length < MIN_SESSION_SECRET_LENGTH || !backendUrl || !apiKey) {
      return misconfigured()
    }
    if (!verifySession(readCookie(req, SESSION_COOKIE), secret)) {
      return unauthorized()
    }

    const url = new URL(req.url)
    const path = url.searchParams.get('path') ?? ''
    url.searchParams.delete('path')
    if (!ALLOWED_PATHS.some((re) => re.test(path))) {
      return jsonResponse({ error: 'not found', httpCode: 404 }, 404)
    }

    const headers: Record<string, string> = { 'X-API-Key': apiKey }
    const contentType = req.headers.get('content-type')
    if (contentType) headers['Content-Type'] = contentType

    const hasBody = req.method !== 'GET' && req.method !== 'HEAD'
    const upstream = await fetch(`${backendUrl.replace(/\/$/, '')}/${path}${url.search}`, {
      method: req.method,
      headers,
      body: hasBody ? await req.arrayBuffer() : undefined,
      signal: req.signal,
    })

    return new Response(upstream.body, {
      status: upstream.status,
      headers: { 'Content-Type': upstream.headers.get('content-type') ?? 'application/json' },
    })
  },
}
