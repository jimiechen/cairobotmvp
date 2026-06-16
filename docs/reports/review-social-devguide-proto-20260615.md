# Social Service 开发规范 + Proto 文件评审意见

> **评审日期**：2026-06-15
> **评审对象**：`social-service-dev-guide.md` + 6 个 proto 文件（message/user_base/group_base/topic_base/third_base/inbox_base）
> **评审视角**：与已产出 PRD/ADR/OpenAPI/协议注册表的一致性、MVP 可落地性、工程规范合规性
> **关联文档**：PRD-social-app-mvp.md / ADR-social-data-level-and-cache-strategy.md / social-openapi.yaml / 协议编号注册表.md

---

## 一、评审结论

| 维度 | 评价 | 说明 |
|---|---|---|
| 7 层架构设计 | **优秀** | Gateway→Invoker→Servant→Handler→Service→Repo+Model 分层清晰，职责边界明确 |
| 三条铁律 | **通过** | 一协议一 svc 文件 / proto 冻冻 / model 独立 proto，三条规则合理且必要 |
| Permission Service | **通过** | 7 个能力方法 + 广场虚拟成员特化，设计完整 |
| 开发顺序（12 步） | **通过** | 从 proto→model→repo→permission→servant→handler→svc→test，顺序合理 |
| **proto 与 PRD 的 minType 不一致** | **必须修正** | proto 文件是线上在用定义，dev guide 引用的 minType 与 proto 完全对齐，但与我们 PRD/注册表的设计编号不同 |
| **model.go 残留关注字段** | **必须修正** | MemberStats 仍有 followers_count/following_count，与"移除关注模型"决策矛盾 |
| **域范围超出 MVP** | **需确认** | dev guide 包含 5 个域（member/group/topic/third/inbox），PRD 只覆盖前 3 个 |
| **proto package 路径** | **需确认** | 使用 `github.com/jimiechen/mineplanet` 前缀，需确认是否迁移到 cairobotmvp 组织 |

---

## 二、核心问题逐项分析

### 2.1 问题 P0：Proto 实际 minType 与 PRD/注册表设计的编号体系不一致

这是最严重的问题。**proto 文件是线上产品正在使用的真实协议定义**，而我们的 PRD/注册表是重新设计的理想化编号。两者存在系统性偏差。

#### 成员域（maxType=1000）对比

| 功能 | Proto 实际 minType (Request) | Dev Guide 引用 | PRD 设计 | 注册表 | 差异 |
|---|---|---|---|---|---|
| 用户注册 | **1021** | 1021 | 1001 | 已删(改后) | **差 20** |
| 用户登录 | **1023** | 1023 | 1003 | 已删 | **差 20** |
| 用户登出 | **1025** | 1025 | 1005 | 已删 | **差 20** |
| 刷新令牌 | **1027** | 1027 | 1007 | 已删 | **差 20** |
| 获取用户信息 | **1029** | — | 1009 | 已删 | **差 20** |
| 更新用户信息 | **1031** | — | 1010 | 已删 | **差 21** |
| 拉黑用户 | **1039** | 1039 | 1011 | 已删 | **差 28** |
| 解除拉黑 | **1041** | 1041 | 1013 | 已删 | **差 28** |
| 黑名单列表 | **1043** | 1043 | 1015 | 已删 | **差 28** |
| 拉黑数量 | **1047** | — | 无 | — | PRD 缺失 |
| 批量获取用户信息 | **1049** | — | 无 | — | PRD 缺失 |
| 用户统计 | **1045/1046** | — | 已移除 | 已删 | **proto 仍存在** |
| 会员升级 | **1051/1052** | — | 无 | — | PRD 缺失 |
| IM 签名 | **1074/1075** | — | 无 | — | PRD 缺失 |
| 用户配置 | **1091-1100** | — | 无 | — | PRD 缺失（10 个协议） |
| 通知设置 | **1053-1056** | — | 无 | — | PRD 缺失（4 个协议） |

