# H03 独立工程验证工具

本目录只配置受控的两个工程角色：`eng.sum@0.1.0` 消费固定材料并生成 total/reference，`eng.reconcile@0.1.0` 消费前者产物并计算 difference。固定示例为 40+60=100、100-95=5，不代表年报 Gold、D2 或生产能力。材料、ScenarioPackage 与私有 profile 的 SHA 分别固定；profile 的宿主路径不进入语义摘要。

`--help` 与 `config` 均不启动服务。`start` 才会显式启动；调用者必须在当前已批准任务的本地工作根下建立独占输出目录，协调同一台 Docker 主机的单一操作者。此处不构建、不拉取镜像、不继承现有容器、数据库、银行或模型环境，也不安装新依赖。

## 预备现有镜像

先依原部署构建路径产生本次源码对应的镜像和收据，再提供本地操作员 JSON 清单。所有镜像值必须是本机已有的完整 `sha256:<64 hex>` 镜像 ID。`source_lock_sha256` 是 `deploy/sources.lock.json` 原始字节 SHA256；清单中恰好列所选服务，禁止可变 tag。示意结构（占位内容须替换）：

```json
{
  "schema_version": "h03-local-images/v1",
  "source_lock_sha256": "<64 hex>",
  "images": {
    "credit": "sha256:<64 hex>",
    "orchestrator": "sha256:<64 hex>",
    "engineering-sum": "sha256:<64 hex>",
    "engineering-reconcile": "sha256:<64 hex>",
    "bus-db": "sha256:<64 hex>",
    "rdc": "sha256:<64 hex>",
    "gateway": "sha256:<64 hex>"
  }
}
```

清单是本地构建收据的入口，不能仅因填写了源码锁摘要就宣称镜像确由该源码构建。真实源码/镜像闭包由后续原生验证核实。`--ui original` 额外要求 `swarm`、`front` 镜像及 `ui` 收据：`host_source_revision` 必须等于锁文件的 WorkSwarm commit，`model_free` 必须为 true，`plugin_sha256` 是四个原插件文件路径→文件 SHA256 映射按排序紧凑 UTF-8 JSON 编码后的 SHA256。缺原宿主、插件生成物或无模型运行前提即 BLOCKED；不以模拟页面代替宿主验证。

## 配置、显式启动与清理

在 Linux/WSL 的现有本机 Docker 默认 context 下运行，替换占位路径；先阅读帮助：

```bash
python3 launch_h03.py --help
python3 launch_h03.py config --project h03-owned-example --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/config-only --images <private-images.json>
python3 launch_h03.py start --project h03-owned-run --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/run --images <private-images.json>
python3 launch_h03.py stop --project h03-owned-run --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/run
```

每次 config/start 必须使用新的输出目录，默认只选七个 core 服务；原宿主 UI 须显式 `--ui original`。工具保存未展开的 compose、材料/profile/场景只读副本、源码锁关联和镜像 ID；检查原始模板、Compose 展开配置以及启动后实际容器的环境、挂载、网络、权限与所有者。生成的运行身份和数据库等凭据只写本地 `private-local.json`，不得提交、打印或附到公开报告。`configuration-audit.json` 只含环境键名，不含秘密值；不要公开完整 Compose config 或 Docker inspect 输出。

所有服务使用新建内部网络、新数据卷、独立名称和所有者标记。Go 与 Runtime 共享独立运行凭据，角色容器没有该凭据。仅 credit 的 18191 端口及可选原 UI 的 18192 端口绑定回环地址。受控 RDC 配置的无签名卡例外仅限该独立网段；它不是生产身份方案。`SWARM_*` 环境变量即使为空也拒绝，Docker/Compose 外部配置覆盖拒绝；镜像中未知环境变量、非原入口、外部/宿主网络均拒绝。

`stop` 只处理本目录记录对应的项目、镜像及所有者，保留数据卷和 Task/UNKNOWN 历史，不重派或删除未决执行。启动后审计失败会返回 BLOCKED；操作员须依据原有 ownership 记录检查并调用同一 stop 命令，不得把失败当成成功再启动另一组服务。内部网络配置本身不是出站实测证据；真实双角色消费、Begin/Commit、错误路径、出站观察、原 UI 消费和复现属于 T313，工具不会自动签收。

