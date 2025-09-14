import { http, HttpResponse } from 'msw'

// 簡易なメモリ状態（各テストで server.use で上書き可）
let state: 'WAITING' | 'ACTIVE' = 'WAITING'
let users: string[] = []

export const resetMockState = () => {
  state = 'WAITING'
  users = []
}

export const handlers = [
  // 状態
  http.get('*/status', () => {
    return HttpResponse.json({ status: state, count: users.length })
  }),

  // ユーザー一覧
  http.get('*/users.json', () => {
    return HttpResponse.json(users)
  }),

  // 切替
  http.post('*/switch-video', async ({ request }) => {
    // videoId バリデーションっぽいもの
    try {
      const body = (await request.json()) as { videoId?: string }
      if (!body?.videoId) return new HttpResponse('bad request', { status: 400 })
    } catch {
      return new HttpResponse('bad request', { status: 400 })
    }
    state = 'ACTIVE'
    users = []
    return new HttpResponse(null, { status: 200 })
  }),

  // 今すぐ取得
  http.post('*/pull', () => {
    users.push(`User-${users.length + 1}`)
    return new HttpResponse(null, { status: 200 })
  }),

  // リセット
  http.post('*/reset', () => {
    state = 'WAITING'
    users = []
    return new HttpResponse(null, { status: 200 })
  }),
]

export const __mock = {
  get state() {
    return state
  },
  set state(v: 'WAITING' | 'ACTIVE') {
    state = v
  },
  get users() {
    return users
  },
  set users(v: string[]) {
    users = v
  },
}
