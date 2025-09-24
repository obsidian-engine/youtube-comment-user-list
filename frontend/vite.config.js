import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/mocks/setup.ts'],

    // 並列実行設定（大幅な速度向上）
    pool: 'threads',
    poolOptions: {
      threads: {
        maxWorkers: '50%', // CPUコア数の50%を使用
        minWorkers: 1
      }
    },

    // タイムアウト設定の最適化
    testTimeout: 10000, // デフォルト10秒（統合テスト用）
    hookTimeout: 10000, // setup/teardownのタイムアウト

    // キャッシュ設定は削除（警告回避）

    // レポーター設定（軽量化）
    reporter: process.env.CI ? ['junit', 'github-actions'] : ['verbose'],

    // ファイル監視の最適化
    watch: process.env.CI ? false : true,

    // 並列実行時の分離設定
    isolate: true
  },
})