## 无依赖 Java plain-main probe

两个 probe 类须由对应 Java Task 提供。本 runner 只接受以下精确组合：

| module | probe |
| --- | --- |
| `examples/credit-orchestrator` | `com.openjiuwen.demo.credit.H03BindingProbe` |
| `examples/professional-agent-runtime` | `com.openjiuwen.demo.professional.H03EngineeringProbe` |

```bash
python3 run_java_probe.py --help
python3 run_java_probe.py --module examples/credit-orchestrator --probe com.openjiuwen.demo.credit.H03BindingProbe --maven-cache <existing-maven-cache> --build-image sha256:<existing-build-image-id> --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/orchestrator-probe
```

runner 复制现有模块和现有 Maven cache 到独占输出目录，记录文件 SHA，使用 `mvn -o`，运行精确 plain-main 类，要求明确 `H03_PROBE_PASS` 标记。镜像模式必须使用 Linux 本机已有的不可变构建镜像，以 `--network none`、只读根文件系统、无 capabilities 和私有临时目录运行；不会调用应用 Main、启动 Bus/业务服务或下载依赖。宿主模式要求已安装 Java/Maven，且仅是 plain-main probe，不提供网络隔离证明。缺缓存、工具、probe 类或断言回执均 BLOCKED。probe PASS 仅表示局部断言通过，真实 Runtime 验收仍为 NOT_RUN。