**关键发现**：
1. Proto 的成员域从 **1021 开始**（跳过了 1001-1019），说明 1000 段的前 20 个编号已被占用或有历史原因
2. 我们的 PRD/注册表从 **1001 开始**设计，与实际 proto 完全错位
3. Proto 中有大量我们 PRD 未覆盖的协议：IM签名(1074)、用户配置(1091-1100)、通知设置(1053)、会员升级(1051)

#### 群组域（maxType=2000）对比

| 功能 | Proto 实际 minType | PRD 设计 | 差异 |
|---|---|---|---|
| 创建群组 | **2005/2006** | 2001/2002 | 差 4 |
| 更新群组 | **2009/2010** | 2003/2004 | 差 6 |
| 删除群组 | **2011/2012** | 2005/2006 | 差 6 |
| 加入群组 | **2013/2014** | 2011/2012 | 差 2 |
| 退出群组 | **2015/2016** | 2013/2014 | 差 2 |
| 禁言成员 | **2019/2020** | 2015/2016 | 差 4 |
| 解除禁言 | **2021/2022** | 2017/2018 | 差 4 |
| 封禁成员 | **2023/2024** | 2019/2020 | 差 4 |
| 解除封禁 | **2025/2026** | 2021/2022 | 差 4 |
| 移除成员 | **2027/2028** | 2023/2024 | 差 4 |
| 修改角色 | **2029/2030** | 2025/2026 | 差 4 |
| 续费成员 | **2037/2038** | 2017/2018 | 差 20 |
| 进入圈子 | **2087/2088** | 2007/2008 | 差 80 |
| 权限查询 | **2059/2060** | 2027/2028 | 差 32 |
| 角色权限修改 | **2055/2056** | 2029/2030 | 差 26 |
| 快速设置 | **2057/2058** | 2031/2032 | 差 26 |
| 折扣配置 | **2067/2068** | 2033/2034 | 差 34 |
| 应付金额 | **2073/2074** | 2035/2036 | 差 38 |
| 成员配额 | **2089/2090** | 无 | PRD 缺失 |
| 统计刷新 | **2041/2042** | 无 | PRD 缺失 |
| 名称检查 | **2051/2052** | 无 | PRD 缺失 |
| Slug 检查 | **2053/2054** | 无 | PRD 缺失 |

**同样存在系统性偏移**：Proto 编号普遍比 PRD 设计大 4~80 不等。

#### 主题域（maxType=3000）对比

| 功能 | Proto 实际 minType | PRD 设计 | 差异 |
|---|---|---|---|
| 创建主题 | **3001/3002** | 3001/3002 | **一致** |
| 主题列表 | **3005/3006** | 3003/3004 | 差 2 |
| 删除主题 | **3009/3010** | 3005/3006 | 差 4 |
| 点赞主题 | **3061/3062** | 3007/3008 | 差 54 |
| 收藏主题 | **3063/3064** | 3009/3010 | 差 54 |
| 置顶主题 | **3029/3030** | 3011/3012 | 差 18 |
| 搜索主题 | **3049/3050** | 3013/3014 | 差 36 |
| 添加回复 | **3043/3044** | 3015/3016 | 差 28 |
| 回复列表 | **3065/3066** | 3017/3018 | 差 48 |
| 删除回复 | **3055/3056** | 3019/3020 | 差 36 |
| 点赞回复 | **3077/3078** | 3021/3022 | 差 56 |
| 置顶评论 | **3081/3082** | 3023/3024 | 差 58 |
| 精选主题 | **3091/3092** | 3025/3026 | 巹 66 |
| 举报类型 | **3093/3094** | 3027/3028 | 差 66 |
| 提交举报 | **3095/3096** | 3029/3030 | 差 66 |
| 操作权限查询 | **3099/3100** | 3031/3032 | 差 68 |
| 用户状态检查 | **3083/3084** | 3033/3034 | 差 50 |
| Topic IDs | **3085/3086** | 无 | PRD 缺失 |
| 批量 Topic | **3057/3058** | 无 | PRD 缺失 |
| TopicStat | **3035/3036** | 无 | PRD 缺失 |
| Refresh Stat | **3037/3038** | 无 | PRD 缺失 |

