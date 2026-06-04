/**
 * T0 冒烟测试：验证工程基础配置
 *
 * 验证项：
 * 1. @proto/* alias 可解析到生成的 TS 代码
 * 2. google-protobuf 可正常导入（精确版本 3.21.2）
 * 3. MessagePacket 类可实例化（getter/setter API）
 */
import { describe, it, expect } from 'vitest';
import { com as messageCom } from '@proto/base/message';
const { MessagePacket, Platform } = messageCom.mineplanet.pojo;

describe('T0 工程配置冒烟', () => {
  it('@proto alias 可导入 MessagePacket', () => {
    // 验证 @proto/base/message 可正确解析（嵌套命名空间 com.mineplanet.pojo）
    expect(MessagePacket).toBeDefined();
    // protoc-gen-ts 生成的类：实例化 + 序列化能力在下一个用例验证
    const packet = new MessagePacket();
    expect(packet).toBeInstanceOf(MessagePacket);
  });

  it('Platform 枚举值正确', () => {
    expect(Platform.UNKNOWN).toBe(0);
    expect(Platform.WEB).toBe(1);
    expect(Platform.PC).toBe(2);
    expect(Platform.ANDROID).toBe(3);
    expect(Platform.IOS).toBe(4);
    expect(Platform.OTHER).toBe(5);
  });

  it('MessagePacket 可序列化和反序列化（getter/setter API）', () => {
    const packet = new MessagePacket();
    // protoc-gen-ts 使用 setter 属性赋值，而非 setMaxType() 方法
    packet.maxType = 2100;
    packet.minType = 2101;
    packet.platform = Platform.WEB;

    const binary = packet.serialize();
    expect(binary instanceof Uint8Array).toBe(true);
    expect(binary.length).toBeGreaterThan(0);

    const restored = MessagePacket.deserialize(binary);
    expect(restored.maxType).toBe(2100);
    expect(restored.minType).toBe(2101);
    expect(restored.platform).toBe(Platform.WEB);
  });
});