## 本地自测

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s . -p test_launch_h03.py -v
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s . -p test_run_java_probe.py -v
```

上述测试全部使用 mock Docker/Maven，不启动服务。实际 Compose parser 检查必须另行记录并明确 synthetic 镜像 fixture 与真实镜像审计的区别。禁止将历史 native 报告或 parser/probe 的局部 PASS 写成当前 H03 原生成功。

## T313 原链验证驱动

`verify_h03.py` 默认仅准备本地证据清单；与计划所列命令一致，不会隐式启动 Docker 或连接服务：

```bash
python3 verify_h03.py --project h03-engineering --scenario engineering-two-role --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/prepared
```

只有在源码冻结、已有镜像来源与独占部署经统一协调后，显式 `--run` 才连接 T312 已启动的原链。CLI 使用本地固定回环端口，拒绝代理/远端 Docker 覆盖，重核每个容器的所有者、镜像、运行状态、实际环境与内部网络。它不启动服务，不调用 stepPut，不生成执行结果，不创建另一调度器。本轮该固定端口未兑现，CLI 原生结果为 BLOCKED；下述已验证操作员 API 路径须单独记录。

```bash
python3 verify_h03.py --run --project h03-owned-run --deployment-root <existing-T312-owned-root> --image-receipts <fixed-native-image-receipts.json> --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/baseline
```

镜像来源清单须使用真实原件，格式为 `{"schema_version":"h03-native-image-receipts/v1","images":{"credit":{"image_id":"sha256:...","receipt":"relative-build-receipt.json","receipt_sha256":"raw-file-sha256","source_manifest":"relative-source-manifest.json","source_manifest_sha256":"raw-file-sha256"}}}`。images 恰含所选七或九个服务，image_id 必须等于 T312 ownership 的固定镜像 ID。原件路径限定在当前 HUB-NEXT/evidence 内；清单做原件摘要及镜像关联一致性检查，不代替独立来源审查。不得伪造 build/test/来源 PASS 来填清单。

驱动通过原 MCP initialize/select_capabilities/propose_plan，检查完整两角色选择与预检，再走原 `/ui/login` 登录和原确认页表单批准。原 Java Worker 自动领取并执行原 Runtime/Bus 调用。驱动只通过自有 credit 容器的固定 `/bin/cat /data/state.json` 读取私有状态，在内存解析；该运行工具缺失即 BLOCKED。原 `/debug/state` 不作取证出口：受控 profile 的实际外层 handler 返回 404 静态正文，不泄露私有状态。build/history 的内层 403/no-store 是另一条门禁，由其独立测试覆盖。

报告只保存允许的正式输入/输出字节、原 InputDigest（包含模型与 goal 的原式、尾 NUL）、合同/闭包摘要、原 Task/context、产物摘要及原调用记录计数。租约 token、节点 capability token、密码、Runtime 身份不落入报告。两个独立角色必须分别形成原 Task 和 context，正式下游依赖必须引用上游已提交产物。原 `/api/graph` 的 committed 消费状态与同 Run 确认页引用也必须一致。

`--case baseline` 要求固定材料 40/60/95，结果 100/5；`--case changed` 要求另一个独占部署的已批准固定材料 41/60/95，结果 101/6。不得热改原已采纳快照、借换 ID 绕过 UNKNOWN 或把旧结果当新结果。两次实际运行后可离线比较：

```bash
python3 verify_h03.py --project h03-comparison --compare-baseline <baseline-verification.json> --compare-changed <changed-verification.json> --output-dir <H>/.specify/feature/HUB-NEXT/evidence/native/comparison
```

正常链之后，`--unknown` 可另做显式故障负例：暂停本次自有 gateway，等待原 SendMessage 超时、原 Run waiting 且原租约已由 Done 清空，finally 恢复 gateway。随后严格核对 credit 容器所有者与固定镜像，实际 restart 同一容器（保留原 DataDir/数据卷），检查容器 StartedAt 已改变，等原控制服务恢复，重新建立原 MCP session 和登录 session，再恢复同一 Run。重启前后及 Resume 后的同一 Call 身份和发送记录数必须保持，且不创建新 Plan/Task 来绕过 UNKNOWN。该项目保留未决历史，不代做外部结果签收；暂停/重启窗口必须由同一原生操作者协调。驱动仅在此明确选项下重启现有自有 credit 容器，不执行 up/start 或新建部署。

工具自身的错误批准、period/basis 排除、exec/model/member/debug 拒绝和完成后 Commit fencing 负例与正常业务分开记录。原 Done 会清空租约，所以完成后重复 Commit 应为 409；这不证明有效租约内幂等或字段校验。有效租约内重复 Commit、慢 RDC、持久化失败等独立原生故障窗口仍需后续实际证据。

此驱动当前只判定原工程链与 Go 页面投影；真实 WorkSwarm 宿主/任务后台/选中任务的浏览器观察、独立出站连接与拒绝计数、实际 Handler 模型和银行请求计数均须另外取证。缺前置保持 BLOCKED，正常原链可继续分项；框架 ModelEvents=0 不等于实际出站计数=0，Go 确认页不代替原 WorkSwarm 页面。整体报告不会仅凭本工具把 T313、Gold、D2 或生产标成 PASS。

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s . -p test_verify_h03.py -v
```

这些自测均用合成状态与 mock 传输验证驱动边界，不是原生运行证据。

## 本轮实际交接（2026-11-17）

本轮已实际走通原 MCP 选择/提案→原登录与确认页批准→原 Java Worker/Runtime/Bus→两个独立角色→正式 Step 提交。独占 baseline09 的 P1/R1 得到 total=100、difference=5；独占 changed10 的 P1/R1 得到 101、6。两组各自有两个不同 Task/context，每个角色各一次 SendMessage；正式输入、上游产物引用、原 InputDigest/ArtifactDigest 和原 Go Graph 的 committed 投影均通过校对，离线比较确认材料变化传播到采纳/输入/产物摘要。

baseline09 另一次明确故障负例 R2 在原 SendMessage 超时后进入 UNKNOWN。原租约清空后，实际重启同一 owned credit 容器、保留数据卷、重建 MCP/登录会话，再 Resume 同一 R2；原 Call 身份与摘要保持、仍 UNKNOWN、发送数保持 1。基线已完成 R1 在 QA reader 修复后仅只读复验，未重新派发补成绩。两个项目最后均经原 owned stop 停止，保留数据卷、原 Run/Task/UNKNOWN 和早期失败证据。

