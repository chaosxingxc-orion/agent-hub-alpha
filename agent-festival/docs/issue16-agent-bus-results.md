# Issue 16 Agent Bus 接入验收记录

2026-11-11，在 `feat/issue16-agent-bus` 完成执行计划的总线接入范围。基线提交 `5051d3a` 保存原工作区；实现提交 `7bb7f27`，带镜像/配置身份的验收脚本提交 `d17f304`；沿用 openJiuwen、WorkSwarm、Agent Bus 和 Loop Forest，新增内容为 Java 远程成员适配、Runtime 角色封装、Go 结果控制和部署/验收脚本。

**本记录验证授信模拟场景的三类专业服务，不代表公开年报场景或 Issue #16 全部完成。** W01/W06 的材料、指标计算、引用检查、年报报告验收尚未交接；#6 仍不影响启动。原型规则判断及关联方 Loop Forest 仍在宿主内，5 个普通专业节点由独立 Runtime 执行。

## 实际链路与版本

WorkSwarm 从受控目录生成候选图，用户确认后固定版本；Java Graph 调用 Gateway DIRECT，RDC 负责真实卡片和实例发现。Go 不选择物理地址、不推进业务图，只校验/保存调用与正式结果并生成界面投影。

| 项目 | 实际版本/行为 |
| --- | --- |
| agent-solution / RDC / Gateway | `bf702bc95064d3491cf8b374f54d271363d84230` / `0.1.3` |
| Core / Runtime | Maven `0.1.17` / `0.1.4` |
| A2A SDK / wire / 客户端 | `1.0.0.Final` / `1.0 JSON-RPC` / JDK 21 HttpClient 窄适配 |
| JDK / 编译目标 / Boot | 21 / 17 / 4.0.7 |
| WorkSwarm | `1142e5fb3984a2b1bc6352669a464976e204b2dc`；沿用原型可运行版本，与 #16 引用版本不同 |
| PostgreSQL | 15-alpine；镜像 digest 已固定在 `sources.lock.json` 和 Compose |
| 模型 | `deepseek-flash`，实际 provider 通过环境注入；凭据不进入源码/证据 |
| 卡片版本 / 协议 | 三类服务均 `1.0.1` / `JSONRPC 1.0`，非 streaming，DIRECT |
| 资料角色 | `credit-material`：intake |
| 分析角色 | `credit-analysis`：financial、industry、bureau；每调用隔离上下文和工具挂载 |
| 写作角色 | `credit-writer`：report |

未引入另一个注册中心、前端主壳、Go 调度器或 broker。卡片与运行任务由上游 Runtime 产生；新增客户端只映射 SendMessage/GetTask 与现有 Graph 成员接口。

## 验收结果

原始证据根目录：`~/.cache/agent-hub-alpha/issue16/evidence/agent-bus/`。同名 R1/R2 出现在不同 Compose 项目，须连同项目名核对。

