import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/mocks/setup.ts'],

    // 並列実行設定（軽量化のため並列実行有効化）
    pool: 'forks',
    poolOptions: {
      forks: {
        maxWorkers: 4, // 軽量化のため並列実行
        minWorkers: 1
      }
    },

    // タイムアウト設定の軽量化
    testTimeout: 10000, // 10秒に短縮
    hookTimeout: 10000, // setup/teardownも10秒に短縮

    // キャッシュ設定は削除（警告回避）

    // レポーター設定（さらなる軽量化）
    reporter: process.env.CI ? ['junit', 'github-actions'] : ['basic'],

    // ファイル監視の最適化
    watch: process.env.CI ? false : true,

    // 軽量化のため分離を無効化（単体テストは問題なし）
    isolate: false,

    // 依存関係最適化（バンドル化による高速化）
    deps: {
      optimizer: {
        web: {
          enabled: true,
          include: ['react', 'react-dom', '@testing-library/react', '@testing-library/user-event']
        }
      }
    }
  },
})