| 本轮消费来源 | 固定版本或摘要 | 解释 |
| --- | --- | --- |
| Go 运行源码 | `e09f71e3d9973fb44fd15fa54ca7bc95ff5e992c` | Linux binary SHA256 `032bfa1d14a609d4b6c576db0e0f48c71727a906a0ca8e445fee988dd2188b72` |
| credit 运行镜像 | `sha256:3b2d2056ee0d3d0910c5a84d5df4a8acbe5f0b1ec8b6014b0224521d3fd5e096` | 固定原 runtime base 上刷新已验证 binary，未宣称完整重跑原 Dockerfile |
| QA reader 源码 | `9f6365a2311d8d167e5d62517ec028c17365bfe0` | 持久 StepRec 使用原 `remote_call` tag；HTTP Begin 的 `call` 不变。修复后 16 个 stdlib 测试通过，不替换已运行 Go/Java 字节 |
| Java 装配来源 | `fe14af28d51296a48ecea4bdfe791054f1b4c670` | 九个模块源码文件复核未变；orchestrator JAR SHA256 `0b947bced7f698fc1266436a5ba876f70e7d3ea7e4fd846c5ee37cef3dc1c9ed`，professional JAR SHA256 `177c51d33499d9e9fd40a42a4dcf9202645b748ea5aee0e1532a0c36e3787a89`。沿用其实际离线 Maven/probe 回执，镜像刷新未重跑全 Maven |

Bus 运行依赖另绑定本轮 T405 的原 source closure `27bb09f7dfae5057b2573a4c021ce47ce37f25309ef01eacda048b21c61090bd` 与实际 build/test 收据，不能把可变 tag 或历史测试当作本轮来源。交接记录分别封存 SourceImage/binary/JAR 的前后 SHA、原 base image ID、输出 image ID 和材料派生摘要；后续仅文档提交不会改变这些已验证运行来源。

### 可信本机操作员 API 前置与步骤

本轮 Docker internal 网络的 credit 容器虽声明回环 PortBindings，实际 `NetworkSettings.Ports["8091/tcp"]` 为 null，固定 `127.0.0.1:18191` 不可用。成功的工程结果来自已批准的本机操作员 API：仅将原 `Client` 实例的 `base` 配置为严格审计后的 owned bridge IP，仍调用以下原函数链。它不是 CLI 原生成功，也不是可接受任意 URL 的新增入口。

1. 冻结源码、binary/JAR 与本机不可变镜像及来源收据；由单一操作者使用原 `launch.prepare(project, ui, output, images_path, *, here=HERE, festival=FESTIVAL)` 和 `launch.start(prepared)` 建立新 owned 部署。先完成 raw/resolved/actual 环境、挂载、网络、用户与镜像审计；不得改已 prepare 的 ownership/config 或运行中镜像。
2. 调用 `deployed(project, root, receipts)`，重核原 source receipts、配置/材料摘要及全部容器。只对返回的 `deployment["containers"]["credit"]` 做本机 inspect；重核 ID、固定 image、owner/service/project 标签及唯一 `<project>_h03` 网络，取该网络的 IP，并验证它在已审计 expected IPAM 子网内。缺一项就 BLOCKED；不要打印完整 inspect、私有 env 或 state。
3. 构造原 `client = Client()`，仅此实例设置 `client.base = "http://" + owned_ip + ":8091"`。记录安全的容器/镜像身份及地址来源；不修改 `Client` 类、不 monkeypatch 审计或模板 pin、不加 proxy/echo/新 runner、不停用门禁、不修改网络路由。
4. 调用 `client.initialize()`、`client.login(password)`；密码仅从本次 deployment 的私有配置在内存读取。调用 `planning_negatives(client)`，再 `plan, selection = propose(client)`；原函数通过 MCP 执行 select/propose。调用 `client.approve(plan)`，由原登录 cookie 和确认表单取得 Run ID；Java Worker 自动领取，操作员不写 Step 或伪造结果。
5. 用 `wait_state(deployment, run_id, {"done"}, timeout)`、`formal_receipt(state, run_id, expected_a)`、`graph_receipt(client.graph(run_id), formal)` 逐项校对；再核原 `/ui/run/<run_id>` 引用。`native_negatives(deployment, client, state, run_id)` 的拒绝探测单独计时记录。若只是 QA reader 修正，读取同一个已存在 Run，禁止再次 propose/approve 补成绩。
6. changed 使用另一独占目录，在 prepare 前由可信操作员派生固定材料：仅 component_a 的原字节 40→41，更新其实际材料 SHA、场景引用及 profile 的材料/场景摘要；其它字节保留，记录 before/after SHA。通过 prepare 的原 `here` 参数消费该固定目录，不声称 CLI 自动生成 changed 材料。使用同一函数链实际执行后，以 `compare(baseline, changed)` 比较安全回执。
7. UNKNOWN 是另一个显式故障负例，调用原 `unknown_check(deployment, client, password, timeout)`；只对本次 owned gateway/credit 操作。它会产生一次负例提案/Run，重启后必须恢复同一 Run，不能重建任务绕 UNKNOWN。完成后调用原 `launch.stop(project, output)` 保留卷和未决历史。不得为了重现本轮成绩重新运行已完成的项目。

