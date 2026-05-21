/// <reference types="vitest" />
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import path from 'path'

export default defineConfig({
	plugins: [react()],
	resolve: {
		alias: {
			'@proto': path.resolve(__dirname, '../../proto/generated/ts'),
			'@utils': path.resolve(__dirname, 'src/utils'),
			'@pages': path.resolve(__dirname, 'src/pages'),
			'google-protobuf': path.resolve(__dirname, 'node_modules/google-protobuf'),
		},
	},
	optimizeDeps: {
		include: ['google-protobuf'],
	},
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['tests/**/*.test.{ts,tsx}'],
	},
})
