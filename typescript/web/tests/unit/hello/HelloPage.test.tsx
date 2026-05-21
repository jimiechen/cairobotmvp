/**
 * HelloPage 组件单元测试
 *
 * 测试范围：UI 组件渲染、状态切换、数据展示
 * Mock 外部依赖（fetch），不依赖真实 Gateway 服务
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { HelloPage } from '@pages/hello/HelloPage';

beforeEach(() => {
	vi.clearAllMocks();
});

describe('HelloPage', () => {
	it('显示加载状态', async () => {
		global.fetch = vi.fn(() => new Promise(() => {})) as any;

		render(<HelloPage />);

		expect(screen.getByTestId('hello-loading')).toBeTruthy();
	});

	it('显示错误状态', async () => {
		global.fetch = vi.fn(() => Promise.reject(new Error('Network error'))) as any;

		render(<HelloPage />);

		await waitFor(() => {
			expect(screen.getByTestId('hello-error')).toBeTruthy();
		});
	});

	it('显示两个服务的消息', async () => {
		global.fetch = vi.fn((url: string) => {
			if (url.includes('8080')) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({
						result: { code: 10200, message: 'success' },
						message: 'Hello from Go!',
						timestamp: '2024-01-01T00:00:00Z',
					}),
				});
			}
			if (url.includes('8081')) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({
						result: { code: 10200, message: 'success' },
						message: 'Hello from Python!',
						timestamp: '2024-01-01T00:00:00Z',
					}),
				});
			}
			return Promise.resolve({ ok: false });
		}) as any;

		render(<HelloPage />);

		await waitFor(() => {
			expect(screen.getByTestId('hello-page')).toBeTruthy();
			expect(screen.getByTestId('go-message')).toBeTruthy();
			expect(screen.getByTestId('py-message')).toBeTruthy();
		});
	});
});
