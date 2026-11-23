# 有限节点字段消费：私有接线候选 v1

本候选落实已批准的受控工程接线，仍待 Hub 上游接收。它不是 SDK、A2A 或 Bus 公共协议升级；原 Java 调度器、原 Begin/Commit、原 Plan/Run/Call/StepRec 继续拥有执行和正式结果。

## 采纳与字段约束

新 profile 为 `two-role-identity/v1`。服务器固定 `hub-node-profile/v1`，在原 eligible 硬筛选阶段检查类型、单位、期间、口径与材料来源；未知不入合格排序，已知不相容排除，保留全部候选分母和原因。原评分、阈值、选用与组图算法不重写。高分不相容候选不能盖过较低分的兼容候选。

可选 `Plan.execution_binding` 使用 `hub-node-binding/v1`：profile、完整场景快照、材料 pins、node_contracts 和 input_bindings。私有 Go 采纳闭包另外固定 `bus_tenant`，防止配置变化后向另一租户查询/派发旧 Task；不把租户或凭据添加到公共 BusBinding、A2A 或节点输入/输出。所有原 PlanHash 维度保留，新闭包增加一个维度。场景、角色、规则和材料解析均取同一采纳快照，当前权限另外复验。

节点合同 `hub-node-contract/v1` 的输入、输出字段为 name/type/unit/period/basis，值为规范有限十进制字符串。首版只支持 identity 连线、无条件链、所有节点必需，不做隐式转换、循环或表达式。字段集合必须与原卡 Requires/Provides 和选用图一致。

## 原接口上的有限内容

原 `RemoteCall.Input` 字符串携带 `hub-bound-input/v1`：闭包/合同摘要、原 Plan/Run/Node/attempt、字段和依赖。控制端只从采纳材料或同任务已正式提交的上游构造，调用者不能自报 input/goal。随机 context/request/message 仍由原 Call 生成一次并持久复用，不放进确定性输入。

原 Job 的 `execution_profile` 和 member 的 `node_contract` 是私有只读投影；Runtime 核合同而不用提前借专业角色凭据。角色沿原 `/agent/context` 仅拿本节点输入、合同及原 call 身份，不拿完整 Plan、其它节点数据、Runtime 凭据或全局权限。

原 artifact.output 字符串携带 `hub-node-output/v1`，绑定原 input_digest、闭包/合同和完整输出字段。Worker 验传输/身份，原 CommitCall 再核正式关联并落盘；这不把 Go 变成第二个业务判断器。重复同产物复用，另一摘要冲突；未知只查询原 Task，不重派或用 StepPut 冒充成功。

## 摘要精确约定

- 原 Hub `sha(parts...)` 给每段追加 NUL，输出 `sha256:` 前缀；NodeExecutionID 和原 ArtifactDigest 沿此算法。
- 原 InputDigest 是 `canonicalHash({input:规范输入字符串,goal:采纳任务,binding:原BusBinding,model:原runModel})`；其内部同样追加尾 NUL，不能用裸输入 SHA 代替。
- 新闭包/节点合同为裸 RFC8785 规范 JSON 的 SHA256，再加 `sha256:` 文本前缀，不添加 NUL。
- 原材料 pin 的 `sha256` 为裸 64 hex。**节点输入/消费投影中的 material dependency.sha256 则是 `sha256:` 加该 pin**；上游依赖用原 `artifact_digest`，不是另一种裸 payload SHA。前端与 Java 需核对这一真实消息形状。

## 权限、恢复和兼容

新 profile 默认关闭，在独立受控部署启用。Runtime 用独立 `X-Hub-Runtime-Credential` 再核原 lease/epoch；专业角色仅持本节点 capability token，不能领取、批准或 Done。受控 profile 部署不采纳/领取 legacy 计划，防止转入未隔离模型/银行路径。普通未启用部署保持旧行为；旧 typed 任务缺原配置时停止，不能退回 legacy 解释。

所有必需节点必须有正式已提交结果才可 Done，skipped/missing/unknown 均不算完成。保存失败冻结，锁外发现后锁内复核，不在锁里做远端发现。当前租户、材料权限和版本漂移在外发/续查前拒绝。

新构建关联或 typed 历史存在时，匿名 debug 不能导出整份状态；关闭新功能也不能重新暴露历史主体、原始输入或角色 token。Graph/WorkSwarm 仅投影字段名、依赖身份、摘要和 bound/committed 状态，不投影数值与凭据。

## 资格边界

两个工程角色分别真实读取字段、计算和提交，输入值变化应传播到下游。这只证明有限工程消费链，不认证年报专业判断、独立 Gold/D2、生产 IAM 或模型收益。H04/H05、跨任务知识复用、正式共享部署与生产激活另验。
