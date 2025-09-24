const isDevelopment = import.meta.env.DEV

export const logger = {
  log: (...args: unknown[]) => {
    // 本番環境でもデバッグのため一時的に有効化
    console.log(...args)
  },
  error: (...args: unknown[]) => {
    console.error(...args)
  },
  warn: (...args: unknown[]) => {
    console.warn(...args)
  },
}