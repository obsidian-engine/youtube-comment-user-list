/// <reference types="vite/client" />

// backend の URL と API key は bundle に含めないため、VITE_ の変数を定義しない。
// 本番は api/proxy.ts、npm run dev は vite.config.js の proxy が読む。
