import { defineConfig, devices } from '@playwright/test';

const PORT = process.env.PORT || '8080';

export default defineConfig({
  testDir: 'tests/visual',
  timeout: 30_000,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: `http://localhost:${PORT}`,
    screenshot: 'only-on-failure'
  },
  projects: [
    {
      name: 'Desktop',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 800 } },
    },
  ],
});
