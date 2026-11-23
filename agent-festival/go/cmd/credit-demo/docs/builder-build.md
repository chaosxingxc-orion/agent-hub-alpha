# 从原 Hub 构建 Intent GATE 候选

这是默认关闭的应用接线候选。沿原 `/ui/login` 登录并打开 `/ui/builder`，选择可信操作者登记的参考任务，再提交构建。Hub 保存原命令关联后才可能发送一次；原 Intent Host 拥有命令领取、Task、代际及结果。构建、独立接收、Bus 部署注册和专业业务验收是不同事实。

## 有限配置

保留 H01 的 `BUILDER_HUB_ENABLED=1` 及原查询配置。新增配置仅在 `BUILDER_BUILD_ENABLED=1` 时启用：

| 变量 | 要求 |
| --- | --- |
| `BUILDER_BUILD_URL` | 与 H01 相同的 Host；HTTPS 或显式数字 loopback HTTP；无路径、用户信息、query 或 fragment |
| `BUILDER_BUILD_CHANNEL_TOKEN` | 原 Host 明确配置的同一通道凭据，16～4096 字符，无 CR/LF；不是用户构建权 |
| `BUILDER_BUILD_KEY_HEX` | 原独立动作签名密钥的十六进制，32～128 字节；不复用查询密钥、登录密码或会话密钥 |
| `BUILDER_BUILD_ISSUER` / `AUDIENCE` / `SCOPE` | 与 H01 的同一 Host 身份边界一致 |
| `BUILDER_BUILD_KEY_ID` | 原 Host 已信任的动作密钥标识 |
| `BUILDER_BUILD_USERS_JSON` | 原登录用户到受控 subject 的可信映射；Host 仍独立核验当前 build grant |
| `BUILDER_BUILD_TEMPLATES_JSON` | 不可变原业务请求 alias 到有限 Command 模板的服务器登记；最多 64 个，不由浏览器注册 |

`CREDIT_DEMO_PASSWORD` 必须显式设置且不是默认 `demo`。不要将密钥或私有配置提交仓库。受控用户映射不是生产 IAM，原查询前置也不能提供构建授权。

原 Hub 要求状态文件及所在目录同步成功后才能接受动作。本轮 Windows 启动在目录同步时出现 `Access is denied`，尚未产生构建命令；Windows 原生运行未获资格。受控验收改用 Linux/WSL 中的原 Hub 程序及独立可同步的数据目录，保留同一持久化门禁。跨系统部署须先确认双向本机可达和数据目录同步能力，不能忽略同步错误或删除历史来恢复发送资格。

模板仅含 `schema_version=builder-hub/v1`、固定 `scenario={id:intent,version:1.0,profile:gate}`、`operation=build`、`input_ref`、40 位 `expected_revision` 和 `expected_generation=0`，**没有 command_id**。输入引用须来自原 Host 的完整固定 GATE 闭包：query、意图、配置、目录及相关依赖全部核验；样本/规则在原独立接收阶段对应同一冻结来源。缺项或漂移不能签发已核候选，一个文件 SHA 不能代替完整闭包。

本批没有任意自然语言/文件上传、HTTP 材料注册、取消、评测、模型调用或生产激活入口。原四例通过是有限工程资格，不是任意材料/专业质量证明。

## 原页面与恢复

- `POST /api/builder-actions/build` 只接受 `{"alias":"已登记请求"}`，拒绝浏览器 request_id、命令身份、路径或凭据覆盖。
- `GET /api/builder-actions/builds` 读取当前用户的有限别名及已保存关联。存储失败返回不可用，不能伪装成空历史。
- `GET /api/builder-actions/commands/{command_id}` 核对同一原命令，再按当前原 Receipt 的引用读取 Operation/Evidence；不接受浏览器指定的上游地址或证据来源。GET 无构建/批准/预算副作用。

刷新、换浏览器或重启后重新登录，从后端找回原关联，不依赖浏览器存储命令键。先保存后崩溃、丢回执、超时、Host 404 都不能恢复首发资格。原 `builder_actions` 网络关联只表示可能曾外发，不拥有执行终态。

同一原 actor/alias 的 subject、scope 或内容变化会冲突，不能悄悄生成新命令。真正的新业务请求由可信操作者登记新的具名请求版本。撤权拒绝披露，不删除已发生动作。

独立环境必须安装并实际调用同一新 wheel。原 RDC/Bus 接收沿 [本机部署指南](../deploy/builder-h02b/README.md) 使用独立操作者权限，不能继承用户 build grant；旧不兼容 Bus 与既有兼容候选分别核验。

浏览器自验入口为 `node e2e/builder-build.mjs --hub-url http://127.0.0.1:端口/ --alias 已登记请求 --output 本工作根下的新证据目录`。密码从 `BUILDER_E2E_PASSWORD` 读取，现有浏览器可用 `BROWSER_EXECUTABLE` 指定；不安装新浏览器依赖，不拦截或模拟后端。默认构建模式、`--case disabled` 和 `--case denied` 分别需要对应受控服务状态。脚本记录浏览器 POST 次数；原 Host 发送次数须独立核验，不能把两者混算。

关闭新开关可回滚，保留状态、原 Task 与 UNKNOWN。正式 Hub 接收、真实业务、共享部署及生产资格仍分别确认。
