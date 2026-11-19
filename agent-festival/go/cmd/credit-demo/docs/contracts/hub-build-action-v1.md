# Hub → 原 Builder：有限构建动作候选 v1

本应用私有候选由需求发起者批准实施，仍须上游代码接收；不是新的公共 SDK/A2A 协议或生产 IAM。

`X-Builder-Hub-Action-Context` 使用 `builder-hub-action-context/v1\n` 独立签名域。原始 method、path、body SHA256、issuer/key_id/audience、subject/scope、issued_at/expires_at 和 request_id 全部签名。Host 当前身份映射及逐动作权限独立复核；通道 bearer 不授予用户构建权，H01 query proof 不能授权 build。

## 规范命令

沿原 `builder-hub/v1` Command，operation=build、scenario=intent/1.0/gate。原 Python 规范模型含 input_ref、expected_revision、expected_generation 的显式默认 null。
`sourceDigest = SHA256(UTF8(sorted compact JSON(normalized Command minus command_id)))`，**不加 NUL**。发送原始 body 的 SHA 包含 command_id，不能拿 sourceDigest 替代。五份[跨语言向量](../../testdata/build-command-vectors.json)含完整 UTF-8/hex/摘要和公开合成签名 key；不是部署凭据。

服务器从 `[build,alias,sourceDigest]` 的规范 JSON 摘要生成内部 key，沿原 actor/subject/scope/key 派生命令身份。浏览器仅传 alias，不能指定稳定键或业务输入字节。

## 持久恢复与原所有者

复用原网络意图形状及原 Service.save；持久成功才有一次首发机会。保存失败零发送。保存后发前崩溃、丢响应、重启或 404 都只查询原 command，不重试 POST。

原 writeJSON 会格式化 RawMessage；读取时严格解析、核对同一规范内容，既不因空白差异生成新身份，也不允许重复键或材料变化。同一原 actor/alias 的 subject、scope 或内容变化冲突；旧关联保留。新的具名原业务请求由可信所有者登记，不是 UI 换键。

原 Host/TaskStore 唯一拥有领取、Task、代际和结果；Hub 关联不存执行终态、批准或预算。Host 当前 Receipt → Operation → Evidence 的类型、引用摘要、scenario、源修订逐项绑定，错绑不能提升显示。缺独立证据时，即使 Task completed 也不能称候选已核。GET 不领取、签新历史或调用模型。Host 401/403 与当前映射撤销拒绝披露，保留原动作。

输入闭包限原 Host 完整支持的固定 GATE 请求；原 Task 必须消费同一冻结字节。独立安装、Bus 接收、专业质量及生产激活各自有证据和操作者，互不替代。
