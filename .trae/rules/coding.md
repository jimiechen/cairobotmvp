# 编码规范

## 1. 命名规范

### 1.1 变量命名
- 变量、函数参数、局部变量：lowerCamelCase
- 布尔变量必须以 is/has/can/should 开头
- 集合变量优先使用复数名
- 避免无意义命名，如 data、temp、obj、handle、doit
- 禁止拼音变量名，除非是无法翻译的领域词

示例：
```text
doorStatus
phonePresent
isLocked
hasParentPermission
allowedIntents
currentState
```

不推荐：
```text
d
tempData
obj1
kaiguan
```

### 1.2 常量命名
常量使用 UPPER_SNAKE_CASE

示例：
```text
MAX_FILE_LINES
DEFAULT_LIGHT_LEVEL
LOCK_TIMEOUT_MS
```

### 1.3 方法命名
- 方法名使用 lowerCamelCase
- 方法名必须表达动作或判断
- 查询方法优先使用 get / find / list / calculate
- 布尔判断方法优先使用 is / has / can / should
- 状态变更方法使用 set / update / transition / enter / exit

示例：
```text
classifyIntent()
rewritePrompt()
isDoorClosed()
canStartLearning()
transitionToLocked()
```

### 1.4 类命名
- 类名使用 UpperCamelCase
- 类名必须体现职责
- 状态机类以 StateMachine 结尾
- 服务类以 Service 结尾
- 控制器类以 Controller 结尾
- 策略类以 Policy / Strategy 结尾

示例：
```text
IntentClassifier
PromptRewriteService
DeviceStateMachine
ParentUnlockPolicy
LearningSessionController
```

### 1.5 接口命名
- 接口使用名词，不强制 I 前缀
- 如果项目内已有 I 前缀风格，统一保持，但不要混用

示例：
```text
DeviceRepository
LightController
DoorSensor
```

### 1.6 枚举命名
- 枚举类型使用 UpperCamelCase
- 枚举值使用 UPPER_SNAKE_CASE

示例：
```text
IntentType
- HOMEWORK_EXPLANATION
- VOCABULARY_HELP
- OPEN_GAME
```

### 1.7 文件命名
- 类文件使用类名（如 IntentClassifier.java）
- 模块文件使用 lower_snake_case（如 intent_classifier.py）
- 脚本文件使用 lower_snake_case（如 build.sh）

## 2. 文件规模限制

### 2.1 单文件行数限制
| 项目 | 推荐上限 | 超过后动作 |
| --- | ---: | --- |
| 单文件 | 300 行 | 必须评估拆分 |
| 单文件绝对上限 | 500 行 | 原则上禁止，需说明 |
| 单类 | 200 行 | 优先拆职责 |
| 单方法 | 30 行 | 超过需拆分 |
| 单方法绝对上限 | 50 行 | 原则上禁止 |
| 单个测试文件 | 300 行 | 按场景拆分 |

### 2.2 单个 PR 改动限制
| 项目 | 推荐上限 | 超过后动作 |
| --- | ---: | --- |
| 单个 PR 改动文件数 | 10 个以内优先 | 超过需说明 |
| 单个 PR 代码行数 | 300 行以内优先 | 超过需拆分或说明 |

### 2.3 超限拆分规则
如果一个文件过长，应按以下优先级拆：
1. 先按职责拆类
2. 再按功能拆文件
3. 再按层次拆模块
4. 测试按场景拆分

例如：
```text
IntentClassifier 太大
→ 拆成 AllowedIntentMatcher / ForbiddenIntentMatcher / ParentRequiredIntentMatcher
```

## 3. 注释规范

### 3.1 总原则
- 默认使用中文注释
- 注释解释"为什么"和"约束"，不重复代码表面意思
- 复杂规则、状态机、协议转换必须有注释
- 对外接口、核心类、核心方法必须有注释
- 禁止无意义注释

### 3.2 文件头注释
每个核心文件建议有文件头注释，内容包括：
1. 文件职责
2. 所属模块
3. 依赖的 PRD / ADR
4. 不应处理的职责边界

示例：
```text
本文件负责学习请求的意图分类，不负责提示词改写和回答审核。
相关文档：
- PRD-07-AI引导者与提示词过滤
- ADR-0003-受控AI访问策略
```

### 3.3 类注释
每个核心类必须说明：
- 这个类负责什么
- 不负责什么
- 关键输入输出
- 状态约束

### 3.4 方法注释
以下方法必须有注释：
- 核心业务方法
- 状态机迁移方法
- 协议解析方法
- 安全过滤方法
- 带副作用的方法
- 对外暴露的方法

方法注释至少说明：
1. 输入
2. 输出
3. 前置条件
4. 失败行为

### 3.5 行内注释
只在这些情况写：
- 业务规则不直观
- 与硬件时序有关
- 与安全边界有关
- 与协议兼容有关
- 临时 workaround 需要说明

### 3.6 TODO / FIXME 规范
统一格式：
```text
TODO(姓名或模块, 截止阶段): 说明原因和后续动作
FIXME(姓名或模块, 问题编号): 说明问题和影响
```

示例：
```text
TODO(ai, MVP2): 当前只支持基础意图分类，后续补充年级自适应
FIXME(firmware, FW-023): 某些断连场景下状态灯恢复延迟
```

### 3.7 禁止无意义注释
例如禁止：
```text
// 设置变量
// 判断是否为空
// 调用函数
```

## 4. 类与模块职责规范

### 4.1 单一职责
每个类、文件、模块尽量只承担一种明确职责。

不允许这种类：
```text
LearningAssistantManager
既做意图分类
又做提示词改写
又做 OCR
又做大模型调用
又做 TTS
又做日志
```

应该拆成：
```text
IntentClassifier
PromptRewriteService
OcrService
ModelGateway
AnswerReviewService
LearningLogService
```

### 4.2 状态机类
状态机类只做：
- 状态定义
- 状态迁移
- 状态合法性检查

状态机类不要直接做：
- UI 更新
- 网络请求
- 日志落盘

应通过回调、事件或依赖服务完成。

### 4.3 服务类
服务类做一类业务能力。

例如：
```text
PromptRewriteService 只改写提示词
AnswerReviewService 只审核回答
ParentAuthService 只做家长授权
```

### 4.4 控制器类
控制器类负责协调服务、状态机、UI，不做具体业务逻辑。

### 4.5 工具类限制
工具类只能放：
- 纯函数
- 无状态逻辑
- 格式转换
- 日期/字符串/协议辅助

禁止把业务规则随便塞进 Utils。

## 5. 代码风格补充

### 5.1 提前返回
优先使用提前返回减少嵌套层级。

### 5.2 减少嵌套层级
嵌套层级建议不超过 3 层，超过需重构。

### 5.3 禁止魔法数字
所有数字常量必须定义为具名常量并说明含义。

### 5.4 错误处理要求
- 不允许吞掉异常
- 关键错误必须记录日志
- 用户可见错误必须友好

### 5.5 日志输出要求
- 日志必须包含时间、模块、级别
- 日志内容要可追溯
- 敏感信息不能记日志
