export const logger = {
  log: (...args: unknown[]) => {
    // テスト環境ではログを無効化
    if (process.env.NODE_ENV !== 'test') {
      console.log(...args)
    }
  },
  error: (...args: unknown[]) => {
    console.error(...args)
  },
  warn: (...args: unknown[]) => {
    console.warn(...args)
  },
}