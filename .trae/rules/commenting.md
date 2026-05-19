# 中文注释规范

## 1. 总则

CaiRobot MVP 项目默认使用中文注释。注释应解释"为什么"和"约束"，而非重复代码表面意思。

## 2. 各语言注释规则

### 2.1 Go 注释规则

**必须使用中文注释的场景：**
- exported type（导出类型）必须说明用途和设计意图
- exported function（导出函数）必须说明输入、输出、前置条件
- exported interface（导出接口）必须说明契约和实现约束
- 常量、枚举必须说明取值含义

```go
// MessagePacket 表示网关层消息包，负责封装底层协议细节
// 不负责业务逻辑校验，校验由上层 Service 完成
type MessagePacket struct {
    // Version 表示协议版本号，当前固定为 1
    Version uint8
}
```

### 2.2 Protobuf 注释规则

**message 和 field 必须有中文注释：**

```protobuf
// HealthCheckRequest 健康检查请求
// 用于服务间心跳检测，不携带业务数据
message HealthCheckRequest {
  // service_name 表示请求方服务名称，用于日志追踪
  string service_name = 1;
}
```

### 2.3 Tars 组件注释规则

以下组件必须有中文说明：
- **Filter**：过滤逻辑和触发条件
- **AuditSink**：审计记录内容和格式
- **Invoker**：调用目标和超时策略
- **Router**：路由匹配规则和优先级
- **Protocol Adapter**：协议转换映射关系
- **Sanitizer**：脱敏规则和数据范围

### 2.4 TypeScript 注释规则

- public class 必须有中文说明职责
- public function/hook/service 必须有中文说明参数和返回值
- JSDoc 使用中文描述

```typescript
/**
 * RouterService 负责前端路由守卫和权限校验
 * 不处理业务数据获取，只做路由级控制
 */
export class RouterService {
  /**
   * canAccess 判断用户是否有权限访问目标页面
   * @param targetPath 目标路径
   * @param userRole 用户角色
   * @returns 是否允许访问
   */
  canAccess(targetPath: string, userRole: string): boolean { ... }
}
```

### 2.5 Python 注释规则

- public class/function 必须有中文 docstring
- docstring 包含 Args、Returns、Raises 说明

```python
class IntentClassifier:
    """意图分类器，负责将用户输入分类为预定义的意图类型
    
    不负责：
    - 提示词改写（由 PromptRewriteService 负责）
    - 回答审核（由 AnswerReviewService 负责）
    """

    def classify(self, text: str) -> str:
        """对输入文本进行意图分类
        
        Args:
            text: 用户原始输入文本
            
        Returns:
            意图类型字符串，如 HOMEWORK_EXPLANATION
            
        Raises:
            ValueError: 当输入文本为空时
        """
```

## 3. 特殊场景注释规则

### 3.1 审计逻辑

审计相关代码必须解释：
- 为什么记录此项
- 记录哪些字段
- 数据保留期限

### 3.2 鉴权逻辑

鉴权相关代码必须解释：
- 鉴权流程为什么这样设计
- 各角色的权限边界
- 安全假设和风险点

### 3.3 脱敏逻辑

脱敏代码必须解释：
- 脱敏原因（合规/安全）
- 脱敏规则选择依据
- 是否可逆

### 3.4 协议编号

协议编号变更必须解释：
- 为什么新增此字段
- 兼容性影响
- 版本迁移方案

### 3.5 错误码

错误码定义必须解释：
- 触发场景
- 用户可见性
- 恢复建议

## 4. 禁止事项

- ❌ 无意义注释：`// 设置变量`、`// 调用函数`
- ❌ 重复代码表面意思的注释
- ❌ 过时的注释（与代码不一致）
- ❌ 只写英文注释（除非是无法翻译的技术术语）

## 5. CI 检查

CI 通过 `make comment-check` 或 `python3 scripts/ci/check_chinese_comments.py` 执行检查：

1. Go exported 类型/函数/接口是否有注释
2. Protobuf message/field 是否有注释
3. 关键业务逻辑（审计/鉴权/脱敏/路由）是否有注释