| 验收 | 实验与实际结果 | 原始证据 |
| --- | --- | --- |
| BUS-01/03 | 主环境 R10、全新环境 R1 完成 5 个远端节点；3 类卡片真实可查；财务/行业执行时间重叠；上下文均不同 | `normal-1794397789.json`、`normal-1794396780.json` |
| BUS-02 | MCP 拒绝未知能力和 URL 形式目标；未授权完整征信计划预检失败；不产生业务执行；能力漂移和跨节点/角色/上下文/版本授权由控制层测试覆盖 | `protocol-admission-1794396891.json`、`go-race.jsonl` |
| BUS-04 | 相同节点换输入拒绝；相同结果幂等；不同结果、错误摘要/任务/上下文及旧宿主权限不能覆盖正式产物 | `go-race.jsonl`（控制层测试） |
| BUS-05 | 主环境 R11、全新环境 R2 在 intake 完成、分析已受理后强杀宿主；原图恢复、intake 复用；产物 ID/摘要及模型事件集合不变 | `host-kill-1794397901.json`、`host-kill-1794397010.json` |
| BUS-06 | R12 在 GetTask 断线；R13 延迟响应 35 秒超过客户端 30 秒超时；显示未知后由登录用户核对原任务，最终完成；代理记录每个业务节点恰好一份 SendMessage | `disconnect-1794397958.json`、`late-1794397997.json` |
| BUS-07 | 控制层覆盖派发意图/结果写入失败；R14 拦住已完成远端任务的回复，再让真实 state.json atomic rename 失败；不出现正式产物、不推进下游；恢复后查询原任务完成，模型记录不增加 | `storage-failure-1794398063.json`、`go-race.jsonl` |
| BUS-08 | R15 重启 Gateway；R16 重启资料 Runtime。原任务不可查询后保持未决；再次核对仍使用同一请求/任务，无第二次派发，下游均待执行 | `gateway-restart-1794398125.json`、`runtime-restart-1794398141.json` |
| BUS-09 | 真实缺鉴权 401；未知任务为 HTTP 200 + ROUTE_OWNERSHIP_UNKNOWN；错误消息/未知能力拒绝；非法节点凭据在 Runtime 失败；RDC 停机返回 503 RDC_UNAVAILABLE，目录显式不可用 | `protocol-admission-1794398171.json`（含 RDC 停机） |
| BUS-10 | 重复模型事件不重复计数；乱序状态不倒退；全新环境 Go/宿主重启后图与正式结果一致 | `go-race.jsonl`、`fresh-restart.json` |
| BUS-11 | 干净 `agent-hub-issue16-fresh` 项目和三个新卷启动，R1 真实全链路完成；缺模型配置启动前拒绝；5 个正式产物重启保留 | `fresh-start.log`、`normal-1794396780.json`、`missing-config.json`、`fresh-restart.json` |
| 对话入口 | 全新 WorkSwarm 的“授信编排助手”真实调用目录和计划工具，生成 7 节点 P4；预检通过，等待用户确认；随后以 alice 从页面批准，R3 完成 5 个远端节点 | `workswarm-candidate.json`、`workswarm-approved-execution.json`、`ui-approval.png`、`ui-completed.png` |

正常全新环境 R1 的节点模型尝试数：intake=2、financial=2、industry=2、bureau=2、report=1。这些是模型客户端调用前持久记录的事件，不是账单统计。恢复实验逐项比较事件集合和产物，不能用截图或报告文本相同代替“零增量”证明。

Go 可靠性测试新增 11 项，覆盖存储失败回滚、完成后重启复用、身份与契约绑定、重复/乱序事件、权限隔离、旧接口绕过及旧计划缺少总线绑定的拒绝。18 项 Go 测试全部通过 `go test -race`；`go vet`、Java/Maven 构建和图卡构建均有执行记录；实际 HTTP/A2A 故障覆盖另见上表，不能将单元测试说成真实网络实验。

最终复测覆盖上述全部故障路径，运行版本 `d17f304` 与实际镜像 ID/配置摘要写入每份复测 JSON 的 `environment`。已恢复正常 DIRECT 地址，移除故障代理；全新测试项目及其测试卷已清理，主环境保留原有用户数据。

## 保证和限制

- 调用意图先持久化再派发；正式结果按节点身份、输入摘要、尝试、任务/上下文校验，落盘成功后才返回 Graph。一个节点正式产物最多一份。
- 派发回执丢失且无远端任务号时保留未决，禁止推断未执行后重发。存在任务号时，恢复仅查询原任务。自动取消、自动重试和跨任务 exactly-once 均未承诺。
- 当前平面 Graph 的节点路径是已确认节点 ID、循环轮次为 0。没有将该实现宣称为通用嵌套子工作流/任意循环身份方案。
- Gateway 归属索引和 Runtime 默认任务存储在内存中；其重启导致的任务丢失需要人工核对。为此补持久化需作为后续上游能力接入，不能将宿主可恢复夸大成全链路高可用。
- 存储失败后服务冻结写入、健康检查 503。修复并重启后核对磁盘；既不假成功，也不自动重发未知调用。
- 静态凭据、演示账号、Docker 私网卡片发现及关闭卡片签名验证仅用于单机受控 Demo。浏览器代理只开放 UI/只读图/带登录批准，执行器和调试接口不经前端代理开放。
- 启用 Bus 后，缺少真实能力绑定的旧待确认计划会被拒绝，需重新生成并确认。新界面和提示中的原型“S1”标签已改为风险判断规约；历史已采纳图和截图保留原字段，不改写历史记录。
- 能力匹配用于选候选；执行仍以已确认图、目录契约及节点作用域为准。W06 年报业务核验、跨任务材料复用、暂停/反馈等 #16 其他要求不因本次通过而关闭。

## 调试失败记录

首次正常实验 `normal-1794395568.json`（主环境 R2）因 Gateway 租户属性名称错误停在未知，没有误判完成或自动重发；修正为上游实际读取的 `TENANTID/PRINCIPALID` 后 R3 通过。第一次新环境探测 `normal-1794396722.json` 发生在服务启动等待结束前，被目录可用性门槛拒绝；完整启动已增加三角色在线检查，之后 R1 通过。

