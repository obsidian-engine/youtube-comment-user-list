import { beforeAll, afterAll, afterEach, vi } from 'vitest'
import { setupServer } from 'msw/node'
import { handlers, resetMockState } from './handlers'
import 'whatwg-fetch'
import '@testing-library/jest-dom/vitest'

// whatwg-fetch が jsdom の File.text() を壊すため再定義
// File.prototype.text が未定義の場合に Blob の実装を継承させる
if (typeof File !== 'undefined' && !File.prototype.text) {
  File.prototype.text = function () {
    return new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = () => reject(reader.error)
      reader.readAsText(this)
    })
  }
}

// matchMediaのグローバルモック
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// localStorageのグローバルモック
Object.defineProperty(window, 'localStorage', {
  writable: true,
  value: {
    getItem: vi.fn(() => null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
  },
})

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
