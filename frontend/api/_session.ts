import { createHash, createHmac, timingSafeEqual } from 'node:crypto'

// _ で始まる file は Vercel が関数として公開しないため、共通処理をここに置く。

export const SESSION_COOKIE = 'ycul_session'
const SESSION_TTL_SEC = 7 * 24 * 60 * 60

// 合言葉の総当たりを現実的でなくする最低限の長さ。これより短い設定は設定ミスとして扱う。
export const MIN_PASSPHRASE_LENGTH = 16
export const MIN_SESSION_SECRET_LENGTH = 32

function sign(payload: string, secret: string): string {
  return createHmac('sha256', secret).update(payload).digest('base64url')
}

// 長さが違うときに早期 return すると応答時間から長さを推測できるため、両方を hash して同じ長さにしてから比べる。
function safeEqual(a: string, b: string): boolean {
  const ha = createHash('sha256').update(a).digest()
  const hb = createHash('sha256').update(b).digest()
  return timingSafeEqual(ha, hb)
}

export function checkPassphrase(input: string, expected: string): boolean {
  if (expected.length < MIN_PASSPHRASE_LENGTH) return false
  return safeEqual(input, expected)
}

export function issueSession(secret: string, now = Date.now()): string {
  const exp = String(Math.floor(now / 1000) + SESSION_TTL_SEC)
  return `${exp}.${sign(exp, secret)}`
}

export function verifySession(
  token: string | undefined,
  secret: string,
  now = Date.now(),
): boolean {
  if (secret.length < MIN_SESSION_SECRET_LENGTH || !token) return false
  const [exp, mac, ...rest] = token.split('.')
  if (!exp || !mac || rest.length > 0) return false
  if (!safeEqual(mac, sign(exp, secret))) return false
  return Number(exp) > Math.floor(now / 1000)
}

export function readCookie(req: Request, name: string): string | undefined {
  const header = req.headers.get('cookie') ?? ''
  for (const part of header.split(';')) {
    const [key, ...value] = part.trim().split('=')
    if (key === name) return value.join('=')
  }
  return undefined
}

export function sessionCookie(token: string): string {
  return `${SESSION_COOKIE}=${token}; Path=/api; HttpOnly; Secure; SameSite=Strict; Max-Age=${SESSION_TTL_SEC}`
}

export function jsonResponse(
  body: unknown,
  status: number,
  headers: Record<string, string> = {},
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json', ...headers },
  })
}

export function unauthorized(): Response {
  return jsonResponse({ error: 'unauthorized', message: 'login required', httpCode: 401 }, 401)
}

export function misconfigured(): Response {
  return jsonResponse({ error: 'server misconfigured', httpCode: 500 }, 500)
}