以上 API 步骤需要本机能直接访问已审计的 owned bridge IP；不具备该前置就 BLOCKED，不改共享网络或退回未审计地址。安全交接证据包含两组正式回执、比较、UNKNOWN restart、派生材料和镜像来源冻结；秘密只留原私有部署目录。

上述早期工程记录保留原 WorkSwarm、独立外发观察及有效租约故障窗口的缺证。后续增量补验另记如下；完成后 Commit409 与框架 model_events=0 均不代替这些新证据。合成材料不构成年报 Gold、业务 D2、生产 IAM 或发布准入。详见[整体消费交接](../../../../../docs/h03-bound-node-consumption.md)。


## 原 WorkSwarm CLI 宿主增量配方

本轮已实际验证原宿主、原插件和原任务后台；使用原 CLI 的 WSL/Linux 本机路径。九服务 Docker `--ui original`、Windows 原生以及生产 IAM 不由此获得资格。原 H03 固定 localhost CLI 的网络限制也没有被改写成通过；插件只连接经 `deployed` 审计的 owned Hub。

### 固定来源与安装

- 原宿主源码：`openJiuwen-ai/jiuwenswarm@1142e5fb3984a2b1bc6352669a464976e204b2dc`。保留原 runtime、前端、manifest 与原打包钩子 `scripts/build_config.py`，不能拿旧工作副本或缺文件的稀疏下载代替。
- 本轮 CPython3.12.3、安装工具 uv0.11.7；265个实装包（包括原宿主）与原 uv.lock 匹配、pip check 通过，6个 Git 依赖提交匹配。运行依赖中的 uv0.12.5 是原锁内容，不冒称它也是安装工具版本。
- 原前端按原 package-lock 执行 `npm ci --ignore-scripts`、`npm run build`，包括原 prebuild/tsc/Vite；不得换空 dist。打包阶段不要携带自行建立的 node_modules 依赖链接，本轮仅移除自己的链接，未删除其目标依赖。
- venv/cache 使用 Linux 原生独占目录，源码与本地证据另存；本轮 NTFS 大量文件落盘较慢的阶段和传输/打包失败均保留，未计为另一份成功。

安装时固定 `UV_PROJECT_ENVIRONMENT`、`UV_CACHE_DIR` 与 `UV_PYTHON_DOWNLOADS=never`，从固定原源码目录调用原 uv0.11.7：

```bash
"$UV_BIN" sync --frozen --no-dev --no-editable --link-mode hardlink
"$UV_BIN" pip check --python "$UV_PROJECT_ENVIRONMENT/bin/python"
```

`UV_BIN` 和各目录须为操作者已核验的本地绝对路径；这两个命令不重新解析或升级原锁。不要复制用户既有 workspace、API key、SWARM_* 或模型/银行配置。

### 独占原配置与插件

