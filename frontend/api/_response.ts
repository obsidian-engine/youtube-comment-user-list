// _ で始まる file は Vercel が関数として公開しないため、共通処理をここに置く。

export function jsonResponse(body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

export function misconfigured(): Response {
  return jsonResponse({ error: 'server misconfigured', httpCode: 500 }, 500)
}
