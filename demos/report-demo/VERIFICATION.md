# Demo 验证报告

本轮内部收敛基于已发布接收基线 `0358acbad0ce1b0ce73772f31eda845057ba4dc5`，分支 `codex/report-demo-w03-w04`。下面分开记录内部收敛验证与既有接收基线；接收方已经固定的0358acb源码/制品仍可取得，不把本轮整理视为新的联合验收。运行入口见 [LOCAL-RUN](LOCAL-RUN.md)，接口见 [contracts](contracts/README.md)，依赖见 [source-manifest](source-manifest.json)。

## 内部收敛验证（2026-11-17，基于0358acb）

- 裁剪 W02 年报投影不会产生的授信结论、effect/blackbox 分组及旧版本展示分支；保留图采纳、恢复、BLOCKED、HUMAN、Bus详情与报告操作。
- Java 故障测试必须显式指定本环境 `REPORT_FAULT_SOURCE_TASK`；删除没有当前调用方的一次性自然修订历史核对脚本，原脚本仍可在已发布固定提交取得。
- 操作文档使用本环境任务参数，统一已有真实浏览器验收入口；历史实验记录引用固定链接，公开接收 JSON 的字段与内容保持原样。
- Go控制、公开HTTP/MCP/DTO/schema、Java主代码、专家/身份/Bus装配、跨任务复用、隔离效果接口、旧MCP图卡和备份恢复均未改变。执行绑定算法保持不变。

验证环境：Windows。

- 前端类型检查、构建、7项Node测试及4项构建页面浏览器测试通过，0失败/跳过。
- 以真实W04任务投影构造5种快照，对照0358acb源代码，节点位置、数据边、状态和条件标签一致；新构建页另以受控只读RPC显示8节点、DELIVERED和Bus详情，6项检查通过、命令数0。它不是新的真实WorkSwarm端到端或HUMAN业务验收。
- Java编排30项（含新的显式源任务检查及13个跨进程窗口），0 failure/error/skip；故障提供方仍为隔离HTTP，不冒称真实Bus故障全覆盖。
- Go导出契约一致性测试通过；Go/API/schema、Java主代码、专家、部署脚本、来源清单和公开接收JSON均与0358acb一致。实际execution_profile与model_binding_digest均保持原值。
- 本轮没有新模型调用、没有服务重启，没有重跑Linux全链；不将历史主链当作本地改动后的新模型运行，不声明外部接收或业务签认完成。测试日志和截图留在私有`.local/report-demo/evidence/cleanup-*`。

## 已发布接收基线

## HUMAN 入口与 Bus 环境绑定增量：2026-11-17

插件对当前有效、尚未处理的 HUMAN 检查显示 Go 可信处理页链接；仍使用独立登录、会话/CSRF 和检查/报告对象绑定审核。人工处理后刷新投影移除待审提示，保留原始 HUMAN 检查；没有把完整审核表单复制到插件。

Task/Binding 固定 `environment_id/tenant/issuer/audience/caller_subject`，进入输入摘要。Go 在采纳、解析/提交、发送意图前检查；Java 在发现、发送和恢复查询前检查。失配返回 `BUS_BINDING_CHANGED`，节点进入 BLOCKED；历史缺绑定记录只读，不从当前配置补齐。逻辑身份不包含 Gateway/RDC 地址和秘密，支持同环境代理及凭据轮换。

| 实际检查 | 结果及边界 |
| --- | --- |
| Windows 完整 Go/真实 PG | 49 顶层、101 含子用例，0失败、0跳过测试；vet通过。新增换租户/环境/身份域、缺绑定历史记录、零发送意图、原记录不变反例；补测 HUMAN HTTP 投影处理前后变化 |
| Windows Java | 29 项，0 failure/error/skip。真实 Go 进程/PG + 隔离 HTTP 提供方：发送前及旧任务查询前换环境零外呼；同环境新代理只 GetTask、无第二次 SendMessage，原 call/remote task 不变 |
| Windows 插件 | 类型检查、构建、5项 Node 测试和4项实际构建页面浏览器测试通过。浏览器使用模拟 RPC，证明提示、链接和处理后投影刷新；不称真实 WorkSwarm HUMAN 业务验收 |
| Linux 原生针对性回归 | Go 3顶层/13含子用例、Java协议与跨语言10项、插件类型检查/构建/5项测试通过。未重新跑本增量的 Linux 完整 PG 或真实模型主链 |
| Windows 更新后真实模型主链 | 新对话规划→插件确认→4专家/5次远端调用→一次定向修订→最终程序PASS→插件交付/原字节下载，8节点全部正式提交。规划6项、主链10项、PR25身份/对账20项核对通过 |