1. 给原 CLI 一个新独占 HOME，原 `jiuwenswarm-init` 不加 `--force`。原默认根是该 HOME 下的 `.jiuwenswarm`；以原 `get_config_dir`、`get_domain_data_dir` 和 Gateway `get_root_dir` 实际返回值核对，再写配置。仅设置 DATA_DIR 而与默认 HOME 分离可能回退包内配置，本轮首个尝试因此被本地断言拒绝，不能继续向包内或共享目录写入。
2. 模型默认列表为空，不提供真实 key；原配置 updater.enabled=false。**另须在原进程启动前设置 `JIUWENSWARM_CONFIG_URL=off`**，使用已锁依赖原支持的关闭方式，避免默认远程登录/赠送模型活动配置预热。该设置只作用于此本机参考环境，不改变原 Hub 登录或任务权限，也不宣称生产 IAM。
3. 使用原已支持的无前缀 `GATEWAY_PORT`、`AGENT_SERVER_PORT`、`WEB_PORT`、`FRONTEND_PORT` 设置独占端口。本轮分别为19401、19402、19403、19404，frontend host绑定127.0.0.1。不要以不被原入口识别的变量替代这些变量，也不要占用他人默认端口。
4. 将 `workbench-plugin` 的 `extension.py`、`extension.yaml`、`frontend/dist/index.html`、`frontend/dist/app.js` 按实际 SHA 安装到原 domain 下的 application_plugins/credit-workbench；由原 ExtensionManager 加载。实际 domain 路径必须由原 helper 核对，不猜目录。
5. 原 `CREDIT_INTERNAL_URL` / `CREDIT_MCP_URL` 只能指向本次严格核验来源、所有者、镜像、唯一内部网络及地址的原 Hub，不能由浏览器注册任意后端。使用原 `jiuwenswarm-start all`，不运行要求 SWARM_* 模型变量的演示 aha-configure 包装。

这是一份经原 CLI 实跑的本机配方，不是新增部署/调度系统。没有变更宿主、SDK、公共协议、数据库或原锁源码。

### 实际验收和观察边界

独立操作者实际打开原 WorkSwarm→「编排台」→原插件 iframe，使用真实原 WebSocket plans/graph：R1 正式两节点结果、选用分母/理由、字段/材料/上游摘要匹配原 Go Graph；R2 waiting/UNKNOWN 及缺失远端 Task 回执保持原样。刷新后仍选同一原任务；前后原身份、SendMessage/GetTask记录不变。实际 served index/app.js 字节 SHA 与原插件/本次安装收据相等。没有通过模拟 RPC、假任务、聊天、批准或 Resume 生成成绩。“跳过首次模型指引”只写宿主设置，不能把任务/业务只读扩大成整个宿主绝无设置写入。

观察使用启动前建立的独占 cgroup及只过滤本组的私有 syscall observer，先附加后启动原 init/CLI 子进程，持续覆盖页面读取，子进程收尾后关闭。精确端点、显式数据报、DNS、TLS、向量不完整与map损失分别记录；不改变共享 tracing/防火墙。最终 off 配置窗口相应非许可计数及损失为0，只对这个已观测生命周期/协议范围签收，不是任意语义全称零或未来保证。没有记录旧monotonic起点时不补造时长；墙钟标签与生命周期顺序分开。

早期默认配置和仅关闭 updater 的窗口确有外连/opaque TLS。定向 DNS 为 aigw.openjiuwen.com，原依赖 remote_config.py 定义 GET https://aigw.openjiuwen.com/v1/config；旧 TLS 内容并未解密，不追认为零调用，也不归因为已经被否定的 updater 假设。新 off 配置是独立具名运行；旧失败与原 NO-GO 不覆盖、不改写。

### 收尾

原 CLI stop 在本轮报告失败，不能写成该命令通过。随后只向本次已核对的独占原启动器 PID 发送 SIGTERM，由原 supervision 收尾；确认自有 cgroup为空、私有观察FD关闭、四端口不可连接，再移除空cgroup，保留 workspace/凭据所在本地目录供有权复核。不要批量终止同名进程、删除共享目录或删除原 Task/UNKNOWN 以恢复首发资格。原七服务另走原 `launch_h03.py stop`，保留数据卷与未知历史。

本轮有限工程补验可供代码评审；真实业务 Gold/D2、专业年报角色、生产身份、共享部署、H04/H05 及发布仍分别验收。
