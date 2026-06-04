import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@proto': path.resolve(__dirname, '../../proto/generated/ts'),
      'google-protobuf': path.resolve(__dirname, 'node_modules/google-protobuf'),
    },
  },
  optimizeDeps: {
    include: ['google-protobuf'],
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
    host: '127.0.0.1',
    port: 3001,
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['tests/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'text-summary', 'lcov'],
      exclude: [
        'scripts/**',
        'src/main.tsx',
        'src/App.tsx',
        'src/vite-env.d.ts',
        // CLI 入口和命令（需 E2E 或真实 Gateway 测试）
        'src/cli/index.ts',
        'src/cli/commands/send.ts',
        'src/cli/commands/trace.ts',
        'src/cli/commands/run.ts',
        'src/cli/commands/capture.ts',
        'src/cli/browser.ts',          // Playwright，需真实浏览器
        // 页面级组件（适合 E2E 测试）
        'src/routes/**',
        'src/routes/components/**',
        // antd 图标依赖复杂（jsdom 环境限制）
        'src/components/TokenInjector.tsx',
        'src/components/UserSwitcher.tsx',
        'src/components/ExtendParamsPanel.tsx',
        // 复杂 UI 组件（需 Playwright E2E 测试覆盖）
        'src/components/ProtoFormRenderer.tsx',
        // IndexedDB 依赖（需 fake-indexeddb 或 E2E）
        'src/store/history.ts',
        'src/store/historyCleanup.ts',
        'src/store/protocols.ts',
        // 测试文件自身
        'tests/**',
        '**/*.d.ts',
        '**/*.test.{ts,tsx}',
        '**/node_modules/**',
      ],
      thresholds: {
        statements: 80,
        branches: 75,
        functions: 68,
        lines: 80,
      },
    },
  },
});