真实新任务 `T-753df58e598a9aa5244ea5e374cedb9a`，状态 DELIVERED；最终报告 `sha256:92a3164634b20bf385a07dc061150d6e5e8f2150bcf5e816d79747f9e4def79c`。任务中固定的 Bus 绑定与发布配置相同，原 delivery_id 重放返回同一 receipt。实际模型结果与增量制品摘要见脱敏样例的 `bus_environment_and_human_validation`；上一轮双平台完整主链及制品保留为下节固定基线，不能将它们标为本增量的 Linux 全链证据。独立业务签认、真实 HUMAN 全链和外部协作者复跑仍未由这些检查证明。

## 基线双平台结果：2026-11-17（f657d9f）

实现提交 `f657d9fa31eedb39b5c9fadcd09a928a0c5ab790`。同一套 Go/Java/Python/WorkSwarm 插件代码，分别在 Windows 和本机 WSL2 的独立 Linux 工作区构建、启动和执行真实模型任务。浏览器客户端均为 Windows Edge；Linux 验证的后端、静态前端构建和服务均由 Linux 原生进程承担，未借用 Windows exe。测试前停止 Windows 实例，并核对 Windows 客户端与 WSL 内部返回同一 Go control_session_id，排除误连旧服务。

| 项目 | Windows | Linux / WSL2 |
| --- | --- | --- |
| 环境 | Windows x64；Oracle JDK21.0.12；Python3.14.6 | Ubuntu26.04.1 x86_64；OpenJDK21.0.13；Python3.14.4 |
| 共用工具/数据库 | Go1.25.0、Maven3.9.16、Node24.22.1、PostgreSQL16.16 | 同左；独立数据库/运行目录，原生工具链 |
| 新任务 | `T-30fcc5df8694b705c1b80a1d14eaa740` | `T-10a3e5de04d794c8e6e18c59a6d7b294` |
| 新对话 | `web_1a1e984e803_96636623157b` | `web_1a1e976619f_867fa611a064` |
| 主链 | 新模型规划→插件确认→4专家/4远端task→初检PASS→批准→DELIVERED→原字节下载 | 同左 |
| 正式节点/有条件跳过 | 6 / 2；未发生修订，不把跳过节点记为正式结果 | 6 / 2；未发生修订 |
| 主链/规划只读核对 | 10 / 6 项通过 | 10 / 6 项通过 |
| PR25身份与对账核对 | 16项通过，6个正式结果有结构化身份 | 同左 |
| Go/真实PG | 47顶层、93含子用例；0失败、0跳过测试；vet通过 | 同左 |
| Java编排/专家 | 26 / 6；0 failure/error/skip | 同左 |
| 服务/真实身份检查 | 16项就绪、6项身份正反例通过 | 同左 |

Go日志另外包含selection包的`[no test files]` package事件；这是无测试包，不计为跳过的测试。Java故障窗口使用真实Go进程/PG和隔离HTTP提供方，不计为真实业务模型调用。两端真实主链和隔离故障证据分别记录。

Windows报告摘要 `sha256:88bbe30afcd47c87d28523c816ff0ca663e5b49d1692e1186e90c9c1257c33bc`；Linux报告摘要 `sha256:9332d816d604d632213a5826e15cab61b5cc9c5e61c75ab048b2d6efbb313735`。摘要最终以[脱敏样例](evidence/integration-sample.json)的`current_platform_validation`实际字段为准。两端分别保留原始下载字节、任务、调用、观察和规划工具记录；相同delivery_id重复交付返回原receipt_id，主链10项独立核对通过。

两端直接启动实际专家JAR，模型名漂移、预算漂移、缺绑定三例均在模型出口前拒绝，本地计数出口请求为0；4个跨语言预算golden向量通过。Core真实客户端测试验证所校验的max_tokens确实进入请求。该隔离出口测试不代替真实Bus上共同M1—M4验收。

实际Go重启测试验证新control_session、相同正式投影、原call/task关联、审计游标无重复和产物完整性。它不把内存Runtime重启后的远端任务恢复推定为已完成；原有故障用例和尚缺组合继续单列。

### 本轮发现并处理的问题

