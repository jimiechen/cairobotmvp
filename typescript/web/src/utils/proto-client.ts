/**
 * Proto-Gateway 客户端工具函数
 *
 * 封装与 proto-gateway 通信的通用逻辑：
 *   - MessagePacket 构建（Protobuf 二进制序列化）
 *   - HTTP POST 请求发送
 *   - 响应反序列化
 *
 * 使用 proto/generated/ts 官方生成的 Protobuf 类型
 */

import { com as messageCom } from '@proto/base/message';

const {
	mineplanet: {
		pojo: { MessagePacket, Platform },
	},
} = messageCom;

const DEFAULT_GATEWAY_URL = 'http://localhost:8080/api/hello';

export interface GatewayResponse {
	status: number;
	data: ArrayBuffer;
	packet?: InstanceType<typeof MessagePacket>;
}

export interface PacketOptions {
	maxType: number;
	minType: number;
	method: string;
	data?: string;
	traceId?: string;
	requestId?: string;
}

export function buildPacket(options: PacketOptions): Uint8Array {
	const packet = new MessagePacket({
		maxType: options.maxType,
		minType: options.minType,
		platform: Platform.WEB,
		extend: new Map<string, string>([
			['method', options.method],
			['traceId', options.traceId || `ts-client-${Date.now()}`],
			['requestId', options.requestId || `ts-req-${Date.now()}`],
		]),
		data: Uint8Array.from(Buffer.from(options.data || 'Hello World', 'utf-8')),
	});
	return packet.serializeBinary();
}

export async function postGateway(
	body: Uint8Array,
	url?: string,
): Promise<GatewayResponse> {
	const gatewayUrl = url || process.env.GATEWAY_TEST_URL || DEFAULT_GATEWAY_URL;

	const resp = await fetch(gatewayUrl, {
		method: 'POST',
		headers: { 'Content-Type': 'application/octet-stream' },
		body: Buffer.from(body) as unknown as BodyInit,
	});

	const data = await resp.arrayBuffer();

	let packet: InstanceType<typeof MessagePacket> | undefined;
	if (data.byteLength > 0) {
		try {
			packet = MessagePacket.deserializeBinary(new Uint8Array(data));
		} catch (e) {
			console.warn('[ProtoClient] 响应反序列化失败:', e);
		}
	}

	return { status: resp.status, data, packet };
}

export function getGatewayUrl(): string {
	return process.env.GATEWAY_TEST_URL || DEFAULT_GATEWAY_URL;
}

export async function isGatewayAvailable(url?: string): Promise<boolean> {
	const gatewayUrl = url || getGatewayUrl();
	try {
		const resp = await fetch(gatewayUrl, { method: 'HEAD' });
		console.log(`[ProtoClient] Gateway 可达: ${gatewayUrl} (HTTP ${resp.status})`);
		return true;
	} catch (err) {
		console.warn(`[ProtoClient] ⚠️  Gateway 不可达: ${gatewayUrl}`);
		return false;
	}
}

export { MessagePacket, Platform };