页面和编译并发检查期间发生宿主内存不足；随后只停止本项目的测试 WorkSwarm，顺序重跑检查。未改宿主全局代理、Docker 设置或其他项目服务。测试过程的失败文件保留，不计作通过证据。

## 复跑与体验

操作命令、故障 overlay、登录后核对、全新部署、备份和回滚见[部署说明](../go/cmd/credit-demo/deploy/README.md)。主体验入口为 `http://localhost:5192`；正常链路样例为 `http://localhost:5192/credit/ui/run/R10`，完成节点复用样例为 `http://localhost:5192/credit/ui/run/R11`。默认演示账号 `alice`，密码由 `CREDIT_DEMO_PASSWORD` 配置（未配置时为 `demo`）。

源码配置、镜像 ID、各实验通过/失败状态及文件摘要汇总在[版本与证据索引](evidence/issue16-agent-bus.json)；脱敏原始证据归档在缓存根的 `agent-bus-evidence.tar.gz`。私有凭据、浏览器会话和原始数据备份不进入归档。

## 页面调用证据补充（2026-11-11）

已沿用同一个 WorkSwarm MCP 图卡渲染器，将执行证据接入确认页和运行页。图卡顶部显示 DIRECT、绑定服务数、远端节点数、已关联任务、已保存结果、待核对及复用数量；节点标明 Agent Bus、宿主内执行或宿主内 Loop Forest。点击节点或展开“调用清单”即可检查详情，清单也能查看尚未开始、暂未出现在运行图中的节点。

详情分为 RDC 能力绑定、Gateway DIRECT、专业 Agent 执行、正式结果四部分。新调用在派发意图落盘时保存 RDC 查询时间及候选实例；Java 宿主记录 SendMessage/GetTask 的请求标识、起止时间、HTTP 状态、协议结果和安全错误码。控制层在原有执行器 fencing、任务/上下文及事件顺序校验后持久保存观测，按 RPC ID 去重，向 UI 投影汇总。刷新页面仅读取已保存记录，不重新访问 RDC。

- “已保存绑定”表示候选计划确定的能力；待批准时不表示已经调用。
- 候选实例健康值属于当时快照。**实际执行实例未采集**，不将候选实例冒充实际执行目标。
- Gateway 请求/回执时间来自 **Java 宿主观测**，不宣称是 Gateway 服务端追踪。HTTP 200 + 协议错误保持待核对，只有校验通过的任务回执才能计入有效回执。
- 历史调用缺少的快照和时刻显示“未采集”，不使用当前数据补写历史。恢复查询只增加原任务查询记录；结果复用保留原调用证据。
- 关联标识和摘要支持展开、复制；产物展示为图模型提供的内容摘要。复制及 API 投影均不包含令牌、认证头、路由句柄或完整远端输入。

本次真实样例：`http://localhost:5192/credit/ui/run/R18`，5 个远端节点经 3 类服务完成，各有 RDC 快照、一次派发观测、有效任务回执及正式产物。待批准样例为 `http://localhost:5192/credit/ui/plan/P22/v1`。对原故障样例 R16 再次核对，观测到 GetTask 的 HTTP 200 协议错误；原 request/task 与模型事件集合不变，仍处于等待状态。R10 保持历史观测缺失，R11 保持原产物复用展示。

验证：23 项 Go 测试通过 `go test -race` 和 `go vet`；5 项前端语义测试及 TypeScript 检查通过；Java/Maven 与前端、Go 镜像构建通过。浏览器验证覆盖桌面、390px 窄屏、复制、键盘操作、待批准/完成/复用/未决状态和刷新保留展开状态。MCP 嵌入协议测试验证初始化、工具结果通知、详情展开后的高度上报及数据变化时保留关联标识展开；这项测试使用宿主页测试夹具，未将其计作一次新 WorkSwarm 对话验收。

页面检查时主机内存紧张，前两次浏览器导航超时；暂停本项目 WorkSwarm 后通过检查，最终恢复。运行任务 R18 已在暂停前完成，未重启总线或 Runtime。新增证据与截图独立保存在 `~/.cache/agent-hub-alpha/issue16/evidence/agent-bus-ui/`，索引见[页面证据索引](evidence/issue16-agent-bus-ui.json)，不覆盖上一轮故障验收归档。
