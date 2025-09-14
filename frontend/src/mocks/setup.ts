import { beforeAll, afterAll, afterEach, vi } from 'vitest'
import { setupServer } from 'msw/node'
import { handlers, resetMockState } from './handlers'
import 'whatwg-fetch'
import '@testing-library/jest-dom/vitest'

export const server = setupServer(...handlers)

beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' })
})

afterEach(() => {
  server.resetHandlers()
  resetMockState()
  vi.useRealTimers()
})

afterAll(() => server.close())