**只有 CreateTopic(3001) 是一致的**，其余全部有偏差。

#### 根本原因分析

Proto 文件的编号不是按"连续奇偶分配"的规则生成的，而是**按上线时间顺序递增分配**的。中间存在大量被跳过或废弃的编号（如 reserved 字段注释所示）。我们的 PRD/注册表采用了理想化的"从 N001 开始连续分配"策略，但这与实际 proto 定义冲突。

---

### 2.2 问题 P0：model.go 残留 followers_count / following_count

[dev-guide.md §9 model.go](docs/tabbit/inbox/2026/06/protocols/social-service-dev-guide.md#L711-L722) 中的 `MemberStats` 结构体：

```go
type MemberStats struct {
    UserID         string `gorm:"primaryKey;type:char(32);column:user_id"`
    TopicsCount    int64  `gorm:"default:0;column:topics_count"`
    GroupsCount    int64  `gorm:"default:0;column:groups_count"`
    FollowersCount int64  `gorm:"default:0;column:followers_count"`    // ← 应删除
    FollowingCount int64  `gorm:"default:0;column:following_count"`    // ← 应删除
    UpdatedAt      int64  `gorm:"column:updated_at"`
}
```

同时 [user_base.proto §UserStats](docs/tabbit/inbox/2026/06/protocols/base/user_base.proto#L377-L391) 也包含：

```protobuf
message UserStats {
    string user_id = 1;
    int32 topics_count = 2;
    int32 comments_count = 3;
    int32 likes_given = 4;
    int32 likes_received = 5;
    int32 groups_count = 6;
    // ...
}
```

注意：**proto 的 UserStats 没有 followers/following 字段**（已移除），但 dev guide 的 Go model 还保留着。这是 proto 冻冻后清理了字段但 dev guide 模板未同步的结果。

---

### 2.3 问题 P1：Dev Guide 域范围 vs PRD 覆盖范围

| 域 | maxType | Dev Guide | Proto 文件 | PRD 覆盖 | 状态 |
|---|---|---|---|---|---|
| member | 1000 | 有 | user_base.proto | 部分 | **PRD 遗漏 IM/配置/通知/会员升级** |
| group | 2000 | 有 | group_base.proto | 部分 | **PRD 遗漏进入圈子/折扣/应付/配额/权限lite** |
| topic | 3000 | 有 | topic_base.proto | 部分 | **PRD 遗漏状态检查/TopicIDs/批量/刷新/举报/操作权限** |
| third | 4000 | 有 | third_base.proto | **无** | PRD 完全未覆盖 |
| inbox | 5000 | 有 | inbox_base.proto | **无** | PRD 完全未覆盖 |

**third 和 inbox 两个域在 dev guide 中有完整的目录结构、servant/handler/svc 模板和路由配置，但 PRD 完全没有涉及。**

---

### 2.4 问题 P1：Permission Service 中的广场虚拟成员概念

Dev guide §10 引入了 **广场虚拟成员（Plaza Virtual Membership）** 概念：

```go
// 广场群组特化规则:
//   isPlazaGroup == true:
//     普通成员  →  active 用户即具备，不查 group_members
//     管理员/嘉宾  →  从 group_members 查 role = admin/guest
```

这个概念在以下位置出现：
- permission/service.go 完整实现（§10，约 120 行）
- 数据等级规范表（§12）：广场普通成员 / 广场 admin/guest 各自独立规则
- 测试维度表（§14）：广场虚拟成员 / 广场 ban 成员 为必测项
- 禁止行为清单（§15）：广场普通成员写入 group_members 为禁止行为

**问题**：这个概念在我们的 PRD-social-app-mvp.md、ADR-social-data-level-and-cache-strategy.md、social-openapi.yaml 中**均未出现**。这是一个重要的架构决策（属于 ADR 级别），不应只存在于 dev guide 中。

---

### 2.5 问题 P2：Proto Package 路径与项目组织不匹配

所有 proto 文件使用：
```
go_package = "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base";
package com.mineplanet.pojo.{user|group|topic|third|inbox};
```

当前项目是 `cairobotmvp`，路径应调整为项目的 go module 路径。这影响：
- `make proto` 生成的 Go 代码导入路径
- svc_*.go 中的 pb import 路径（dev guide §6 handler.go 已写死为 `github.com/jimiechen/mineplanet/...`）

---

## 三、通过项（无需修改）

| # | 项 | 评价 |
|---|---|---|
| 1 | **7 层架构图**（§2） | 清晰准确，与 CaiRobot MVP Gateway→TarsInvoker→Servant 链路完全一致 |
| 2 | **Rule 1: 一协议一 svc 文件** | 正确。防止 vibecoding，便于 code review 和独立测试 |
| 3 | **Rule 2: proto 冻结** | 正确。内部 model 解耦是标准做法 |
| 4 | **Rule 3: model 独立于 proto** | 正确。避免 proto 变更波及 DB schema |
| 5 | **routes.yaml 按 maxType 分 Servant**（§4） | 正确。minType 通过 extend 注入，不在 routes.yaml 拆分 |
| 6 | **servant.go 模板**（§5） | 标准 TarsGo bytes 接口，职责限定清晰 |
| 7 | **handler.go dispatchProto 泛型函数**（§6） | 用 Go 泛型消除 case 内重复代码，设计优雅 |
| 8 | **svc_*.go 五步法**（§7） | 校验→权限→写库→事件→响应，流程标准化 |
| 9 | **repository 接口定义**（§8） | 接口粒度合理，1级/2级分离 |
| 10 | **Mock Repository 测试模式**（§14） | 正确。Mock 隔离 DB，单元测试可运行 |
| 11 | **禁止行为清单**（§15） | 17 条禁止项覆盖全面，每条都有违反规则和后果 |
| 12 | **缓存 Key 规范表**（§12） | 与 ADR 中的 Cache Aside 策略一致 |
| 13 | **数据等级强制约束表**（§12） | 1级/2级读写规则明确，铁律标注清晰 |
| 14 | **开发顺序 12 步**（§13） | 依赖关系正确，批次划分合理 |
| 15 | **测试命名规范**（§14） | 中文场景描述 + 英文函数名，符合项目编码规范 |
| 16 | **GroupMemberRole 枚举**（user_base.proto L90-98） | owner/admin/member/guest/banned/visitor 六级清晰 |
| 17 | **MemberInfo 合并消息**（group_base.proto L163-188） | 将 GroupMember + GroupMemberView 合并为统一 MemberInfo，减少接口数量 |
| 18 | **TopicNavType 导航枚举**（topic_base.proto L30-40） | 支持最新/问答/精华/媒体等导航类型，与产品形态匹配 |
| 19 | **CheckTopicActions 权限查询**（topic_base.proto L824-850） | 返回 available_actions + state 快照，减少客户端额外请求 |
| 20 | **inbox_base.proto 私信支持** | 支持私信发送(5061)/未读(5063-5066)，收件箱功能完整 |

---

## 四、必须修改项汇总

| # | 问题 | 影响 | 建议 | 优先级 |
|---|---|---|---|---|
| 1 | **PRD/注册表的 minType 与 proto 实际编号不一致** | 开发时无法对号入座 | **方案 A（推荐）**：以 proto 文件为基准，重写 PRD §8 和注册表 §4，使用 proto 实际编号；**方案 B**：保持 PRD 理想编号，在注册表中增加"proto 实际编号"列做映射 | **P0** |
| 2 | **model.go MemberStats 残留 followers/following 字段** | 与"移除关注模型"决策矛盾，生成多余 DDL | 删除 FollowersCount/FollowingCount 两行 | **P0** |
| 3 | **缺少 ADR-plaza-virtual-membership** | 广场虚拟成员是重要架构决策，散落在 dev guide 中 | 新建 `docs/adr/ADR-plaza-virtual-membership.md`，从 dev guide §10 抽取 | **P1** |
| 4 | **third/inbox 域无 PRD 覆盖** | dev guide 有完整目录结构但无需求文档 | 补充 third/inbox 的最小 PRD 或明确标注为 P2 | **P1** |
| 5 | **proto package 路径需迁移** | import 路径指向旧组织 | 全部替换为 `github.com/cairobotmvp/...` 或项目实际 go mod 路径 | **P1** |
| 6 | **handler.go import 路径硬编码** | dev guide §6 写死 `github.com/jimiechen/mineplanet` | 替换为项目实际路径 | **P1** |

---

## 五、建议方案：以 Proto 为基准重构编号体系

鉴于 proto 文件是**线上在用的真实协议定义**，建议采用以下策略：

### 方案：Proto-First（推荐）

**原则**：以 proto 文件中已有的 minType 为唯一事实来源（Single Source of Truth），PRD 和注册表向其对齐。

**操作步骤**：

1. **遍历所有 6 个 proto 文件**，提取每个 Request/Response 的 (maxType, minType, messageName) 三元组
2. **更新协议编号注册表 §4**：将 member/group/topic/third/inbox 五个域的条目全部替换为 proto 实际值
3. **更新 PRD §8 协议组表格**：使用 proto 实际 minType
4. **更新 social-openapi.yaml**：使用 proto 实际 minType
5. **更新 dev guide 目录结构中的 minType 注释**：确保与 proto 一致
6. **新增"编号映射说明"**：解释为什么某些编号区间有跳跃（reserved/废弃/第三方预留）

**优点**：
- 开发人员打开 proto 文件和 dev guide 看到的是同一个编号
- 不需要维护两套编号体系的映射关系
- `make proto-check` 可以直接验证一致性

**缺点**：
- 编号不连续（如 1021, 1023, 1025... 跳过了偶数 Response 编号后的下一个 Request）
- 需要在注册表中标注哪些编号是废弃的

---

## 六、Proto 文件清单与协议数量统计

| Proto 文件 | maxType | Request 数 | Response 数 | 总消息数 | 枚举数 | 复合类型数 |
|---|---|---|---|---|---|---|
| message.proto | — | 1 (MessagePacket) | — | 1 | 1 (Platform) | 0 |
| user_base.proto | 1000 | ~22 | ~22 | ~44 | 7 | 6 |
| group_base.proto | 2000 | ~28 | ~28 | ~56 | 10 | 9 |
| topic_base.proto | 3000 | ~24 | ~24 | ~48 | 11 | 7 |
| third_base.proto | 4000 | ~25 | ~25 | ~50 | 5 | 6 |
| inbox_base.proto | 5000 | ~11 | ~11 | ~22 | 4 | 4 |
| **合计** | **—** | **~111** | **~111** | **~222** | **38** | **32** |

**对比我们的 PRD 设计**：
- PRD 设计：17 对协议（34 个编号）
- Proto 实际：约 **55 对协议（111 个编号）**，仅 member+group+topic 三域
- 加上 third + inbox：约 **66 对协议（132 个编号）**

**差距原因**：我们的 PRD 只覆盖了核心 MVP 子集，遗漏了大量已有协议（IM 签名、用户配置、通知设置、分享追踪、钱包、开屏广告、OAuth/Passkey/2FA、举报、操作权限查询等）。

---

## 七、总结建议

### 立即执行（本次任务）

1. **决策**：是否采用 "Proto-First" 方案重构编号体系？还是保持 PRD 理想编号并增加映射层？
2. **修正**：dev guide model.go 删除 followers_count/following_count
3. **新建**：ADR-plaza-virtual-membership.md（从 dev guide §10 抽取）

### 后续迭代（P1/P2）

4. **补充**：third 域（OSS/OAuth/Passkey/2FA/分享/钱包）的最小 PRD
5. **补充**：inbox 域（收件箱/私信/IM 回调）的最小 PRD
6. **迁移**：proto package 路径到项目实际 go mod
7. **对齐**：所有文档的 minType 以 proto 文件为唯一基准

---

*评审完成。以上意见基于"dev guide + proto 文件全集 vs 已产出 PRD/ADR/OpenAPI/注册表"的一致性视角输出，供项目主控决策参考。*
