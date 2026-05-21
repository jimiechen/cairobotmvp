/**
 * Proto-Gateway E2E 链路验收测试
 *
 * 测试范围：HTTP → Gateway → LocalInvoker → modules 全链路
 *
 * 前置条件：
 *   1. proto-gateway 已启动（默认 http://localhost:8080/api/hello）
 *   2. 或设置环境变量 GATEWAY_TEST_URL 指向目标地址
 *
 * 参考：go/gateway/proto-gateway/cmd/testclient/main.go
 */

import { describe, it, expect, beforeAll } from 'vitest';
import {
	buildPacket,
	postGateway,
	isGatewayAvailable,
	MessagePacket,
} from '@utils/proto-client';

describe('Proto-Gateway E2E Modules 链路验收', () => {
	let gatewayAvailable = false;

	beforeAll(async () => {
		gatewayAvailable = await isGatewayAvailable();
		if (!gatewayAvailable) {
			console.warn('[E2E]    请先启动: cd go/gateway/proto-gateway && go run ./cmd/server/');
		}
	});

	describe('HealthCheck 路由 (minType=2097)', () => {
		it('正常请求返回 code=10200 + HTTP 200', async () => {
			if (!gatewayAvailable) return;

			const body = buildPacket({
				maxType: 2100,
				minType: 2097,
				method: 'HealthCheck',
				data: 'ts-e2e-health',
			});

			const resp = await postGateway(body);

			expect(resp.status).toBe(200);
			expect(resp.packet).toBeDefined();
			expect(resp.packet!.maxType).toBe(2100);
			expect(resp.packet!.extend?.get('code')).toBe('10200');

			if (resp.packet?.data?.length) {
				const responseData = new TextDecoder().decode(resp.packet.data);
				console.log(`[E2E] HealthCheck 响应数据: ${responseData}`);
			}
		}, 10000);
	});

	describe('HelloWorld 路由 (minType=2101)', () => {
		it('正常请求返回 code=10200 + HTTP 200', async () => {
			if (!gatewayAvailable) return;

			const body = buildPacket({
				maxType: 2100,
				minType: 2101,
				method: 'HelloWorld',
				data: 'TypescriptE2E',
			});

			const resp = await postGateway(body);

			expect(resp.status).toBe(200);
			expect(resp.packet).toBeDefined();
			expect(resp.packet!.maxType).toBe(2100);
			expect(resp.packet!.extend?.get('code')).toBe('10200');

			if (resp.packet?.data?.length) {
				const responseData = new TextDecoder().decode(resp.packet.data);
				console.log(`[E2E] HelloWorld 响应数据: ${responseData}`);
			}
		}, 10000);
	});

	describe('错误处理链路', () => {
		it('未匹配路由返回业务错误码 10404', async () => {
			if (!gatewayAvailable) return;

			const body = buildPacket({
				maxType: 9999,
				minType: 9999,
				method: 'UnknownMethod',
			});

			const resp = await postGateway(body);

			expect(resp.status).toBe(200);
			expect(resp.packet).toBeDefined();
			expect(resp.packet!.extend?.get('code')).toBe('10404');
		}, 5000);
	});
});
