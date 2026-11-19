# H02-B 本机原 Bus 候选装配与 Hub 交接

仅用于获准的本机合成实验。本次 T302 从参考提交
`4b0b78f89f9524f45f77ce2a7e96487a33fc85c6` 有限复用本目录五份文件；启动器、模板和测试保持原字节，
只适配本说明。当前尚未运行本批原 Bus 构建、Maven、部署或注册；单元测试和帮助检查不能替代这些验收。
原组件基线仍为 W05 sources.lock 的
`agent-solution@bf702bc95064d3491cf8b374f54d271363d84230`，不改原锁；受限兼容候选须另有明确授权及
匹配的实际源码/构建/原 Maven 回执，本次移植不授予修改 Bus 源码的权限。只装自有 Postgres/RDC/Gateway，复用原 Dockerfile.bus，不启动完整授信/
WorkSwarm/真实模型/银行。不启用 PUSH，不改 SDK/原卡，不另建注册或调度系统。

## 旧基线与候选资格分开

未经修改的 bf702 原源码仍不兼容：卡校验要求根 version 并用卡/协议版本覆盖部署业务版本；
原 [Gateway HTTP客户端](https://gitcode.com/openJiuwen/agent-solution/blob/bf702bc95064d3491cf8b374f54d271363d84230/common/agent-bus/agent-gateway/src/main/java/com/openjiuwen/gateway/direct/HttpAgentRuntimeClient.java#L44)
同步/流式/订阅不设置 `A2A-Version`。只换登记模式无法解决两项问题；旧默认路径保留原
compatibility gate，不会因为候选出现就改成 PASS。

候选静态条目显式 `version-source: deployment-binding`，原 service-id 仍取 native_agent_id，
逻辑别名仅在 Hub。私有材料必须为严格24字段 publication-reference/v2，显式 capability_version（1..16）、
deployment_version（1..32）、instance_id（1..64），字母数字起始、后续仅字母数字/点/下划线/
连字符。完整 artifact_sha256 仍为64位独立摘要，不能截短充当业务版本。缺项/旧v1/额外字段/
超长或不安全token拒绝。来源/租户/实例精确匹配；原内容/网络/签名
门禁不放宽。新模式可使用原生1.0卡并规范同源完整URL；默认 `agent-card` 行为不变。
本机合成配方沿此前明确的 unsigned Demo 例外，模板显式 `verify-signatures: false`，
并限制 loopback；这不证明生产签名准入。原组件默认签名策略不因新版本模式改变，
启用签名时仍执行原验签与拒绝逻辑。
Gateway 通过 `GATEWAY_RUNTIME_A2A_PROTOCOL_VERSION=1.0` 显式选择头，默认空保留旧行为。
本参考是单协议部署，不提供逐目标协议协商，也不据流式头部测试声明流式原生通过。

当前完整双包 Bus→Hub 原生资格仍待实际报告及独立审查确认。构建成功、模块测试或准备
配置成功各自有意义，但不能替代实际原登记、业务 Task/查询、旧绑定及自有清理。
历史接收失败及未执行项单独保留，不计入本次工具移植或后续新原生验收；不借旧包补本轮成绩。

## 原服务启动器与资格顺序

在本目录运行；默认只显示帮助，不默认构建、测试或启动。下例变量均由操作者指向私有
本机目录，构建/测试/实验状态目录须新建，源码须为获准候选；不从用户请求中取路径。

```sh
python launch.py --help
python launch.py build --candidate --source-dir "$BUS_SOURCE" --state-dir "$BUILD_EVIDENCE"
python launch.py test --source-dir "$BUS_SOURCE" --state-dir "$TEST_EVIDENCE" \
  --build-receipt "$BUILD_EVIDENCE/image-build.json"
python launch.py prepare --candidate --source-dir "$BUS_SOURCE" --state-dir "$NEW_LAB" \
  --descriptors "$PRIVATE_DESCRIPTORS_JSON" \
  --build-receipt "$BUILD_EVIDENCE/image-build.json" \
  --test-report "$TEST_EVIDENCE/test-report.json"
python launch.py up --state-dir "$NEW_LAB"
python launch.py status --state-dir "$NEW_LAB"
python launch.py down --state-dir "$NEW_LAB"
```

- `build --candidate` 记录 bf702基线加允许的 tracked/untracked 源文件闭包摘要、原 Dockerfile
  摘要和实际镜像 ID，前后核验；源范围外变化、链接文件、来源错绑均拒绝。派生 CodeGraph
  不作为产品源码或发布证据。候选镜像标签包含闭包摘要，不冒用纯 bf702 的旧镜像标签。
- `test` 消费确切构建回执，在自有临时容器中以非root运行原 Maven，保留 Surefire XML。
  资格检查重读原 XML、源闭包、构建回执和镜像 ID；关键新增 PG/卡/协议测试必须实际运行，
  七组关键测试必须非零且无跳过；有失败、缺报告或关键跳过不得准入候选。
- `prepare --candidate` 必须同时消费匹配的构建/测试回执，并核实际卡接口；设置静态实例新
  模式和 Gateway 的1.0配置。成功输出仍是 `PREPARED_NOT_STARTED`、registration=NOT_RUN。
- 不带 `--candidate` 的旧模式只接受固定原源码，保留原不兼容判定，不能用fixture或修改卡
  字段绕过。`status` 的可达不等于已登记；实际 card/instances/Task 仍须独立核对。

若固定原基础镜像拉取受限，可显式 `build --candidate --mirror-original-base` 使用原相同
基础镜像的镜像传输路径；原始标签/镜像引用与解析摘要进入回执，不能偷偷换基础版本。
若现有受控环境必须用既有代理，可显式 `--use-existing-proxy` 仅转交已有代理环境变量名，
不在命令/日志/回执中公开值；没有已配置代理就拒绝。两项都是明确选择，不是新源码或依赖。
未验证来源/传输条件时保留失败，不宣称可复现构建已通过。

若原固定Postgres镜像名称无法通过本机网络获取，可先由操作者获取公共镜像中的**相同原摘要**，
然后显式 `prepare --mirror-pinned-postgres`。该选项仅核验已存在镜像的RepoDigest、linux/amd64
平台及imageID，不拉取或启动；ownership.json保存公开传输来源，up再次核验同一ID。默认仍为
原 postgres:15-alpine@sha256:09e4f20b14ddb3dfe3a0c825b206032aaf8f28300ba2070c0b60fc1c10c6abc7。
不改全局镜像源、不换版本。Compose若发现环境变量覆盖生成的private.env键立即拒绝，防止
已核验镜像被外部shell值替换。代理URL必须无凭据且为合法HTTP(S)，报错不泄露配置值。

真实 Builder 描述符数组不能由 fixture 代替。prepare 可选
`--retained-descriptors "$PRIVATE_RETAINED_JSON"` 消费最多32条显式历史v2描述符，默认不合成。
当前1..2条与历史集合按固定身份元组检查冲突：同 tenant/native/capability 不能指向不同完整SHA；
同 tenant/instance 不能指向不同原生ID/包/版本/地址/卡及契约摘要。冲突在源码或Docker操作前拒绝。
只有当前条目生成部署行，历史集合不会登记或授予查询权限。prepare 可指定 --db-port/--rdc-port/--gateway-port，
默认15432/18092/18080；占用即拒绝。源码/构建上下文较大时可在获准Linux本地文件系统使用
精确副本，但必须保留相同基线和工作字节，重新核对闭包与工具来源，不能只复制一个PASS文件。

私有 private.env 保存 RDC_REGISTRY_API_TOKEN、H02B_GATEWAY_TOKEN 和 H02B_TENANT 等，
不打印/提交值。仅数字 loopback；RDC/Gateway host network 访问自有候选，DB只映射自有
loopback端口。Docker `default` context 须是本机 Unix socket；Buildx `default` 须为单本机
节点的 docker driver，启动器固定 --context default --builder default，拒绝远端或环境覆盖，
不切换全局 builder。源码、镜像和测试材料是不同来源证据，不能只核容器可启动。

仅可按 nonce 与归属凭据停止本次自有资源。数据卷、原Task/命令、旧UNKNOWN及历史原件保留，
不执行 down -v，不停止共享服务。清理无证明就保留未知，退出进程不等于清理完成。

## Hub 私有配置

可信本机 JSON 文件形状如下，publications 须填本轮真实描述符，空数组会拒绝启动：

```json
{
  "registry_url": "http://127.0.0.1:18092",
  "tenant": "builder-h02b",
  "allowed_users": ["alice"],
  "publications": []
}
```

只接受数字 loopback HTTP+显式端口，无路径/用户信息/查询/片段。最多8个唯一逻辑别名。
可选 retained_publications 数组最多32条，按上述同一身份元组拒绝冲突，仅用于历史一致性检查，
不为历史logical_id增加路由读权限。两数组中的每条材料都必须严格v2；v1历史fixture保留作拒绝测试。
配置是私有常规文件，不使用符号链接。allowed_users 必须显式列原已存在用户。

原独立程序启动仍依赖原Schema和有效Judgment Spec，不是只复制一个二进制即可运行。
本机参考必须保留完整Hub检出目录，并以模块目录为工作目录；原Schema查找器按工作目录
解析 `../../../schema`。`CREDIT_SPEC` 显式指向原受校验的有效样例，不能使用私有目录内
不存在的默认basename；不得放宽校验器来绕过启动失败。

以下路径均用WSL/Linux绝对路径：HUB_ROOT是获准Hub检出根，HUB_BINARY是已核对来源的程序，
CREDIT_DATA是本轮独立、可写的私有状态目录；不得沿用共享 `/data` 或其他实验的状态。

```sh
: "${HUB_ROOT:?set an absolute trusted Hub checkout}"
: "${HUB_BINARY:?set the verified executable absolute path}"
: "${CREDIT_DATA:?set an absolute private writable state directory}"
case "$HUB_ROOT" in /*) ;; *) exit 1 ;; esac
case "$HUB_BINARY" in /*) ;; *) exit 1 ;; esac
case "$CREDIT_DATA" in /*) ;; *) exit 1 ;; esac
cd "$HUB_ROOT/agent-festival/go/cmd/credit-demo"
export CREDIT_SPEC="$HUB_ROOT/agent-festival/schema/examples/judgment/valid/judgment-spec-credit.json"
test -r "$HUB_ROOT/agent-festival/schema/judgment.schema.json"
test -r "$CREDIT_SPEC"
test -x "$HUB_BINARY"
mkdir -p "$CREDIT_DATA"
export CREDIT_DATA
export BUILDER_PUBLICATION_ENABLED=1
export BUILDER_PUBLICATION_CONFIG=/private/h02b-publication.json
export BUILDER_PUBLICATION_RDC_TOKEN="$RDC_REGISTRY_API_TOKEN"
export CREDIT_DEMO_PASSWORD="$H02B_LOGIN_PASSWORD"
export CREDIT_ADDR=127.0.0.1:18091
export CREDIT_PUBLIC_URL=/
"$HUB_BINARY"
```

`CREDIT_PUBLIC_URL=/` 明确选择本机根路径。当前接收基线 #30/#35 的默认值为空，
未设置或设置空值也保留无代理前缀；上例继续显式使用 `/`，便于核对实际部署选择。
浏览器表单、静态资源与查询链接必须与实际启动入口一致。
启动后的原状态初始化与后续只读查询分开记录，只有查询前后状态字节一致才可宣称零写入。

从独立模块 `agent-festival/go/cmd/credit-demo` 使用 Go1.25 和原锁构建可执行文件。
上面变量由部署时私有注入；RDC_REGISTRY_API_TOKEN 取本轮私有部署记录中的同名值，不在页面填写。
RDC token 至少16字符，登录密码须显式且非 demo；与 H01 查询密钥/
通道令牌、Gateway 令牌分别生成。不设置 BUILDER_PUBLICATION_ENABLED 时不加载配置、不挂新路由。

`/ui/builder` 原页面还需按 [模块 README](../../README.md) 开启 H01 配置；未齐 reader 时
仍只显示真实查询状态。publication GET 本身不依赖 H01 远端，不触发构建/登记/业务调用。
原 `POST /ui/login` 使用 form 参数 user/password/next=/ui/builder，保存 cd_session cookie，
然后携 cookie 访问 `GET /api/builder/publications/intent-gate`。未登录/撤权403、未知别名404、
错误方法405、参数或正文覆盖400；发现失败200且 available=false/status=indeterminate。
响应始终 governance=unverified、execution_authorized=false，不泄露上游正文/地址/秘密。

观察器核对原 agentId/native及语义卡摘要，并要求发现项的instanceId与serviceId均精确等于
instance_id，显式capability_version及contract匹配且ONLINE；任意其他非空实例不算匹配。
部署版本显示declared-only，不宣称原RDC响应认证它；原始卡字节digest和私有语义digest分别核验。
新版本用新原生 ID；旧 Task 固定旧 native ID/实例/版本查询，不跟随别名切换，不在回执丢失后重派。
关闭配置只影响本批发布区；不能停止共享资源、删除历史/DB 卷或声称重启恢复/取消已验证。
原 Demo IAM 不代表生产身份；正式 Hub 接收和共享治理仍未完成。本机查询或 Gateway 调用不证明
知识已生成真实图、年报 D2、专业质量、效率收益或生产资格。


历史失败不覆盖：build05/test01实际RDC458项中2项因旧SHA64版本字段超过原VARCHAR(16)失败；
Gateway独立328项通过不能替代RDC资格。修订后新build06/test02和最终双包原生结果另记，
没有完整正向报告前不宣称登记、Hub发现或业务调用已完成。


原生JSONRPC1.0卡的空repeated skills可省略；仅显式deployment-binding模式将缺省视为已有
空数组的相同语义，不改卡字节或SDK。显式null、非数组、非法/恶意skill内容仍拒绝；旧默认
agent-card模式仍要求skills字段。首次双包登记因这一遗漏兼容问题失败的原件保留，新候选另验。
