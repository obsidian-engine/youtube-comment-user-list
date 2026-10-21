import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// npm run dev 用。本番では api/proxy.ts が同じ役割を担う。
// VITE_ で始まらない変数は bundle に含まれないため、API key は dev server の proxy だけが読む。
const devEnv = loadEnv('development', process.cwd(), '')

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api/b': {
        target: devEnv.BACKEND_URL || 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/b/, ''),
        headers: { 'X-API-Key': devEnv.API_KEY || '' },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/mocks/setup.ts'],

    // 並列実行設定（軽量化のため並列実行有効化）
    pool: 'forks',
    maxWorkers: 4,
    minWorkers: 1,

    // タイムアウト設定の軽量化
    testTimeout: 10000, // 10秒に短縮
    hookTimeout: 10000, // setup/teardownも10秒に短縮

    // キャッシュ設定は削除（警告回避）

    // レポーター設定（さらなる軽量化）
    reporters: process.env.CI ? ['junit', 'github-actions'] : ['default'],

    // ファイル監視の最適化
    watch: process.env.CI ? false : true,

    // テスト分離を有効化（コンポーネントレンダリングに必要）
    isolate: true,

    // 依存関係最適化（バンドル化による高速化）
    deps: {
      optimizer: {
        web: {
          enabled: true,
          include: ['react', 'react-dom', '@testing-library/react', '@testing-library/user-event'],
        },
      },
    },
  },
})
