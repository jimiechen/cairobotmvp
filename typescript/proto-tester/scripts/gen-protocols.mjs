/**
 * gen-protocols.mjs - 构建期协议元数据提取脚本
 *
 * 职责：
 * - 使用 protobufjs 解析 proto/base/ 下所有 .proto 文件
 * - 从协议编号注册表读取协议路由信息
 * - 递归提取所有 message 的字段元数据（含嵌套 message）
 * - 输出 src/data/protocols.json 作为运行时唯一数据来源
 *
 * 约束：
 * - 本脚本是唯一允许引入 protobufjs 的位置
 * - 运时代码禁止 import .proto 或使用 protobufjs
 */
import { readFileSync, writeFileSync, mkdirSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import protobuf from 'protobufjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const PROJECT_ROOT = resolve(__dirname, '..');
const PROTO_DIR = resolve(PROJECT_ROOT, '../../proto/base');
const OUTPUT_PATH = resolve(PROJECT_ROOT, 'src/data/protocols.json');
const REGISTRY_PATH = resolve(PROJECT_ROOT, '../../docs/api/协议编号注册表.md');

/** 需要解析的 proto 文件列表（与协议编号注册表一致） */
const PROTO_FILES = [
  'message.proto',
  'hello.proto',
  'health.proto',
  'app_config.proto',
  'i18n.proto',
  'result.proto',
];

/**
 * 解析协议编号注册表的 Markdown 表格
 * 提取 max/min/方向/状态/说明/proto文件/message 等字段
 *
 * 输入：协议编号注册表.md 的文本内容
 * 输出：协议记录数组
 */
function parseRegistryTable(markdown) {
  const lines = markdown.split('\n');
  const records = [];
  let inTable = false;

  for (const line of lines) {
    // 检测表格开始（表头行）
    if (line.includes('|') && line.includes('max') && line.includes('min')) {
      inTable = true;
      continue;
    }
    // 跳过分隔行
    if (inTable && /^\|[\s\-:|]+\|$/.test(line)) {
      continue;
    }
    // 表格结束检测
    if (inTable && !line.trim().startsWith('|')) {
      if (records.length > 0) break;
      inTable = false;
      continue;
    }
    // 解析数据行
    if (inTable && line.startsWith('|')) {
      const cells = line.split('|').filter((c) => c !== '').map((c) => c.trim());
      // 至少需要 max, min, 方向, 报文类型, proto文件, message, openapi, status, 说明
      if (cells.length >= 9) {
        const maxVal = cells[0].replace(/:/g, '');
        const minVal = cells[1].replace(/:/g, '');
        const maxNum = Number.parseInt(maxVal, 10);
        const minNum = Number.parseInt(minVal, 10);
        if (!Number.isNaN(maxNum) && !Number.isNaN(minNum)) {
          records.push({
            maxType: maxNum,
            minType: minNum,
            direction: cells[2],
            messageType: cells[3],
            protoFile: cells[4].replace(/`/g, ''),
            name: cells[5].replace(/`/g, ''),
            openapi: cells[6].replace(/`/g, ''),
            status: cells[7],
            description: cells[8],
          });
        }
      }
    }
  }

  return records;
}

/**
 * 从 protobufjs Root 中递归提取所有 message schema
 *
 * protobufjs 加载后，message 嵌套在 package 命名空间下（如 com.mineplanet.pojo.hello）
 * 需要递归遍历所有层级的 nested 来找到全部 Type 实例
 *
 * @param {object} root - protobufjs 解析后的 Root 对象
 * @returns {Record<string, object>} messageName → MessageSchema
 */
function extractAllMessageSchemas(root) {
  const schemas = {};

  /**
   * 递归遍历命名空间，找到所有 protobuf.Type（message）
   * @param {string} namespacePath - 当前命名空间路径前缀
   * @param {object} node - 当前节点（Root 或 Namespace）
   */
  function traverseNamespace(namespacePath, node) {
    if (!node || !node.nested) return;

    for (const [name, child] of Object.entries(node.nested)) {
      if (child instanceof protobuf.Type) {
        // 构造完整全名：命名空间路径 + message 名
        const fullName = namespacePath ? `${namespacePath}.${name}` : name;
        visitMessage(fullName, child);
      } else if (child instanceof protobuf.Namespace) {
        // 继续深入命名空间
        const childPath = namespacePath ? `${namespacePath}.${name}` : name;
        traverseNamespace(childPath, child);
      }
    }
  }

  function visitMessage(fullName, message) {
    const fields = [];
    const nestedTypes = [];
    const enums = [];

    // 提取 oneof 分组信息，用于标记字段的 oneof 所属组
    const oneofGroups = {};
    if (message.oneofs) {
      for (const [oneofName, oneof] of Object.entries(message.oneofs)) {
        for (const field of oneof.fieldsArray) {
          oneofGroups[field.name] = oneofName;
        }
      }
    }

    // 提取字段
    if (message.fields) {
      for (const [fieldName, field] of Object.entries(message.fields)) {
        fields.push({
          name: fieldName,
          number: field.id,
          type: field.type,
          label: field.repeated ? 'repeated' : (field.map ? 'map' : 'optional'),
          oneof: oneofGroups[fieldName] || null,
          comment: field.comment || null,
        });
      }
    }

    // 排序保证确定性输出
    fields.sort((a, b) => a.number - b.number);

    // 提取嵌套 enum
    if (message.nested) {
      for (const [nestedName, nested] of Object.entries(message.nested)) {
        if (nested instanceof protobuf.Enum) {
          const values = {};
          for (const [valName, valNum] of Object.entries(nested.values)) {
            values[valName] = valNum;
          }
          enums.push({
            name: nestedName,
            fullName: `${fullName}.${nestedName}`,
            values,
          });
        }
      }
    }

    // 提取嵌套 message 并递归
    if (message.nested) {
      for (const [nestedName, nested] of Object.entries(message.nested)) {
        if (nested instanceof protobuf.Type) {
          const nestedFullName = `${fullName}.${nestedName}`;
          nestedTypes.push(nestedFullName);
          visitMessage(nestedFullName, nested);
        }
      }
    }

    schemas[fullName] = {
      fullName,
      protoFile: message.filename || '',
      fields,
      nestedTypes,
      enums,
    };
  }

  // 从根节点开始递归遍历所有命名空间，找到全部 message
  traverseNamespace('', root);

  return schemas;
}

/**
 * 主函数：解析 proto 文件并生成 protocols.json
 */
async function main() {
  console.log('[gen-protocols] 开始解析 proto 文件...');

  // 1. 用 protobufjs 解析所有 proto 文件（keepCase 保留原始 snake_case 字段名）
  const root = new protobuf.Root();
  root.resolvePath = (origin, target) => {
    // 将 import 路径映射到实际文件系统路径
    if (target.startsWith('base/')) {
      return resolve(PROTO_DIR, target.replace('base/', ''));
    }
    return resolve(dirname(origin), target);
  };

  for (const protoFile of PROTO_FILES) {
    const filePath = resolve(PROTO_DIR, protoFile);
    console.log(`[gen-protocols] 解析: ${protoFile}`);
    await root.load(filePath, { keepCase: true });
  }

  // 2. 提取所有 message schema
  const messageSchemas = extractAllMessageSchemas(root);
  console.log(`[gen-protocols] 提取到 ${Object.keys(messageSchemas).length} 个 message schema`);

  // 3. 解析协议编号注册表
  const registryContent = readFileSync(REGISTRY_PATH, 'utf-8');
  const protocolRecords = parseRegistryTable(registryContent);
  console.log(`[gen-protocols] 解析到 ${protocolRecords.length} 条协议记录`);

  // 4. 补充 requestMessage / responseMessage 字段
  // 将同一 maxType 组的 req/rsp 配对
  const protocols = protocolRecords.map((record) => ({
    ...record,
    requestMessage: record.messageType === 'Request' ? record.name : undefined,
    responseMessage: record.messageType === 'Response' ? record.name : undefined,
  }));

  // 5. 构建最终输出
  const output = {
    version: '1',
    generatedAt: new Date().toISOString(),
    protocols,
    messageSchemas,
  };

  // 6. 写入文件
  const outputDir = dirname(OUTPUT_PATH);
  mkdirSync(outputDir, { recursive: true });
  writeFileSync(OUTPUT_PATH, JSON.stringify(output, null, 2), 'utf-8');
  console.log(`[gen-protocols] 已生成: ${OUTPUT_PATH}`);
  console.log(`[gen-protocols] 协议数: ${protocols.length}, message 数: ${Object.keys(messageSchemas).length}`);
}

main().catch((err) => {
  console.error('[gen-protocols] 执行失败:', err);
  process.exit(1);
});