1. Windows先行任务`T-3b5d94d4c2bc6d19f232053df41ddb0b`的写作JSON在6000-token上限处截断，按预期FAILED，未提交半份报告。保留失败记录；默认预算改为12000，允许受限环境覆盖并进入真实模型绑定，写作输出改为紧凑格式。旧任务不改写，重新发布后用新任务验证。
2. 原浏览器测试按折叠区内不可见按钮定位报告，曾在一个已PASS任务上超时；改为用权威artifact_id定位details后展开。最终上表Windows/Linux任务均完成完整单次浏览器主链；另保留原浏览器断开后重连原任务成功的记录。
3. 全新Linux工作区的模型配置教程遮挡已安装助手；Web/Gateway缺少环境别名时也不能读取环境模型目录。安装器只关闭隔离工作区教程，启动器向三个WorkSwarm进程传同一环境模型引用，不将实际密钥写入配置或源码。
4. Gateway已知身份错误、越权、依赖不可用不再全部折叠。真实Go/PG集成验证AUTH_EXPIRED→WAITING_AUTH、CALLER_NOT_AUTHORIZED→BLOCKED、RDC_UNAVAILABLE→WAITING_REMOTE/UNKNOWN，保留分类且已有意图不盲目重发；失败自由文本不进入观察投影。
5. 对话验证器原先假设同一selection只能出现一次提案，误拒绝“先被拒再修正”的真实记录。现在必须同时匹配最终图、tool_call_id、request_id和成功回执的taskId；保留失败提案，不把它计为成功。

### 来源、可复跑与接收状态

Windows步骤见[LOCAL-RUN](LOCAL-RUN.md)，Linux见[LINUX-RUN](LINUX-RUN.md)，机器码和预算绑定见[契约说明](contracts/README.md)。实际依赖解析和平台JAR/Go哈希写入接收样例；不同平台JAR摘要不同，不声称逐字节可重现构建。四专家共用一个JAR、各自角色配置和RDC登记独立。Keycloak归档校验值已与官方发行资产匹配。

W02/W05接收人已确认；W01 ae17ee3已有人工核对，核对独立性由负责人裁定。W06/W07完整业务签认、另一协作者部署、实际执行实例独立证明、GLM推理模型适配、完整M1—M4/人工UNKNOWN恢复的联合验收仍未完成。PR36的独立Root/批准/状态服务未接入；业务检查输入、完整报告摘要和版本组合需共同确认。不得把本机双平台通过写成上述外部门禁关闭。

当前保留Windows服务供本机访问；WSL验证完成后正常停止其工作区服务，源码、工具、配置与持久数据保留，可按Linux运行文档重启。API密钥仍只通过获准进程环境使用。

## 历史实验与原覆盖审计

早期逐轮实验、失败记录、专项验证及原方案 T01—T29 的逐项覆盖审计，完整保留在 [0358acb 固定验证报告](https://github.com/yougq/agent-hub-alpha/blob/0358acbad0ce1b0ce73772f31eda845057ba4dc5/demos/report-demo/VERIFICATION.md)。原接收样例及 Windows/Linux 制品摘要仍在 [integration-sample.json](evidence/integration-sample.json)，本轮不修改其既有字段和记录。历史参考不自动升级为当前共同验收。

明确保留以下边界与待办：

- W02/W05另一环境独立接收、真实等待后的人工恢复、完整M1—M4及其它故障窗口组合，仍需在共同版本执行；隔离HTTP测试不替代真实Gateway/Runtime故障验收。
- 真实身份正反例不等于全部权限/秘密隔离；多用户产物范围授权、同OS用户隔离及完整旁路审计尚未完成。
- W01核对独立性由负责人裁定；程序PASS、测试账号批准/DELIVERED和业务签认分别记录，W06/W07业务接收未完成。
- 当前有限目录/配置排序不宣称语义召回；真实Builder导出和场景替换验收未完成，PR34远端跨任务知识快路径后置。
- 已有跨任务提取复用、来源撤回与隔离效果实验维持原实现/已测范围，不因本轮整理移除或扩大其能力声明。
- W04备份包含正式表和不可变内容；不等于整个RDC/IdP/WorkSwarm环境灾备。内存Runtime重启可能丢远端task，UNKNOWN不盲重发。
- 当前scope短事务门偏向正确性；全并发、长稳、连接池压力、清理竞态、全部IPC/流式中断，以及生产SLA尚无完整结论。
- 真实实际执行实例的独立证明、GLM推理模型适配、共享版本与最终展示安排仍由接收方共同确认。
