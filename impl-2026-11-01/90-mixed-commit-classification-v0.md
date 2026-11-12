# 90 · 混装提交两线正源分类表 v0（四列可多选 + 派生规则 + 钉版规则 + 双线试扫）
> **公仓适配声明（迁移轮加入）**：本表分类对象=私仓 nga 工作树（hub 为其公仓镜像）；TREE-MATCHES/列计数
> 等式须在私仓工作树复跑，本仓内不适用非缺陷。commit 锚指私仓线档案；公仓侧本批=90-96（映射见 94 号头注）；
> 生成器运行件未随迁入公仓（主线化后可入库，见 MR 正文）。


> 性质：**v0.5 草案——96 号续批轮（诚实表述修正+brief/裁剪证据 tar 入树；触发器 1 再命中见 §3.2）；v0.4=外环 R2 M-3 治理，分类仍由我方提议，夏老师核准后方可钉版**（钉版语义同 v0；前轮 v0.3 系收卷后自家触发器 1 命中处置，其「94 号入树 573→574」为带 commit 锚的历史读数，在 §3.2 记录，
> 即 75 号回执 §二.2 认领的形态：「随下轮回执附分类表 v0（定义/规格类×实现/修复类×判定侧×镜像/席位侧）+钉版规则提案，
> 供您核准」——版本虽至 v0.3，核准与钉版权限不变，我方不自钉）。
> 输入依据：70 号 §2.2 两扫描线原文（§1.1 逐字引用）；50 号 §七 B-1 修复链（已知正例回归样本）；
> 76 号任务 #4（交付定义与验收）。修复轮输入：outer-r1（M5 占位符未替换、M6 自指失灵、m 组索引/引文/J6 正则）。
> 作者：智能体乙（沙箱 `~/sandbox-agent-b`，分支 `agent-b-flank`）。日期：2026-11-11（v0=提交 `26d9856`；v0.1=修复位 R1）。
> 数字纪律：本件所有计数均为实跑粘贴；表格与规则索引皆为 §2.2 生成器逐字导出物（§2.3 TREE-MATCHES/GEN-SYNC 实证；
> v0 的 §2.1 为手抄索引，v0.1 起改为 `index` 模式逐字导出——m8 收口）。

---


## 〇、结论摘要（30 秒版）

1. **行数等式成立（v0.1 重钉活树）**：§1.4 数据行数 == `git ls-files | wc -l` == 574（v0.4 注入轮机械派生值——注入命令即该公式，复跑=§3.3 命令 1，非手抄；此前 573=`ec2ff2e` 为 v0.1 时点锚读数，§1.2 粘贴块系该时点记录，命令与输出见 §1.2
   实跑粘贴；v0 锚在 `c9b5965`=571，`da3958b`=569 为任务口径锚）。571→573 的差额=本件自身提交 `26d9856`
   ＋92 号 `d1c1943`（`git diff --name-status c9b5965..ec2ff2e` 粘贴见 §1.2）。**锚口径声明：本表钉在
   HEAD=收卷后树**（v0.1 生成于 `ec2ff2e`@573；v0.2/v0.3 随触发器 1 命中重导出，现值 574@`914c4e5`+本笔；v0.4 轮零文件增删，等式读数不变（注入轮现跑==574）；各修复位只改在档文件不增删文件，等式对当轮 commit 前树同值成立）；
   前序锚 `c9b5965`/`26d9856`/`da3958b` 只作历史对照。顺带按 git 实值更正编排者简报的 77/78 hash 互换：
   **`26d9856`＝90 号件、`c9b5965`＝91 号件**（`git show --name-only` 各恰一文件，输出粘贴见 §1.2 尾两块）。
   计数漂移是 §3 触发器 1 的真实读数（记录见 §3.2），分类表对 HEAD 漂移敏感是设计属性，不是缺陷。
2. **全表纯派生**：四列路径模式规则＋保留模式＋Z 兜底由 §2.2 生成器逐字导出为 §2.1 索引（导出 46，含异体短语与
   Z0 兜底）；整表可从 `git ls-files` 逐字节重导出（§2.3 TREE-MATCHES 实证）。列计数：定义/规格 D=87、
   实现/修复 I=170、判定侧 J=72、镜像/席位 M=262；四列全·（扫描线不判）106；≥2 列跨类 117——v0.4 注入轮现值（非手抄），复跑命令与输出见 §2.4
   （v0.1 相对 v0：D 87（本件自指行入 D9）、Z 族 105@`ec2ff2e` 时点（92 号入 Z1），其余不变；v0.4 轮 Z 族 106——94 号入 Z1，§2.4 注入轮复跑）。
3. **双线试扫（v0.1 锚现跑）**：任务规定窗口=最近 30 提交，实跑 P3=0、失同步=1（#1 `ec2ff2e` 闸 v2.0 出厂件
   [IJ] 独触入 §4.3 清单；历史锚下同窗为 0/0，对照见 §4.3 注）；30 窗口够不到 B-1 链——**此句曾作「76 号云云『够不到就扩
   窗口并声明口径』」引语，76 号两版均无该条款（grep 复验零真源），系 96 号主线审计抓出的凭空出处，本批改直述**：
   76 号验收期望=扫描线命中已知正例、未涉窗口弹性；为达该期望**我方自决**扩窗至 90 并在此声明口径
   （`git rev-list --count HEAD`=377@v0.1 锚）。结果：P3 线疑点 10，**命中已知正例病引入提交
   c89219c（P3=疑；#89@v0.1 锚、#93@v0.2 锚）**——判定规格夹具＋50 号预注册口径与 *_test.go 同批落地，题面「feat(e2e)+schema」即混装字面 ✓；
   （E-1 教训：固定 90 窗在 HEAD 前进后**静默出锚**——P3=9 系 `fb53629` 时点锚读数（v0.4 复算符合，复算路径=同 classify 窗口枚举）；注入轮 scan 90 现跑 P3=8，锚又出窗、断言又无声坠——教训句自证。v0.2 起改
   位稳窗 D+5，D=锚深实跑 93→窗 98@`fb53629`，注入轮 97→窗 102@`94fa9f1`。）失同步线疑点 6@v0.1→9@v0.2（新增 3 行=闸修复独触笔
   5f289ac/d26a810/fb53629 入窗——m15 runbook-stale 的机检显形，§4.3 v0.2 块），**对病引入提交报「—」：v0 失同步线对本仓 B-1 病体形态失灵**（病形=同批触碰镜像文件而镜像内容未共变，
   文件粒度不可见）——按任务要求如实报告，原因分析与 v1 方向见 §4.5/§4.6（这本身是本件最有价值的发现）。
   §4.2/§4.3 的疑点清单在 v0.1 已以现跑粘贴取代 v0 占位符（M5 收口）。
4. **双向基准验证**：修复提交 b02761e（#76@v0.2 锚、#72@v0.1 锚，双侧同批=治法形态）两线均「—」不误报 ✓——
   70 §2.2「两条线方向相反」在本仓同锚成立：P3 线报的病引入提交恰是失同步线的合法形态，
   单条判据必反，拆分必要。
5. **自指收口（M6）**：v0 把本表自身行归 Z1「四列全·（扫描线不判）」——机检线对表体改动结构性失明。v0.1 的 D9
   正则补 classification 面，本件行落 D 集：改表=改定义面，「表+实现同批」即 P3 疑、「表+判定面独触」即失同步疑
   （合成反例与对自身提交行的双线判定现跑粘贴见 §4.7）；表体与内嵌生成器的失同步由 §2.3 diff 探针（触发器 2）守。
   这是 70 §2.2「分类表自身按术语锚定规则管理」意图的机检落点。

---

## 一、正源分类表


### 1.1 依据原文（70 号 §2.2 逐字引用）

> | 病族 | 病形 | 扫描线（疑点判据） | 寿命 |
> |---|---|---|---|
> | **P3 偷渡族**（定义变更藏进修复提交） | 同一提交**双侧同触**（定义/规格类文件＋实现/修复类文件） | 「同一提交同时触碰两侧」→ 疑点，定义变更应独立提交独立审 | 长期 |
> | **死锁失同步族**（判定侧更新、镜像未跟） | 病态提交恰恰**单侧独触**判定侧 | 「判定侧文件变更而镜像/席位侧对应文件未同批变更」→ 疑点（同步要求的机械面） | **过渡性**——生成制落地后漂移不可表达，此线自然消亡 |
>
> 「同时触碰两侧」对病族 B 是**治法形态**（合法的原子同步提交）不是病形……两条扫描线的文件分类
> （定义/规格类、实现/修复类、判定侧、镜像/席位侧）须有正源分类表并随规格钉版，分类表自身按术语
> 锚定规则管理——**无正源分类则扫描线不可机检**。（70 §2.2）

（v0.1 勘补：v0 的引用在 P3 行漏了「文件」字样，已按 70 号原句补回——逐字引用面以 70 §2.2 正文为准；
「分类表自身按术语锚定规则管理」的机检落点＝D9 classification 面＋§4.7 自指判定，即 M6 修复位。）

### 1.2 口径与行数等式（实跑粘贴；v0.1 重钉活树）

以下均在仓库根以原文逐字执行。命令行不带 `$` 提示符前缀（本仓 claim_gate 的 CMD 识别面不含提示符——v0 粘贴行带
`$` 时整段落「引语无命令」命中，v0.1 起命令行为可直接识别的形态）：

```console
git ls-files | wc -l
573
git ls-files | python3 -I /tmp/agentb/gen77.py table | wc -l
573
git log -1 --format='%h'
ec2ff2e
git ls-tree -r c9b5965 --name-only | wc -l
571
git diff --name-status c9b5965..ec2ff2e
M	agent-festival/tools/claim_gate.py
A	impl-2026-11-01/90-mixed-commit-classification-v0.md
A	impl-2026-11-01/92-state-vocabulary-alignment.md
git show --name-only --format='%h' 26d9856 | tail -1
impl-2026-11-01/90-mixed-commit-classification-v0.md
git show --name-only --format='%h' c9b5965 | tail -1
impl-2026-11-01/91-counterexample-acceptance-mapping.md
```

等式：**§1.4 数据行数 == `git ls-files | wc -l`**——等式即公式，现跑读数在此不落裸数（指针：复跑=§3.3 命令 1；上方粘贴块 573/`ec2ff2e` 为 v0.1 时点记录；注入轮读数 574@`94fa9f1` 见 §1.4 题头与页脚）。
行提取域=TABLE-START/TABLE-END 独占整行的唯一标记区间，提取器=§2.3 的 python 件 `table_rows.py`（不用 sed 区间——
BSD sed 对起讫同线自匹配的行会跨区间重开，本位实测踩坑后改用标记行独占整行的判定）；区间外的表形输出行
（§2.1 判别例、§4.7 自指行）不计入数据行——v0 的全文 `grep '^| `' 口径在 v0.1 加入区间粘贴后收紧为标记区间口径
（探针本体修订，对原数据行逐字等价，对照粘贴见 §2.3）。

### 1.3 四列语义（v0 提议定义）

| 列 | 记号 | 语义（疑点判据侧的角色，非文件内容摘要） |
|---|---|---|
| 定义/规格类 | D | 规范性定义真源：契约 schema、判定规格夹具、预注册/口径正文、立法文本、冻结口径面 |
| 实现/修复类 | I | 被规约的实现体：生产码＋**本仓特殊性：P0-A 被测实现 s1Judge/verify 在 *_test.go 内**（50 §U 原文），故测试码亦入 I |
| 判定侧 | J | 出判定/断言的机械面：断言矩阵、文字钉、闸（claim_gate）、双校验器 oracle、核验脚本 |
| 镜像/席位侧 | M | 他面的投影且靠同步维系的副本：一致性夹具、再生归档（testdata）、band mirror（verifier）、台账、孪生件、发送态文书、GitCode 镜像席位、消费端席位 |
| （全 ·） | — | 非双线辖域（通信档/analysis-only/配置面），扫描线对其永不判 |

跨类允许（任务规定「四列可多选」）：如 `verifier_test.go` = I∩J∩M（被测实现∩断言载体∩席位镜像——
B-1 病体正跨此类），`judgment-spec-credit.json` = D∩M（期望本体∩手写镜像）。


### 1.4 全表（行数=`git ls-files | wc -l` 注入轮现跑派生 574@`94fa9f1`（非手抄，指针=§3.3 命令 1）；逐字节可由 §2.2 生成器重导出；等式实跑见 §2.3）

<!--TABLE-START-->
| 文件（git ls-files 全量） | D 定义/规格 | I 实现/修复 | J 判定侧 | M 镜像/席位 | 依据（规则号·短语） |
|---|---|---|---|---|---|
| `.github/GITCODE-MIRROR.md` | · | · | · | ✓ | M4·GitCode 镜像席位维护面 |
| `.github/scripts/sync_gitcode.py` | · | ✓ | · | · | I5·镜像同步实现 |
| `.github/scripts/test_sync_gitcode.py` | · | ✓ | ✓ | · | I5·镜像同步实现；J2·pytest 断言载体 |
| `.github/workflows/sync-gitcode.yml` | · | · | · | ✓ | M4·GitCode 镜像席位维护面 |
| `.gitignore` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `CLAUDE.md` | ✓ | · | · | · | D7·全仓共享契约/立法（多智能体合同） |
| `README.md` | ✓ | · | · | · | D10·口径正文（readme_pin 源码文字钉的被钉面） |
| `achitecture design/05-architecture-brief-and-decisions.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/06-architecture-diagrams-md.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/07-nga-architecture-analysis.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/08-nga-architecture-v0.1.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/09-nga-architecture-diagrams.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/10-credit-decision-walkthrough.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/11-jev-agentic-orchestration.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/12-replay-and-audit-contract.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/13-nga-architecture-v0.2.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/14-v0-adversarial-review.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/15-graph-loop-workflow-control-review.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `achitecture design/README.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `agent-festival/.gitignore` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `agent-festival/README.md` | ✓ | · | · | · | D10·口径正文（readme_pin 源码文字钉的被钉面） |
| `agent-festival/docs/2026-11-01-collision-response-archived.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `agent-festival/docs/attack-corpus.md` | · | · | · | ✓ | M5·台账归档（corpus:NN 被代码/回函交叉引用=状态投影面） |
| `agent-festival/docs/conformance-fixture-format.md` | ✓ | · | · | · | D8·立法文本（夹具格式/守卫模式/runbook/先验） |
| `agent-festival/docs/fixture-guard-pattern.md` | ✓ | · | · | · | D8·立法文本（夹具格式/守卫模式/runbook/先验） |
| `agent-festival/docs/minor-ledger.md` | · | · | · | ✓ | M5·台账归档（corpus:NN 被代码/回函交叉引用=状态投影面） |
| `agent-festival/docs/review-loop-runbook.md` | ✓ | · | · | · | D8·立法文本（夹具格式/守卫模式/runbook/先验） |
| `agent-festival/examples/java-consumer/.gitignore` | · | · | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现） |
| `agent-festival/examples/java-consumer/README.md` | · | · | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现） |
| `agent-festival/examples/java-consumer/pom.xml` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/run.sh` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/src/main/java/spike/FixtureLoader.java` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/src/main/java/spike/LedgerChecker.java` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/src/main/java/spike/Main.java` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/src/main/java/spike/PinAudit.java` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/src/main/java/spike/SchemaChecker.java` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/examples/java-consumer/src/main/java/spike/SuiteRunner.java` | · | ✓ | · | ✓ | M6·消费端席位参考实现（同契约第二独立实现）；I8·席位侧代码实现 |
| `agent-festival/go/00-priors-compose-sign.md` | ✓ | · | · | · | D8·立法文本（夹具格式/守卫模式/runbook/先验） |
| `agent-festival/go/cmd/adopt/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/adopt/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/compose/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/compose/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/evidlog/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/evidlog/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/judgment/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/judgment/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/ledger/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/ledger/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/release/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/release/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/replay/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/replay/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/sign/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/sign/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/cmd/validate/main.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/cmd/validate/main_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/go.mod` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `agent-festival/go/go.sum` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `agent-festival/go/internal/adopt/adopt.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/adopt/adopt_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/adopt/check.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/compose/compose.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/compose/compose_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/compose/wiring_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/correctness_probe_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/doc.go` | ✓ | · | · | · | D11·docstring 正文（文字钉锚定面） |
| `agent-festival/go/internal/e2e/e2e_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/e7_canary_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/health_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/p0a_window_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/t1_llm_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/t2host_rk_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/t2host_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/testdata/p0a_scan.tsv` | · | · | · | ✓ | M2·扫描归档（再生型投影，--rerun 等值检验面） |
| `agent-festival/go/internal/e2e/verifier_test.go` | · | ✓ | ✓ | ✓ | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体；M3·席位侧 band mirror 载体（verify=规则语义手写投影，B-1 病体） |
| `agent-festival/go/internal/e2e/x3_entry_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/e2e/zz_xs2_demo_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/evidlog/evidlog.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/evidlog/evidlog_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/judgment/check.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/judgment/judgment_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/ledger/docstring_pin_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/ledger/grant_lockstep_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/ledger/ledger.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/ledger/ledger_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/ledger/load.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/ledger/readme_pin_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/release/check.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/release/release_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/replay/replay.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/replay/replay_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/rfc8785/differential_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/rfc8785/rfc8785.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/schemalib/schemalib.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/schemalib/schemalib_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/suite/suite.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/suite/suite_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/validate/helpers.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/validate/schemalib.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/validate/schemalib_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/go/internal/validate/validate.go` | · | ✓ | · | · | I1·生产实现（cmd/internal 非测试码） |
| `agent-festival/go/internal/validate/wire_test.go` | · | ✓ | ✓ | · | I2·被测实现（s1Judge/verify 为测试内确定性函数，50§U）；J1·断言矩阵/探针/文字钉载体 |
| `agent-festival/schema/.gitkeep` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `agent-festival/schema/README.md` | ✓ | · | · | · | D10·口径正文（readme_pin 源码文字钉的被钉面） |
| `agent-festival/schema/_suite.py` | · | ✓ | · | · | I3·Python 侧实现 |
| `agent-festival/schema/adoption.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/adoption.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/capability-contract.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/compose.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/contract-envelope.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/event-ledger.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/evidlog.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/evidlog.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/examples/adoption/invalid/bad-grant-effect.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/adoption/invalid/empty-grants.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/adoption/invalid/malformed-digest.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/adoption/invalid/missing-versions.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/adoption/invalid/unknown-extra-field.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/adoption/valid/adoption-request-credit.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/adoption/valid/adoption-request-minimal.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/caller-holds-absent.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/caller-holds-encoded-unreserved.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/caller-holds-invalid-utf8.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/caller-holds-noncanonical-escape.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/chain-false-declarations.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/cost-not-exactly-summable.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/credit-budget-exceeds-parent.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/credit-op1-effect-escalation.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/credit-op1-grant-not-held.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/credit-s1-judge-not-read-only.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/delegate-outside-graph.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/edge-three-items.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/effect-escalation.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/empty-par-group.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/extra-action.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/governance-pipeline-not-held.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/infinite-parent-budget.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/mixed-cost-units.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/node-fails-contract-gate.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/node-scheduled-twice.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/orphan-in-edge.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/orphan-node.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/parent-cost-unit-mismatch.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/phantom-upstream.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/plan-dual-mode-nested.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/plan-two-modes.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/selector-not-held.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/seq-time-overflow.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/split-budget-overflow.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/split-count-not-exactly-divisible.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/split-rounding-edge.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/supply-edge-backwards.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/supply-edge-in-par.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/upstream-action-escalation.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/upstream-effect-escalation.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/invalid/upstream-without-edge.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/chain-closure-held.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/chain-transitive-upstream.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/compensation-any-none.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/credit-slice-a001.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/governance-publish-contract.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/integer-bound-with-decimal-point.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/node-id-trailing-newline.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/parallel-split-budget-ok.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/parallel-split-inexact-share.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/compose/valid/same-key-worst-effect.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/envelopes/invalid/alg-mismatch.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/alg-none.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/bad-typ.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/crit-header.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/duplicate-signer.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/enc-key.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/endorser-only.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/hash-mismatch.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/key-ops-no-verify.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/lone-surrogate.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/no-signatures.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/one-of-two-bad.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/signed-but-invalid-contract.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/tampered-contract.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/unknown-kid.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/invalid/wrong-owner.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/valid/compliance-filer.envelope.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/valid/doc-extractor.envelope.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/valid/literal-fffd-header-surrogate.envelope.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/envelopes/valid/literal-fffd.envelope.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）；envelopes 面由 sign.py 机械再生 |
| `agent-festival/schema/examples/evidlog/invalid/bad-timestamp.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/evidlog/invalid/seq-float-beyond-2e53.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/evidlog/invalid/seq-float-syntax.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/evidlog/invalid/seq-gap.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/evidlog/invalid/seq-int64-overflow.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/evidlog/invalid/unknown-field.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/evidlog/valid/dense-seq.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/actions-duplicated.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/assurance-as-fifth-section.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/bound-negative.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/bound-zero-seconds.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/delegate-non-agent.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/effect-not-in-vocabulary.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/effects-below-requires.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/encoded-unreserved.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/encoded-wildcard-lowercase.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/encoded-wildcard.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/fixture-pin-injection.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/fullwidth-wildcard.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/grant-extra-key.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/invalid-utf8-escape.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/missing-compensation.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/nan-cost.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/noncanonical-percent-escape.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/padded-wildcard.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/read-only-with-effect.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/selector-double-trailing-newline.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/subject-id-empty.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/wildcard-selector.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/invalid/zero-width-wildcard.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/bad-candidate-semantics.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/bad-example-kind.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/bad-level-type.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/empty-levels.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/empty-term-name.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/empty-terms.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/extra-field-in-ownership.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-candidate-set.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-conventions.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-evidence-sources.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-examples.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-levels.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-ownership.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-question-id.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-question.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-resolution.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-rules.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/missing-scope.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/invalid/unknown-extra-field.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/valid/judgment-spec-credit.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/judgment/valid/judgment-spec-minimal.json` | ✓ | · | · | ✓ | D2·判定规格夹具本体（期望表随 50 号冻结）；M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/keys/invalid/binding-unknown-kid.jwks.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/keys/invalid/duplicate-kid.jwks.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/keys/invalid/missing-kid.jwks.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/keys/trusted.jwks.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/adopted-without-grants.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/adopted-without-idempotency.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/bad-grant-in-event.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/bypass-effect-without-adoption.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/conflict-with-idempotency.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/control-before-adoption.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/control-effect-status-transition.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/control-forged-attribution.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/control-with-idempotency.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/double-adopted.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-attribution-mismatch.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-completed-without-dispatched.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-dispatched-after-unresolved.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-double-completed.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-reunresolved.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-revive-after-terminal.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-with-grants.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-with-rejected-status.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/effect-without-idempotency.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/event-missing-outcome.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/idem-key-mismatch.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/malformed-digest.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/missing-model-version.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/non-monotonic-seq.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/orphan-reused.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/reused-with-grants.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/same-key-different-digest.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/seq-2e53-adjacent.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/seq-float-beyond-2e53.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/seq-float-syntax.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/seq-int64-overflow.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/seq-odd-float-twin.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/invalid/status-family-mismatch.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/chain-dispatched-failed.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/chain-dispatched-unresolved.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/chain-unresolved-failed.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/concurrent-double-adoption.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/crash-unresolved-then-complete.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/credit-a001-chain.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/dispatch-refused-then-dispatched.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/high-seq-across-2e53.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/ledger/valid/retry-reuse.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/empty-model-versions.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/empty-release-id.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/invented-in-flight-mode.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-agent-identity.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-correction-contract.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-evidence-separation.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-evidence.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-in-flight-choice.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-references.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-release-id.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-rollback.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/missing-rollout.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/nested-smuggle-in-references.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/invalid/unknown-extra-field.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/valid/release-credit-v2.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/valid/release-hold-variant.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/release/valid/release-minimal.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/divergence-float-syntax.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/divergence-int64-overflow.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/divergence-without-at.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/invented-op.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/match-with-at.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/missing-conformance-pin.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/negative-divergence-at.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/unknown-case-field.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/invalid/wrong-case-version.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/happy-chain.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/mutated-input-located.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/race-outcome-a-first.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/race-outcome-b-first.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/recorded-violation-forks.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/tampered-conclusion.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/unexpected-parameter-refused.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/replay/valid/unknown-field-in-record.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/compliance-filer.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/doc-extractor.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/encoded-iri-path.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/encoded-opaque-segments.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/notes-nan-string.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/owned-side-effect.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/examples/valid/selector-trailing-newline.json` | · | · | · | ✓ | M1·一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） |
| `agent-festival/schema/execution-graph.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/judgment.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/judgment.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/ledger.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/release.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/release.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/replay-case.schema.json` | ✓ | · | · | · | D1·JSON-Schema 契约（字段/类型定义真源） |
| `agent-festival/schema/replay.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/schema/requirements.txt` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `agent-festival/schema/rfc8785_min.py` | · | ✓ | · | · | I3·Python 侧实现 |
| `agent-festival/schema/sign.py` | · | ✓ | · | · | I3·Python 侧实现 |
| `agent-festival/schema/validate.py` | · | ✓ | ✓ | · | I3·Python 侧实现；J4·双校验器 Python 面（oracle 真判定） |
| `agent-festival/tools/claim_gate.py` | · | ✓ | ✓ | · | I4·闸/清单工具实现；J3·机检判定器（claim/corpus 闸） |
| `agent-festival/tools/corpus_manifest.py` | · | ✓ | ✓ | · | I4·闸/清单工具实现；J3·机检判定器（claim/corpus 闸） |
| `design/nga-overall-design.md` | ✓ | · | · | · | D6·架构规格/源快照（05/06 保存原件为权威锚） |
| `experiments/ENV-VERL.md` | · | · | · | · | Z5·环境/服务器笔记/索引——四列全·（扫描线不判） |
| `experiments/KEV-PROTOCOL.md` | ✓ | · | · | · | D4·预注册/协议/参考值（metrics·thresholds 规范性） |
| `experiments/R0-BASELINE-PREREG.md` | ✓ | · | · | · | D4·预注册/协议/参考值（metrics·thresholds 规范性） |
| `experiments/R0-BASELINE-RESULTS.md` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/R3-REFERENCE-TARGETS.md` | ✓ | · | · | ✓ | D4·预注册/协议/参考值（metrics·thresholds 规范性）；M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/R3-SELECTION-CONSUMER-REPAIR.json` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/R3-SELECTION-CONSUMER-REPAIR.md` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/R3-SFT-PREREG.md` | ✓ | · | · | · | D4·预注册/协议/参考值（metrics·thresholds 规范性） |
| `experiments/README.md` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `experiments/REFERENCE-EVALUATION-QUEUE-2026-10-30.md` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/REFERENCE-INFERENCE-REVIEW-2026-10-30.md` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/REFERENCE-TEST-ADDITIONAL-2026-10-30.json` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/REFERENCE-TEST-INVENTORY-2026-10-30.json` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/REFERENCE-TEST-RESULTS-2026-10-30.md` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/SERVER-175.md` | · | · | · | · | Z5·环境/服务器笔记/索引——四列全·（扫描线不判） |
| `experiments/build_tasks.py` | · | · | · | · | Z5·环境/服务器笔记/索引——四列全·（扫描线不判） |
| `experiments/manifest.json` | · | · | · | ✓ | M8·运行归档/结果快照（sha256+计数=run 派生物） |
| `experiments/r0_baseline/.gitignore` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `experiments/r0_baseline/common.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/decision.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/freeze_sample.py` | ✓ | ✓ | · | · | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体 |
| `experiments/r0_baseline/metrics.py` | ✓ | ✓ | · | · | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体 |
| `experiments/r0_baseline/prompt.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/report.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/run_baseline.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/run_sharded.sh` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/scorer.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r0_baseline/split.py` | ✓ | ✓ | · | · | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体 |
| `experiments/r0_baseline/test_harness.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r0_baseline/thresholds.py` | ✓ | ✓ | · | · | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体 |
| `experiments/r3_sft/README.md` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `experiments/r3_sft/r1_code/build_sft.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r1_code/check_scorer.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r1_code/common.py` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/convert_kev.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r1_code/decision.py` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/freeze_sample.py` | ✓ | ✓ | · | ✓ | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/metrics.py` | ✓ | ✓ | · | ✓ | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/prompt.py` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/report.py` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/run_baseline.py` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/run_check.sh` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r1_code/run_sharded.sh` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/scorer.py` | · | ✓ | · | ✓ | I6·实验实现；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/split.py` | ✓ | ✓ | · | ✓ | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/test_build_sft.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r1_code/test_convert_kev.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r1_code/test_harness.py` | · | ✓ | ✓ | ✓ | I6·实验实现；J2·校验/断言脚本；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r1_code/thresholds.py` | ✓ | ✓ | · | ✓ | D5·预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）；I6·同文件实现体；M9·手写孪生镜像（与 r0_baseline 同名件，双份手改） |
| `experiments/r3_sft/r3_code/bench_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/build_kev_sft.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/check_lora_freeze.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/core_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/degenerate_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/diag_fp32.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/diag_padding.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/env_check.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/eval_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/inspect_ckpt.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/k2_loss.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/k2_stats.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/kev_aug.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/launch_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/launch_sft.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/merge_checkpoint.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/merge_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/nga_lora_dropout_patch.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/nga_onecycle_patch.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/prechecks_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/provenance_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/r0rescore_verify.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/r3_sft_dataset.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/runner_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/select_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/selection_evidence_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/test_k2_loss.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/test_provenance_boundary.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/test_r3_data.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/test_select_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/test_selection_evidence_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/test_train_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/testlock_r3.py` | · | ✓ | ✓ | · | I6·实验实现；J2·校验/断言脚本 |
| `experiments/r3_sft/r3_code/track_t_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `experiments/r3_sft/r3_code/train_r3.py` | · | ✓ | · | · | I6·实验实现 |
| `impl-2026-10-31/00-priors-six-pieces.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-10-31/01-piece0-identity-contract.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/00-priors-alignment-batch.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/01-priors-piece6-replay.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/03-priors-exec-batch.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/04-collision.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/04-exec-redteam.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/05-edge-grammar-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/06-xingchao-comments-collision.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/07-response-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/08-blank-collision-round2.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/09-reconciliation.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/10-blank-collision-round3.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/11-reconciliation-round3.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/12-fix-round-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/13-c4-revision-ordering-note.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/14-audit-loop-log.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/15-t1-readout.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/16-t1-run2.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/17-blank-r6.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/18-blank-r9.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/19-artifact-a-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/19-artifact-a-sent.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/20-xingchao-f02abc.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/21-f02abc-reply-sent.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/22-briefing-to-xia.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/23-collision-xingchao-round4.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/24-blank-collision-round4.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/25-reconciliation-round4.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/26-d01-05-reply-sent.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/27-blank-collision-correctness.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/28-blank-collision-correctness-r2.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/29-reconciliation-correctness.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/30-blank-r14.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/31-rk-scope-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/32-blank-r14p.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/33-fullstate-collision.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/34-firstbatch-scope-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/35-r19-battlefield-freeze.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/36-sixexperiments-scope-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/37-sixexperiments-readout.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/38-issue1-five-questions-reply-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/39-xia-receipt-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/40-liu-questions-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/41-xia-r4-collision-report.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/42-review-meta-analysis-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/42-review-meta-analysis-v2.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/43-experiment-proposal-matrix.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/44-collision-xingchao-xia-experiments.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/45-collision-round2-blank.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/46-collision-integrated.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/47-reply-xingchao-r5-archived.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/48-receipt-xia-r4-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/49-receipt-xia-r4-formal.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/50-p0a-scope-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/51-architecture-collision-round3.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/52-architecture-collision-round4.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/53-collision-merged-51-52.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/54-singularity-elevation.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/55-experiment-backlog-from-54.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/56-reply-xingchao-r6-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/56-reply-xingchao-r6-final.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/56-reply-xingchao-r6-sent.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/57-receipt-xia-r5-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/58-receipt-xia-r5-formal.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/59-handoff-n3-unified-entry.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/60-n3-scope-prior.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/61-correction-xingchao-numbers.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/61-correction-xingchao-sent.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/62-input-xia-r6-ruling.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/63-collision-xia-r6-report.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/64-agent-hub-alpha-survey.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/65-mr-n3-agent-hub-sent.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/66-collision-three-source-report.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/67-xs3-preregister.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/68-receipt-xia-r6-draft.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/68-receipt-xia-r6-formal.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/69-handoff-skill-backmerge.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/70-input-xia-r7-ruling.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/71-collision-four-source.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/72-handoff-four-source-collision.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/73-fs72-independent-report.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/73-fs72-workspace.tar.gz` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/74-draft-reply-agent-hub.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/75-receipt-xia-r7-formal.md` | · | · | · | ✓ | M7·发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） |
| `impl-2026-11-01/76-handoff-agent-b-flank.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/90-mixed-commit-classification-v0.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
| `impl-2026-11-01/91-counterexample-acceptance-mapping.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/92-state-vocabulary-alignment.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/93-draft-collaborator-comments.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/94-completion-report-agent-b.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/95-recon-agent-b.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/96-brief-agent-b-restatement-and-mr.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/BATTLEFIELD.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/agentb-evidence-2026-11-12.tar.gz` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/fs72-evidence-2026-11-12.tar.gz` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `impl-2026-11-01/r1-exec-review.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
| `proposal/00-research-proposal.md` | ✓ | · | · | · | D3·工程计划/决策记录（规范面） |
| `proposal/01-work-plan-four-directions.md` | ✓ | · | · | · | D3·工程计划/决策记录（规范面） |
| `proposal/02-status-decisions-next-stage-2026-10-30.md` | ✓ | · | · | · | D3·工程计划/决策记录（规范面） |
| `proposal/03-status-decisions-and-next-plan-2026-10-30.md` | ✓ | · | · | · | D3·工程计划/决策记录（规范面） |
| `proposal/04-reference-test-first-decisions-2026-10-30.md` | ✓ | · | · | · | D3·工程计划/决策记录（规范面） |
| `proposal/05-instruction-order-and-execution-2026-10-30.md` | ✓ | · | · | · | D3·工程计划/决策记录（规范面） |
| `replies/2026-11-01-festival-collision-response.md` | · | · | · | · | Z1·通信归档——四列全·（扫描线不判） |
| `survey/01-object-and-definitions.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/02-propositions-and-core-logic.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/03-semantics-can-the-model-read-the-rules.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/04-consistency-the-primary-metric.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/05-system-business-goal-and-positions.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/06-learning-objectives-data-and-outputs.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/07-evaluation-coverage-and-gaps.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/08-critical-and-adversarial-review.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/09-references.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/README.md` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/nga-survey.pdf` | · | · | · | · | Z2·analysis-only 报告面——四列全·（扫描线不判） |
| `survey/tools/build_pdf.py` | · | ✓ | · | · | I7·报告构建工具 |
| `survey/tools/check_overflow.py` | · | ✓ | · | · | I7·报告构建工具 |
| `tools/README.md` | · | · | · | · | Z4·配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） |
| `tools/build_reference_tables.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/compare_reference_kev.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/export_reference_test_metrics.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/openjev_reference_probe.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/reaggregate_jevify_reference.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/reference_finish.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/reference_kev_eval.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/reference_kev_scale_probe.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/reference_native_kev.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/reference_recovery_worker.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/resume_reference_kev.py` | · | ✓ | · | · | I5·工程工具实现 |
| `tools/test_reference_finish.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/test_reference_kev_eval.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/test_reference_recovery.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/test_resume_reference_kev.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/test_verify_reference_kev.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/test_watchdog.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/verify_model_dedup.sh` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/verify_reference_kev.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `tools/watchdog.py` | · | ✓ | ✓ | · | I5·工程工具实现；J2·机械核验脚本（判定面） |
| `wiki/E1-gdn-fast-ref-gate.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E10-reference-recovery-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E11-reference-kev9-probe-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E12-reference-kev9-test-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E13-openjev2-native-probe-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E2-scorer-admissibility-fp32.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E3-k0-readout-precondition.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E6-reference-inference-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E7-reference-kev-test-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E8-reference-kev-scale-probe-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/E9-reference-kev4-test-20261030.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/F1-fetch-qwen3-0.6b.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/K0-untuned-readout.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/R0S-r0-side-fp32-rescore.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
| `wiki/README.md` | ✓ | · | · | · | D12·记录法口径（CLAUDE.md §8 规则与格式） |
| `wiki/TEMPLATE.md` | ✓ | · | · | · | D12·记录法口径（CLAUDE.md §8 规则与格式） |
| `wiki/W1-wave1-first-weight-updates.md` | · | · | · | · | Z3·实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） |
<!--TABLE-END-->

---

## 二、派生规则（机检生成器）

### 2.1 规则索引表（人读索引；**可执行真源=§2.2 生成器本体**，本索引由生成器源码逐字抽取）

v0 本节为手抄，与生成器存在短语/异体缺列之差（outer-r1 m 组 m8：异体短语未列、Z 行系扩写）；v0.1 起整节为
生成物——下表逐字来自 `python3 -I /tmp/agentb/gen77.py index` 的标准输出（规则号与短语逐字节=§2.2 字符串；行号列
指 §2.2 代码块内行号，shebang 起算——gen77.py 即 §2.2 的同步运行件，§2.3 GEN-SYNC 实证）。手改本节=违 §五-5。

| 规则号 | 列 | 依据短语（§2.2 逐字导出） | §2.2 行号 |
|---|---|---|---|
| D1 | D | JSON-Schema 契约（字段/类型定义真源） | :28 |
| D2 | D | 判定规格夹具本体（期望表随 50 号冻结） | :30 |
| D3 | D | 工程计划/决策记录（规范面） | :32 |
| D4 | D | 预注册/协议/参考值（metrics·thresholds 规范性） | :35 |
| D5 | D | 预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在） | :38 |
| I6 | I | 同文件实现体 | :38 |
| D6 | D | 架构规格/源快照（05/06 保存原件为权威锚） | :40 |
| D7 | D | 全仓共享契约/立法（多智能体合同） | :42 |
| D8 | D | 立法文本（夹具格式/守卫模式/runbook/先验） | :47 |
| D9 | D | 冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） | :49 |
| D10 | D | 口径正文（readme_pin 源码文字钉的被钉面） | :51 |
| D11 | D | docstring 正文（文字钉锚定面） | :53 |
| D12 | D | 记录法口径（CLAUDE.md §8 规则与格式） | :55 |
| I1 | I | 生产实现（cmd/internal 非测试码） | :58 |
| I2 | I | 被测实现（s1Judge/verify 为测试内确定性函数，50§U） | :60 |
| J1 | J | 断言矩阵/探针/文字钉载体 | :61 |
| I3 | I | Python 侧实现 | :63 |
| J4 | J | 双校验器 Python 面（oracle 真判定） | :64 |
| I4 | I | 闸/清单工具实现 | :66 |
| J3 | J | 机检判定器（claim/corpus 闸） | :66 |
| I5 | I | 工程工具实现 | :69 |
| J2 | J | 机械核验脚本（判定面） | :71 |
| I5 | I | 镜像同步实现 | :73 |
| J2 | J | pytest 断言载体 | :74 |
| I6 | I | 实验实现 | :77 |
| J2 | J | 校验/断言脚本 | :79 |
| M9 | M | 手写孪生镜像（与 r0_baseline 同名件，双份手改） | :81 |
| I7 | I | 报告构建工具 | :83 |
| M1 | M | 一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由） | :86 |
| M2 | M | 扫描归档（再生型投影，--rerun 等值检验面） | :89 |
| M3 | M | 席位侧 band mirror 载体（verify=规则语义手写投影，B-1 病体） | :91 |
| M4 | M | GitCode 镜像席位维护面 | :93 |
| M5 | M | 台账归档（corpus:NN 被代码/回函交叉引用=状态投影面） | :95 |
| M6 | M | 消费端席位参考实现（同契约第二独立实现） | :97 |
| I8 | I | 席位侧代码实现 | :98 |
| M7 | M | 发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1） | :102 |
| M8 | M | 运行归档/结果快照（sha256+计数=run 派生物） | :105 |
| J5 | J | review/ 探针（本仓 0 命中——hub 仓探针落点保留） | :108 |
| J6 | J | 验收表（跨仓保留——08 表在 inspect 仓，本仓 0 命中；前缀+词干+.md 全锚，91 号 acceptance 标题不误入） | :110 |
| Z2 | · | analysis-only 报告面——四列全·（扫描线不判） | :112 |
| Z3 | · | 实验记录（证据文档，claim 面非双线辖域）——四列全·（扫描线不判） | :113 |
| Z1 | · | 通信归档——四列全·（扫描线不判） | :114 |
| Z1 | · | 通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） | :115 |
| Z4 | · | 配置/锁文件/索引/二进制（构建面）——四列全·（扫描线不判） | :117 |
| Z5 | · | 环境/服务器笔记/索引——四列全·（扫描线不判） | :118 |
| Z0 | · | 未归入四侧——四列全·（扫描线不判） | :119 |

（导出另向 stderr 打一行 RULE-ROWS 计数；Z 行短语含「四列全·（扫描线不判）」后缀，与 §1.4 表 basis 列逐字一致。）

保留模式声明：J5（review/ 探针）、J6（验收表）为跨仓落点的模式预留（70 §2.2「无正源分类则扫描线不可机检」
要求分类表覆盖两线全部辖域面，跨仓形态变体不应另起炉灶）。m 收口：J6 初版宽形（裸 acceptance 子串匹配）会误命中
91 号文书标题（建造中实测，out-4 已登记），v0.1 在 §2.2 内以全锚正则收紧：`^08-[A-Za-z0-9._-]*acceptance[A-Za-z0-9._-]*\.md$`
——散文口径「08-*acceptance*」的精确化：前缀＋词干＋扩展名三段全锚。双向判别例（本仓树外合成路径，现跑）：

```console
python3 -I /tmp/agentb/gen77.py table <<'EOF'
08-counterexample-acceptance-mapping.md
08-acceptance-matrix.md
impl-2026-11-01/91-counterexample-acceptance-mapping.md
EOF
| `08-counterexample-acceptance-mapping.md` | · | · | ✓ | · | J6·验收表（跨仓保留——08 表在 inspect 仓，本仓 0 命中；前缀+词干+.md 全锚，91 号 acceptance 标题不误入） |
| `08-acceptance-matrix.md` | · | · | ✓ | · | J6·验收表（跨仓保留——08 表在 inspect 仓，本仓 0 命中；前缀+词干+.md 全锚，91 号 acceptance 标题不误入） |
| `impl-2026-11-01/91-counterexample-acceptance-mapping.md` | · | · | · | · | Z1·通信/碰撞/回执工作档（非双线辖域）——四列全·（扫描线不判） |
```

（正例=词干与扩展名齐全即入 J6，即便标题抄自 91 号；负例=真 91 号件无 08- 前缀，判 Z1 不误入——锚的是前缀，不是词干。）


### 2.2 生成器嵌入（唯一真源；`python3 -I` 隔离运行，仅标准库）

与运行件 `/tmp/agentb/gen77.py` 逐字同——「表是生成物，散文是宿主」：改本节必须同步运行件，两者漂移由 §2.3 GEN-SYNC 探针报红。

```python
#!/usr/bin/env python3
"""90 号 v0 正源分类表生成器（双线试扫共用同一 classify()）。

用法:
  git ls-files | python3 -I gen77.py table     # 每行 -> 表格行（4 列可多选 + 规则号依据）
  python3 -I gen77.py scan N [OFFSET]          # HEAD 起 N 个提交双线试扫
  python3 -I gen77.py index                    # 规则索引逐字导出（§2.1 生成物；行号=本文件行号）
规则集=本文件唯一真源；表格任何增行必须经本脚本重导出（机检面=§3 钉版规则触发器 2/3）。
"""
import re, sys, subprocess

ORACLES = {f"{n}.py" for n in ["adoption", "compose", "evidlog", "judgment",
                               "ledger", "release", "replay", "validate"]}
R0_TWIN = {"common.py", "decision.py", "freeze_sample.py", "metrics.py", "prompt.py",
           "report.py", "run_baseline.py", "scorer.py", "split.py", "test_harness.py",
           "thresholds.py", "run_sharded.sh"}

def classify(p):
    """-> (cols:set, [(rule,phrase)...])"""
    b = p.rsplit("/", 1)[-1]
    cols, why = set(), []
    D, I, J, M = "D", "I", "J", "M"
    def hit(c, r, ph):
        if c not in cols: cols.add(c); why.append((r, ph))
        elif not any(x[0] == r for x in why): why.append((r, ph))
    # ---- D 定义/规格 ----
    if p.startswith("agent-festival/schema/") and b.endswith(".schema.json"):
        hit(D, "D1", "JSON-Schema 契约（字段/类型定义真源）")
    if p.startswith("agent-festival/schema/examples/judgment/"):
        hit(D, "D2", "判定规格夹具本体（期望表随 50 号冻结）")
    if p.startswith("proposal/"):
        hit(D, "D3", "工程计划/决策记录（规范面）")
    if p.startswith("experiments/") and "/" not in p[12:] and \
       re.search(r"(PREREG|PROTOCOL|TARGETS)", b):
        hit(D, "D4", "预注册/协议/参考值（metrics·thresholds 规范性）")
    if p.startswith("experiments/") and b in \
       {"thresholds.py", "metrics.py", "split.py", "freeze_sample.py"}:
        hit(D, "D5", "预注册参数载体（thresholds/metrics/split/freeze_sample 以代码形态存在）"); hit(I, "I6", "同文件实现体")
    if p.startswith("achitecture design/") or p == "design/nga-overall-design.md":
        hit(D, "D6", "架构规格/源快照（05/06 保存原件为权威锚）")
    if p == "CLAUDE.md":
        hit(D, "D7", "全仓共享契约/立法（多智能体合同）")
    if p in {"agent-festival/docs/conformance-fixture-format.md",
             "agent-festival/docs/fixture-guard-pattern.md",
             "agent-festival/docs/review-loop-runbook.md",
             "agent-festival/go/00-priors-compose-sign.md"}:
        hit(D, "D8", "立法文本（夹具格式/守卫模式/runbook/先验）")
    if p.startswith("impl-2026-11-01/") and re.search(r"(prior|prereg|attlefield|input-xia|classification)", b):
        hit(D, "D9", "冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面）")
    if p in {"README.md", "agent-festival/README.md", "agent-festival/schema/README.md"}:
        hit(D, "D10", "口径正文（readme_pin 源码文字钉的被钉面）")
    if p.startswith("agent-festival/go/") and b == "doc.go":
        hit(D, "D11", "docstring 正文（文字钉锚定面）")
    if p in {"wiki/README.md", "wiki/TEMPLATE.md"}:
        hit(D, "D12", "记录法口径（CLAUDE.md §8 规则与格式）")
    # ---- I 实现/修复 ----
    if p.startswith("agent-festival/go/") and b.endswith(".go") and not b.endswith("_test.go"):
        if b != "doc.go": hit(I, "I1", "生产实现（cmd/internal 非测试码）")
    if p.startswith("agent-festival/go/") and b.endswith("_test.go"):
        hit(I, "I2", "被测实现（s1Judge/verify 为测试内确定性函数，50§U）")
        hit(J, "J1", "断言矩阵/探针/文字钉载体")
    if p.startswith("agent-festival/schema/") and b.endswith(".py"):
        hit(I, "I3", "Python 侧实现")
        if b in ORACLES: hit(J, "J4", "双校验器 Python 面（oracle 真判定）")
    if p.startswith("agent-festival/tools/") and b.endswith(".py"):
        hit(I, "I4", "闸/清单工具实现"); hit(J, "J3", "机检判定器（claim/corpus 闸）")
    if p.startswith("tools/"):
        if b.endswith(".py") or b.endswith(".sh"):
            hit(I, "I5", "工程工具实现")
            if re.match(r"^(test_|check_|verify_)|_verify\.|^watchdog\.", b):
                hit(J, "J2", "机械核验脚本（判定面）")
    if p.startswith(".github/scripts/") and b.endswith(".py"):
        hit(I, "I5", "镜像同步实现")
        if b.startswith("test_"): hit(J, "J2", "pytest 断言载体")
    if p.startswith("experiments/r"):
        if b.endswith(".py") or b.endswith(".sh"):
            hit(I, "I6", "实验实现")
            if re.match(r"^(test_|check_)|testlock_|_verify\.|^select_|^eval_|^bench_", b):
                hit(J, "J2", "校验/断言脚本")
        if p.startswith("experiments/r3_sft/r1_code/") and b in R0_TWIN:
            hit(M, "M9", "手写孪生镜像（与 r0_baseline 同名件，双份手改）")
    if p.startswith("survey/tools/"):
        hit(I, "I7", "报告构建工具")
    # ---- M 镜像/席位 ----
    if p.startswith("agent-festival/schema/examples/"):
        hit(M, "M1", "一致性夹具=schema 语义手写镜像（文件名即预期拒绝理由）")
        if "/envelopes/" in p: why[-1] = (why[-1][0], why[-1][1] + "；envelopes 面由 sign.py 机械再生")
    if "/testdata/" in p:
        hit(M, "M2", "扫描归档（再生型投影，--rerun 等值检验面）")
    if p == "agent-festival/go/internal/e2e/verifier_test.go":
        hit(M, "M3", "席位侧 band mirror 载体（verify=规则语义手写投影，B-1 病体）")
    if p in {".github/GITCODE-MIRROR.md", ".github/workflows/sync-gitcode.yml"}:
        hit(M, "M4", "GitCode 镜像席位维护面")
    if p in {"agent-festival/docs/attack-corpus.md", "agent-festival/docs/minor-ledger.md"}:
        hit(M, "M5", "台账归档（corpus:NN 被代码/回函交叉引用=状态投影面）")
    if p.startswith("agent-festival/examples/java-consumer/"):
        hit(M, "M6", "消费端席位参考实现（同契约第二独立实现）")
        if re.search(r"\.(java|sh|xml)$", b): hit(I, "I8", "席位侧代码实现")
    if (p.startswith("impl-2026-11-01/") and
        re.search(r"-(sent|formal|final|archived)\.md$", b)) or \
       p == "agent-festival/docs/2026-11-01-collision-response-archived.md":
        hit(M, "M7", "发送态文书副本（draft 正本/对外席位——63:157 文书形态 B-1）")
    if p.startswith("experiments/") and "/" not in p[12:] and \
       re.search(r"(\.json$|-RESULTS\.md$|REPAIR|REFERENCE-)", b):
        hit(M, "M8", "运行归档/结果快照（sha256+计数=run 派生物）")
    # ---- 保留模式（本仓 0 命中，声明跨仓可移植）----
    if p.startswith("agent-festival/review/"):
        hit(J, "J5", "review/ 探针（本仓 0 命中——hub 仓探针落点保留）")
    if re.match(r"^08-[A-Za-z0-9._-]*acceptance[A-Za-z0-9._-]*\.md$", b):
        hit(J, "J6", "验收表（跨仓保留——08 表在 inspect 仓，本仓 0 命中；前缀+词干+.md 全锚，91 号 acceptance 标题不误入）")
    if not cols:
        if p.startswith("survey/") or p == "survey/nga-survey.pdf": r, ph = "Z2", "analysis-only 报告面"
        elif p.startswith("wiki/"): r, ph = "Z3", "实验记录（证据文档，claim 面非双线辖域）"
        elif p.startswith("replies/"): r, ph = "Z1", "通信归档"
        elif p.startswith("impl-"): r, ph = "Z1", "通信/碰撞/回执工作档（非双线辖域）"
        elif re.search(r"(\.gitignore|\.gitkeep|go\.(mod|sum)|requirements\.txt|\.pdf$|\.tar\.gz$|README)", p):
            r, ph = "Z4", "配置/锁文件/索引/二进制（构建面）"
        elif p.startswith("experiments/"): r, ph = "Z5", "环境/服务器笔记/索引"
        else: r, ph = "Z0", "未归入四侧"
        why.append((r, ph + "——四列全·（扫描线不判）"))
    return cols, why

def row(path):
    cols, why = classify(path)
    m = {c: ("✓" if c in cols else "·") for c in "DIJM"}
    basis = "；".join(f"{r}·{ph}" for r, ph in why)
    return f"| `{path}` | {m['D']} | {m['I']} | {m['J']} | {m['M']} | {basis} |"

def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else "table"
    if mode == "table":
        for line in sys.stdin:
            line = line.rstrip("\n")
            if line: print(row(line))
    elif mode == "index":
        import ast, io
        tree = ast.parse(io.open(__file__, encoding="utf-8").read())
        fn = [n for n in tree.body if isinstance(n, ast.FunctionDef) and n.name == "classify"][0]
        rows = []
        def visit(stmts):
            for st in stmts:
                if isinstance(st, ast.If):
                    visit(st.body)
                    visit(st.orelse)
                elif isinstance(st, ast.Expr) and isinstance(st.value, ast.Call) \
                        and isinstance(st.value.func, ast.Name) and st.value.func.id == "hit":
                    a = st.value.args
                    col = getattr(a[0], "value", None) or getattr(a[0], "id", "?")
                    rows.append((st.lineno, a[1].value, col, a[2].value))
                elif isinstance(st, ast.Assign) and isinstance(st.targets[0], ast.Tuple) \
                        and [getattr(t, "id", "") for t in st.targets[0].elts] == ["r", "ph"]:
                    rows.append((st.lineno, st.value.elts[0].value, "·",
                                 st.value.elts[1].value + "——四列全·（扫描线不判）"))
        visit(fn.body)
        print("| 规则号 | 列 | 依据短语（§2.2 逐字导出） | §2.2 行号 |")
        print("|---|---|---|---|")
        for ln, rid, col, ph in rows:
            print(f"| {rid} | {col} | {ph} | :{ln} |")
        print(f"__RULE-ROWS {len(rows)}__", file=sys.stderr)
    elif mode == "scan":
        n = int(sys.argv[2])
        rng = subprocess.run(["git", "log", "--format=%h", "-n", str(n)],
                             capture_output=True, text=True).stdout.split()
        for pos, h in enumerate(rng, 1):
            files = subprocess.run(["git", "diff-tree", "--no-commit-id", "--name-only", "-r", h],
                                   capture_output=True, text=True).stdout.split("\n")
            files = [f for f in files if f]
            cs = set()
            for f in files: cs |= classify(f)[0]
            p3 = ("D" in cs) and ("I" in cs)
            de = ("J" in cs) and ("M" not in cs)
            subj = subprocess.run(["git", "log", "-1", "--format=%s", h],
                                  capture_output=True, text=True).stdout.rstrip("\n")[:72]
            side = "".join(c for c in "DIJM" if c in cs) or "-"
            print(f"#{pos:>3} {h} [{side:<4}] P3={'疑' if p3 else '—'} 失同步={'疑' if de else '—'} | {subj}")
            if (p3 or de) and len(files) > 0 and pos <= n:
                det = "; ".join(f"{f}({''.join(sorted(classify(f)[0])) or '·'})" for f in files)
                print(f"      文件侧: {det}")
if __name__ == "__main__":
    main()
```

---


### 2.3 机检自证（生成器重导出 == 本表，空 diff；嵌入代码 == 运行件，空 diff）

本位现跑（v0.1，对 HEAD `ec2ff2e` 树；命令为逐字执行原文，仓库根）：

```console
git ls-files | python3 -I /tmp/agentb/gen77.py table > /tmp/t.new
python3 -I /tmp/agentb/fix77/table_rows.py | wc -l
573
wc -l /tmp/t.new
573
grep -c '^| `' impl-2026-11-01/90-mixed-commit-classification-v0.md
577
diff /tmp/t.new <(python3 -I /tmp/agentb/fix77/table_rows.py) && echo TREE-MATCHES
TREE-MATCHES
python3 -I /tmp/agentb/fix77/extract_block.py > /tmp/agentb/fix77/gen77-from-doc.py
diff /tmp/agentb/gen77.py /tmp/agentb/fix77/gen77-from-doc.py && echo GEN-SYNC
GEN-SYNC
```

读数：区间内数据行数 == 重导出行数 == 573（TREE-MATCHES）；全文 `grep '^| '` 多出的行=区间外表形粘贴（§2.1 判别例
＋§4.7 自指行），两口径之差由上面成对命令逐字可复算。锚后 HEAD 若新增/删除文件，diff 将如实显示增删行——那是
触发器 1 的探针读数，不是表错。任何行不一致（表被手改/生成器被改而表未再生），diff 非空即触发器 2 命中——
**表不可手改，改分类=改 §2.2 规则后重导出**，此即 70 §2.2「分类表自身按术语锚定规则管理」的机检形态之一。
GEN-SYNC 探针守另一侧：§2.2 内嵌代码（唯一真源）经 `python3 -I /tmp/agentb/fix77/extract_block.py` 提取后与
运行件逐字节等——运行件不是真源，只是同步副本；两者漂移时此探针红。

---

### 2.4 逐类论证（分类按真实路径模式设计）

本节计数一律写裸数值（不带量词）——带量词即入本仓闸的命中面（量词口径），裸数值由下方复跑块全覆盖。
数真源=`/tmp/t.new`（§2.3 首命令 `git ls-files | python3 -I /tmp/agentb/gen77.py table > /tmp/t.new`）。
v0.1 相对 v0 的变化（时点记录）：D 列 87@v0.1（本件自指行入 D9），Z 族 105@v0.1（92 号入 Z1），D9 规则面 15（原 14），其余列计数与
逐规则计数不变——换类实质仅 Z1→D 一处（触发器 2 记录见 §3.2）。v0.4 注入轮 Z 族 105→106（94 号入 Z1，与本件自指同形）；一切现值以下方注入轮复跑块为准（块由 inject77.py 现跑生成，数字非手抄）。

```console
grep -c '^| [^|]* | ✓' /tmp/t.new
87
grep -c '^| [^|]* | . | ✓' /tmp/t.new
170
grep -c '^| [^|]* | . | . | ✓' /tmp/t.new
72
grep -c '^| [^|]* | . | . | . | ✓' /tmp/t.new
262
grep -c '^| [^|]* | · | · | · | ·' /tmp/t.new
106
grep -c '✓.*✓' /tmp/t.new
117
```

逐规则命中分布（uniq 转储，现跑；行=出现次数，键=规则号）：

```console
grep -oE '[DIJMZ][0-9]+·' /tmp/t.new | sort | uniq -c
   9 D1·
   3 D10·
   1 D11·
   2 D12·
  21 D2·
   6 D3·
   4 D4·
   8 D5·
  13 D6·
   1 D7·
   4 D8·
  15 D9·
  24 I1·
  36 I2·
  11 I3·
   2 I4·
  22 I5·
  65 I6·
   2 I7·
   8 I8·
  36 J1·
  26 J2·
   2 J3·
   8 J4·
 211 M1·
   1 M2·
   1 M3·
   2 M4·
   2 M5·
  10 M6·
  13 M7·
  10 M8·
  12 M9·
  67 Z1·
  11 Z2·
  15 Z3·
  10 Z4·
   3 Z5·
```

（转储只列树中真实出现的规则号——J5/J6/Z0 无行即 0 命中，与 §2.1 预留声明和 Z0 兜底现状一致。）

**D 类**。亚族：①**契约本体**（D1 schema 9、D2 判定规格夹具 21、D10 口径正文 3、D11 docstring 1）——
字段/类型增删即规格版本变更，直接喂 §3 触发器 3；②**预注册与冻结口径**（D3 proposal 6、D4 预注册/协议 4、
D5 参数载体 8、D9 冻结口径面 15）——CLAUDE.md §3 明文「experiments/ 下预注册文件对 metrics/splits/thresholds
规范有效」；D5 承认参数以代码形态存在的仓实（thresholds.py 等）：**同一文件既 D 又 I 是「定义藏进实现」的病形
暴露面**，非分类失误（v0.1 起 D5 短语改字面枚举，见 §2.2）。D9 本轮补 classification 面：正源分类表自身入定义集，
自指辖域见 §4.7。③**立法文本**（D6 架构规格/源快照 13、D8 守卫模式/runbook 4、D7 CLAUDE.md 1、D12 记录法 2）。
辖域论证：70 §2.2 的「定义变更应独立提交独立审」要求 D 集穷举一切定义面——D 集漏某类，该类偷渡即永不可检。

**I 类**。生产码（I1 go 非测试 24、I3 schema Python 面 11、I4 闸工具 2、I5 工程/镜像同步工具 22、
I7 survey 构建 2、I8 席位码 8）＋实验实现（I6 65，其中 test_/check_ 面同入 J）＋I2 测试码 36。
I2 的必要性是本仓形态的关键论证：P0-A 的被测实现（s1Judge/verify）按 50 §U 是「测试内确定性函数（非生产件）」——
若按通用仓习惯把 *_test.go 排除出实现类，c89219c 的实现侧将不可见，P3 线对 B-1 链亦失明。故 v0 对「判定与实现同体」
如实双标，并在 §4.5 承担由此而来的失同步粒度失配。

**J 类**。断言载体 J1（36）∪ 核验脚本 J2（26：tools/experiments/.github 的 test_*/check_*/verify_*/
watchdog）∪ 闸 J3（2：claim_gate＋corpus_manifest）∪ Python oracle J4（8——_suite.py 自述「oracles 保持检查器独立
=双校验器互检」，故每个 oracle 都是判定面）∪ 预留 J5/J6（本仓 0，收紧判别见 §2.1）。辖域论证：失同步线的
「判定侧」按 70 §2.2 原文指**出判定的机械面**而非「含断言的文件」——oracle 与闸判工件，测试判实现，三者同侧；
文书（回执/碰撞/裁决转述）不作判定侧：它们是判定结果的记录，记录漂移属文书形态 B-1（M7 列辖域，配对义务 v1 处理，
见 §五-3）。

**M 类**。镜像按同步机制分档（这正是 v1 配对登记表的输入）：①**再生投影**（M2 testdata 归档 1、
M8 运行归档 10）——机器可重生成，等值检验可机检（建造位 #2 --rerun 同族）；②**手写一致性夹具**（M1 examples 全 211：
文件名即预期拒绝理由；envelopes 面例外——由 sign.py 机械再生，其失同步病形只剩「派生旁路/生成器病」两种，
恰合 75 §四.1 残存面表述）；③**席位侧镜像**（M3 verifier_test.go band mirror 1、M6 java-consumer 席位 10、
M4 GitCode 镜像席位 2、M9 r1_code↔r0_baseline 手写孪生 12）；④**台账与文书席位**（M5 attack-corpus/minor-ledger 2
——corpus:NN 被代码回函双向引用；M7 发送态文书副本 13——63:157「同步此口径（对外转述与裁决真源的失同步——
B-1 病的文书形态）」即此列的存在理由）。

**Z 族与跨类**。四列全· 106（Z1 67、Z2 11、Z3 15、Z4 10、Z5 3；Z0 兜底现树 0）——注入轮 §2.4 复跑派生；前值 105 为 v0.1–v0.3 时点，差=94 号入 Z1；≥2 列跨类 117 与 v0
口径一致。v0 时本件自身行属 Z1，v0.1 起入 D9（§4.7）。

---

## 三、钉版规则（随规格钉版；与 70 号「定律随版本保鲜」同构）

**70 号 §五.5（定律保鲜）原文**：
> 「三层不可冒充律」若经确认实验立为定律，其适用版本随确认实验所用的生成器/清单版本**钉版**；
> 结构性变更（清单字段增删、生成器逻辑变更）触发重验（至少重跑最小病注入集），保鲜字段写入
> 定律条目——与 canary 规格联动纪律同构。

### 3.1 本表钉版字段（v0.1 待核准态）

| 字段 | 值（当前） |
|---|---|
| 表版本 | **v0.4-draft（外环 R2 修复位后；待夏核准，未钉版。版本↔commit：v0=`26d9856`·v0.1=`a69b8de`·v0.2=`32b77c0`·v0.3=`02374a0`·v0.4=`0e5d938`·v0.5=88 续批轮（本行随 commit 生长，等式以 §3.3 命令 1 现跑为准）** |
| 生成锚 | HEAD=`94fa9f1` 注入轮现跑（`git log -1 --format=%h`；`ec2ff2e`=v0.1 生成时点）；**本表钉在 HEAD=收卷后树**，修复位不改文件集，等式同值 |
| 规则集 | §2.2 生成器本体（规则面逐字导出见 §2.1；规则增删改=「生成器逻辑变更」——v0.1 已发生：D9 补 classification 面、J6 收紧全锚、D5 短语改字面枚举、新增 `index` 模式） |
| 规格锚（D 面版本） | P0-A 夹具 spec_version **1.1.0**（50 §U）；D1 的 schema 件各带版本字段；预注册件以各自「随 X 号冻结」条款为锚；口径来件=夏 R7（70 号） |
| 最小病注入集（重验样本） | **B-1 链**：c89219c 必须命中 P3=疑；b02761e 必须两线「—」（治法不误报）。回归命令=§3.3 命令 4——v0.1 已重验，读数未变（§3.2 触发器 3 记录） |
| 保鲜字段 | 任一触发器命中→重扫→表版本 +0.1；规则/辖域变更→+1.0 并随回执呈报；**夏核准日=钉版日**，钉版后「生成锚」改记核准提交。偏差如实记：v0.1 含规则变更，按本字段应 +1.0，未钉版期以 +0.1 记——钉版后版本自 v1.0 起严格适用本字段 |

### 3.2 触发器（三种，全部可机检）

1. **新增/删除被跟踪文件**：`git ls-files` 计数 ≠ 表行数即触发（§1.2 等式即探针；建造期间已被
   真实读数见下方「真实触发记录」块。。处置：重跑生成器，表增/删行，钉版字段「生成锚」更新。
2. **文件换类**：重导出表与在档表 diff，凡 ✓/· 翻转即换类（规则命中集变化的机械投影）。处置：
   换类行逐条说明依据变更，表版本 +0.1；换类若涉 D/M/J 三集边界（配对义务变化），同步 §2.4
   论证段与 §4.6-A 配对登记表。
3. **规格版本变更**：任一 D 类件结构性变更（schema 字段增删、spec_version bump、预注册口径
   修订——以 git 提交触碰 D 集为探针，探针=P3 线本身）。处置：重扫两线＋**重跑最小病注入集**
   （B-1 链回归必须仍命中），保鲜字段记「最近重验=提交 X」。

真实触发记录（滚动到本版）：
- 触发器 1 建造期读数：569→571（并发建造位落 91 号/93 号，v0 表 §1.2 原记——历史对照）。
- 触发器 1 本轮读数：571→573（`26d9856` 本件自身＋`d1c1943` 92 号）——v0.1 全表重导出（§1.2/§2.3 等式重建）即该读数的处置；本次修复 commit 不增删文件，修复轮 commit 后树等式同值。
- 触发器 2 本轮读数：规则集变更（D9 补 classification 面、J6 收紧全锚、D5 短语改字面枚举、`index` 模式新增）→全表与索引重导出；列 ✓/· 的实质换类仅此——本件自身行 Z1→D（M6 自指收口，§4.7），其余仅 basis 短语文字随 §2.2 变更。
- 触发器 1 v0.3 读数：573→574=94 号完成报告入树（收卷轮之后本人提交——自家机制抓住自家作者，
  等式由 §3.3 命令 1 现跑实证「生成有而表无: 94 号」后重导出归零）；行提取器口径未变。
- 触发器 1 v0.5 读数：575→578 净增三行（96 号 brief+两件裁剪 tar，tar 行走 Z 兜底纯派生）。事故留痕：c123655 的
  `git add -A` 把按 88§三.5 不该入库的仓外全量 tar 一并扫入（编排侧第 8 次工具越界），本笔已 `git rm --cached`
  脱账（磁盘副本未动）并以 .gitignore 防再犯，表体按脱账后文件集重导出。
- 触发器 2 v0.4 轮读数：无 ✓/· 翻转（94 号行本就在档且落 Z1，列形未动）；表体与 §3.3 命令 1 复跑 diff 空（TREE-MATCHES 574==574 注入轮现跑）。
- v0.4 处置面（外环 R2 M-3）：声明面派生标量残留——§1.4 题头/§0-1·§0-2/§1.2 等式/§2.4 列计数与转储/§0-3 90 窗读数/§3.1 锚面/页脚戳——全部改「现跑重派+脚本注入」或「带 commit 锚历史读数/指针形」；处置本身走机械路径（注入器=/tmp/agentb/review2/fixR-B/inject77.py 可重跑）；同轮修 §3.3 命令 4 分抄自足（m-R2-1）。m15 runbook-stale 族的声明面分支。
- 触发器 3 v0.2 读数：命令位稳化（E-1）+疑点清单随 HEAD 前进刷新（P3 集不变 10；失同步 6→9，
  新增各为闸修复独触笔——非规则变更、非表错，窗口语义所致；§4.3 v0.2 块）。
- 触发器 3 v0.1 读数：无——B-1 链最小病注入集（§3.3 命令 4）现跑重验，c89219c 行 P3=疑、b02761e 行两线「—」读数与 v0 一致。

同构映射（70 §五.5 → 本表）：清单字段增删↔D 集结构变更（触发器 3）；生成器逻辑变更↔§2.2
规则集变更（触发器 2 之因）；至少重跑最小病注入集↔B-1 链回归锚；保鲜字段写入定律条目↔§3.1
末行。差异如实记：定律以「确认实验所用版本」为锚，本表以「核准提交」为锚——分类表是被裁
体裁，钉版权在夏（§5），我方不自钉。

### 3.3 重扫命令（四条，块内顺序可跑，且**逐条分抄亦可跑**——每条自带前置：命令 4 内联 D= 计算；v0.4 修 m-R2-1：此前分抄命令 4 遇 D 未定义→窗=5→锚出窗→rc=1 零诊断，红得无解释；gen77.py 从 §2.2 代码块提取，或直接用同步运行件）

```console
# 1) 重导出整表并 diff 在档表（触发器 1/2 探针；提取域=TABLE-START/END 区间，v0.1 起收紧，见 §2.3 说明）
git ls-files | python3 -I /tmp/agentb/gen77.py table > /tmp/t.new
python3 -I /tmp/agentb/fix77/table_rows.py > /tmp/t.old    # 标记区间提取器（§2.3 注：BSD sed 区间跨区重开坑，故弃 sed）
diff /tmp/t.new /tmp/t.old; echo "exit=$?"
# 2) 双线试扫（窗口口径=§4.1；输出即 §4.2/4.3 疑点清单的再生形态）
D=$(git rev-list HEAD | grep -n -m1 '^c89219c' | cut -d: -f1)   # 锚深随 HEAD 自涨，非常数（时点复算：93@`fb53629`、95@`914c4e5`、96@`02374a0`、注入轮 97@`94fa9f1`——`git rev-list <rev> | grep -n -m1 '^c89219c'` cheap 复算）
python3 -I /tmp/agentb/gen77.py scan $((D+5)) > /tmp/scan77.out   # 位稳窗：永远覆盖锚
grep -c 'P3=疑' /tmp/scan77.out; grep -c '失同步=疑' /tmp/scan77.out
# 3) 规则索引重导出（§2.1 生成物）
python3 -I /tmp/agentb/gen77.py index
# 4) 最小病注入集回归（触发器 3 的重验义务；自带前置——D 未定义则当场内联计算，与命令 2 同式）
D=${D:-$(git rev-list HEAD | grep -n -m1 '^c89219c' | cut -d: -f1)}
python3 -I /tmp/agentb/gen77.py scan $((D+5)) | grep -E 'c89219c|b02761e' \
  | grep -q 'c89219c.*P3=疑' && echo "REGRESSION-OK"
#   期望：c89219c 行含 P3=疑（病引入命中）且 REGRESSION-OK 打出；b02761e 行 P3=— 失同步=—
#   （治法不误报）。断言并入退出位——锚出窗即红，杜绝 E-1 的「无声不达标」（90 固定窗实坠形态）。
```

---

## 四、双线试扫

### 4.1 口径声明（含扩窗）

任务规定窗口=最近 30 提交。v0.1 以修复轮锚（生成时 HEAD `ec2ff2e`）实跑：**窗口 30 疑点 P3=0、失同步=1**
——#1 `ec2ff2e`（claim_gate v2.0 出厂件）[IJ] 独触入失同步疑点，逐行在 §4.3 清单；历史锚 `c9b5965` 下同窗为 0/0
（v0 原记，对照保留）——窗口头新增条目正是 §4.2 注「并发落档可能使窗口头部新增」的兑现。这是校准性事实而非结果：
**30 窗口够不到 B-1 链**（c89219c=第 93、b02761e=第 76 @v0.2 锚；89/72@v0.1 锚，见 §4.4）。窗口弹性非 76 号条款（原「76 号云云」引语系误托出处，96 号批次已改直述，见 §0-3 留痕）——
为达「命中已知正例」的验收期望由我方自决扩窗：v0.1 扩至 90，v0.2 起改位稳窗 D+5（E-1）。
疑点判据按任务口径机检实现：P3=「同批 D 集非空且 I 集非空」；失同步=「同批 J 集非空且 M 集空」。
以下逐段粘贴（每段自带命令行；首段含窗口头 #1 疑点行与其文件侧明细，行号即 §4.2/§4.3 注记所用号）：

```console
git rev-list --count HEAD
377
python3 -I /tmp/agentb/gen77.py scan 30 > /tmp/scan30.out
grep -c 'P3=疑' /tmp/scan30.out
0
grep -c '失同步=疑' /tmp/scan30.out
1
python3 -I /tmp/agentb/gen77.py scan 30 | sed -n '1,8p'
#  1 ec2ff2e [IJ  ] P3=— 失同步=疑 | feat(tools): claim_gate v1.8→v2.0 两档+护栏（建造位 #2 出厂件，978 行）
      文件侧: agent-festival/tools/claim_gate.py(IJ)
#  2 d1c1943 [-   ] P3=— 失同步=— | docs(impl): 92 号五状态↔六区分↔我方判定词三源映射矩阵（建造位 #9，71 号 C8 收口）——三源快照经实时复取比对（gh5 
#  3 26d9856 [D   ] P3=— 失同步=— | docs(impl): 90 号——混装两线正源分类表 v0（建造位 #4：四列可多选 571 行全表+40 规则机检生成器嵌入+钉版规则三触发
#  4 c9b5965 [-   ] P3=— 失同步=— | docs(impl): 91 号反例↔验收框架归位映射表（建造位 #8）——左列 12 行（七探针逐行转录+五 RV 经存档转引标层级+四旧反例
#  5 fd9abf0 [-   ] P3=— 失同步=— | docs(impl): 93 号——三条协作者轻评论草稿（76 号任务 #6；V-C FAIL 轮重建）
#  6 da3958b [-   ] P3=— 失同步=— | docs(impl): 76 号沙箱化——乙辖区改为 ~/sandbox-agent-b（分支 agent-b-flank @ ad0177f）
#  7 ad0177f [-   ] P3=— 失同步=— | docs(impl): 76 号——智能体乙侧翼批交接（#2 闸升级两档/#4 分类表 v0/#8 反例映射/#9 词表对齐/#6 三评论草稿；
python3 -I /tmp/agentb/gen77.py scan 30 | sed -n '9,16p'
#  8 ef088b0 [M   ] P3=— 失同步=— | docs(impl): 75 号发送记录——用户手动发送确认（对夏 R7 回执）
#  9 d0cceb8 [M   ] P3=— 失同步=— | docs(impl): 75 号——对夏第七轮正式回执（两自错收讫+九问落地+预注册执行期三披露含自泄漏+两修正建议+两句计分裁语请求+领取收讫
# 10 1ba4b1a [-   ] P3=— 失同步=— | docs(impl): 74 号发送记录——MR!1 评论 199763878+issue#1 评论 199763908（GET 复核双在场）
# 11 94d3a8b [-   ] P3=— 失同步=— | docs(impl): 74 号——对 agent-hub-alpha 回复草稿（MR!1 五确认逐条+六问三分法入局+五分歧附证据+RV-05
# 12 962177c [-   ] P3=— 失同步=— | docs(impl): 73 号——四源碰撞独立复做报告归档（93 代理/1.71M tok/14 物种 gepa/12 叉 singulari
# 13 8a5118b [-   ] P3=— 失同步=— | docs(impl): 72 号——四源碰撞独立复做交接（任务原样+四源清单+输出要求+独立后对照步骤+已踩坑提醒）
# 14 50e33ef [-   ] P3=— 失同步=— | docs(impl): 71 号——四源深碰撞（夏R7×hub新issues×战场×frame-bench）：支撑7/改变5/待验证9/反驳澄清
# 15 b6024b8 [D   ] P3=— 失同步=— | docs(impl): 70 号——夏第七轮裁决归档（自错两处先认/混装判据拆两条/共变归属链三面/九问全裁复/邀约污染判据/夏方出臂+回避原则
python3 -I /tmp/agentb/gen77.py scan 30 | sed -n '17,24p'
# 16 e4ffbe4 [M   ] P3=— 失同步=— | docs(impl): 68 号正式版——草稿头去除+内部档号全部转译为对外自含引用（上封回执/第四轮回执/定稿回复等）——发送版定稿
# 17 f0f4b83 [-   ] P3=— 失同步=— | docs(impl): 68 号 B 选项随动——§2.1 补 MR!1 评审到达一行（方向接受/R-K 关闭/五反例在修，细节属邢线程）；终态
# 18 110d673 [-   ] P3=— 失同步=— | docs(impl): 69 号防雕花终审改 9 处——勾销/辖域/落点/断链/⑪/EDPA/改真注/60号行/指针界线全钉（每处 1-3 行落
# 19 07be21b [-   ] P3=— 失同步=— | docs(impl): 69 号连锁改真——判据同步 1 连（覆盖 3 连口径；沿革四段；锚 50cf9ce）
# 20 50cf9ce [D   ] P3=— 失同步=— | docs(runbook): 停机判据降为外层 1 次 0B/0M（用户裁定 2026-11-10）——判据沿革四段注记；更强保证改手动启动；叠
# 21 78e55c4 [-   ] P3=— 失同步=— | docs(impl): 69 号深检改真——33 条复算+漏 8 条补齐+幻影㉟删+三裁决入档（五候选全立/hash引用/Unicode不立）+
# 22 38cb2d4 [-   ] P3=— 失同步=— | docs(impl): 69 号——skill 立法回溃交接文件（自包含：背景/回溃清单 A-E/纪律/验收/坑）
# 23 b483128 [D   ] P3=— 失同步=— | docs: 断链改真——runbook/54 号 skill 引用路径指向软链真身（yaojun-skills 软链组已建）
python3 -I /tmp/agentb/gen77.py scan 30 | sed -n '25,31p'
# 24 ad3a5fa [M   ] P3=— 失同步=— | docs(impl): 65 号——MR !1 第二批追加记录（bb85222=67 号预注册入 MR）
# 25 f4ece53 [-   ] P3=— 失同步=— | docs(impl): 68 号 §四.6 备胎残留随动（m-3）——与 §2.2 必跑口径一致；R3 退出（0B/0M/3m，m-1/m-2 
# 26 797fadb [D   ] P3=— 失同步=— | docs(impl): 67/68 R2 清偿——68 备胎语义改必跑；67 附 18-20（裁切四类按类补全/检测通道含 GitHub PR#
# 27 63e2dfc [D   ] P3=— 失同步=— | docs(impl): 67 号附 13-17——R1 清偿（译者送达件封闭枚举解附8/9矛盾/时序锚升级推送事件序/弱形态臂必跑/首译绑定可证
# 28 f7adb04 [D   ] P3=— 失同步=— | docs(impl): 67/68 终审清偿——弱形态臂执行纪律钉死+源节披露补全+方向错误改真
# 29 e989e34 [D   ] P3=— 失同步=— | docs(impl): 67/68 号深审清偿——隔离实况披露+冻结史改真+归因代码级改真+引文逐字
# 30 c95cdda [-   ] P3=— 失同步=— | docs(impl): 68 号闸清偿——请求清单计数附复核命令（42/21 行为引语形豁免）
```


### 4.2 P3 线（定义/规格＋实现/修复 同批 = 疑）——疑点 10（v0.1 锚现跑留档；@v0.2 位稳窗 98 集不变，行号 +4；@v0.4 位稳窗 102 集不变、行号随窗头再进（c89219c #89→93→97）——行号一律以 §3.3 命令 2 输出为准；命令与输出同段=M5 收口）

```console
python3 -I /tmp/agentb/gen77.py scan 90 > /tmp/scan90.out
grep -c 'P3=疑' /tmp/scan90.out
10
grep -A1 'P3=疑' /tmp/scan90.out | sed -n '1,8p'
# 40 fe9749a [DIJM] P3=疑 失同步=— | fix(tools): R77 审查清偿——claim_gate v1.8+内环六迭代收敛（0B/0M）
      文件侧: agent-festival/docs/minor-ledger.md(M); agent-festival/docs/review-loop-runbook.md(D); agent-festival/tools/claim_gate.py(IJ)
# 41 eb5b21f [DIJM] P3=疑 失同步=— | fix(tools): R76 审查清偿——claim_gate v1.6+内环九迭代收敛（0B/0M）
      文件侧: agent-festival/docs/minor-ledger.md(M); agent-festival/docs/review-loop-runbook.md(D); agent-festival/tools/claim_gate.py(IJ)
--
# 43 9321740 [DIJ ] P3=疑 失同步=疑 | fix(tools): claim_gate v1.2——singularity 攻击续清+边界自测规格
      文件侧: agent-festival/docs/review-loop-runbook.md(D); agent-festival/tools/claim_gate.py(IJ)
# 44 8331b9d [DIJ ] P3=疑 失同步=疑 | feat(runbook): §五 改写纪律六条强执行+claim_gate 机械闸（用户口令立法）
grep -A1 'P3=疑' /tmp/scan90.out | sed -n '9,16p'
      文件侧: agent-festival/docs/review-loop-runbook.md(D); agent-festival/tools/claim_gate.py(IJ); impl-2026-11-01/64-agent-hub-alpha-survey.md(·)
# 45 642974e [DIJM] P3=疑 失同步=— | fix(e2e): R75 审查清偿——判别钉三层递进+全树枚举改真+内环五迭代收敛（0B/0M）
      文件侧: agent-festival/docs/minor-ledger.md(M); agent-festival/docs/review-loop-runbook.md(D); agent-festival/go/internal/e2e/t2host_test.go(IJ); agent-festival/go/internal/e2e/x3_entry_test.go(IJ); impl-2026-11-01/64-agent-hub-alpha-survey.md(·)
--
# 51 f671972 [DIJM] P3=疑 失同步=— | fix(e2e): R74 审查清偿——C 支路全流程+常驻 tripwire+四轮内环收敛（0B/0M）
      文件侧: agent-festival/docs/attack-corpus.md(M); agent-festival/docs/minor-ledger.md(M); agent-festival/docs/review-loop-runbook.md(D); agent-festival/go/internal/e2e/t2host_rk_test.go(IJ); agent-festival/go/internal/e2e/t2host_test.go(IJ); agent-festival/go/internal/e2e/x3_entry_test.go(IJ); impl-2026-11-01/60-n3-scope-prior.md(D); impl-2026-11-01/61-correction-xingchao-numbers.md(·)
--
# 53 e6b7bab [DIJ ] P3=疑 失同步=疑 | feat(e2e): N3 统一入口+端口化——三分流+X3 八行+R/K 改走入口+failingProbe 退役
grep -A1 'P3=疑' /tmp/scan90.out | sed -n '17,24p'
      文件侧: agent-festival/go/internal/e2e/t2host_rk_test.go(IJ); agent-festival/go/internal/e2e/t2host_test.go(IJ); agent-festival/go/internal/e2e/x3_entry_test.go(IJ); impl-2026-11-01/60-n3-scope-prior.md(D)
--
# 81 799315c [DIJM] P3=疑 失同步=— | fix(corpus+e2e): R68 清偿——改写行结论保真+R37 宣称句改真（0B/2M/7m；外环连胜 0）
      文件侧: agent-festival/docs/attack-corpus.md(M); agent-festival/docs/minor-ledger.md(M); agent-festival/docs/review-loop-runbook.md(D); agent-festival/go/internal/e2e/verifier_test.go(IJM)
--
# 85 8588cc9 [DIJM] P3=疑 失同步=— | fix(e2e)+docs: R67 清偿——退役同步同族穷举+探针消息契约（0B/4M/4m；外环连胜 0）
      文件侧: agent-festival/docs/attack-corpus.md(M); agent-festival/docs/minor-ledger.md(M); agent-festival/docs/review-loop-runbook.md(D); agent-festival/go/internal/e2e/verifier_test.go(IJM); impl-2026-11-01/50-p0a-scope-prior.md(D)
--
grep -A1 'P3=疑' /tmp/scan90.out | sed -n '25,26p'
# 89 c89219c [DIJM] P3=疑 失同步=— | feat(e2e)+schema: P0-A 逾期语义恢复——事件时间代数+六格判定真源（期望先行冻结）
      文件侧: agent-festival/go/internal/e2e/correctness_probe_test.go(IJ); agent-festival/go/internal/e2e/e2e_test.go(IJ); agent-festival/go/internal/e2e/p0a_window_test.go(IJ); agent-festival/go/internal/e2e/testdata/p0a_scan.tsv(M); agent-festival/go/internal/e2e/verifier_test.go(IJM); agent-festival/schema/examples/judgment/valid/judgment-spec-credit.json(DM); impl-2026-11-01/50-p0a-scope-prior.md(D)
```

逐条注记（疑点≠定罪，处置=转人工判读，语义同 70 §2.2；行号=v0.1 锚，括号内为 v0 历史锚原号——窗口头新增
使旧提交整体后移；**疑点清单以上方粘贴输出为准**，本注按内容对应写）：
- **#89 c89219c（v0 号 #86）＝已知正例命中 ✓**（§4.4 详）。
- #40 fe9749a（#37）/#41 eb5b21f（#38）/#45 642974e（#42）/#51 f671972（#48）/#81 799315c（#78）/#85 8588cc9（#82）
  （闸/清偿批次）：runbook（D8 立法文本）或 50/60 号口径面（D9）与闸/测试实现同批——立法条款在 fix 批次内修订=
  「定义藏进修复提交」的文书面；疑点成立、严重度低（立法即本批主题，属「定义变更应独立提交独立审」的形式缺口，
  非实质偷渡；#85 8588cc9 另见 §4.4 链中间件行）。
- #43 9321740（#40）/#44 8331b9d（#41）（claim_gate v1.2 与口令立法批）：runbook D＋闸 IJ 同批、无 M——真疑点：
  闸规则与闸辖域条款耦合落地，闸变而辖域未独立审即为 P3 病形候选；两线同触，§4.3 并看。
- #53 e6b7bab（#50，N3 统一入口）：60 号 scope-prior（D9 冻结口径）＋*_test.go（IJ）同批、无 M——疑点真实
  （口径随 feat 落地未独立审）。


**v0.2 锚位稳窗复跑（fb53629，D=93→窗 98）**：`P3疑=10 / 失同步疑=9`（scan98 粘贴，锚行与三新增行见 §4.3 v0.2 块；
P3 疑点集合与 v0.1 逐 commit 相同，行号整体后移系窗口头新增之积——非新疑点）。

**v0.4 锚位稳窗复跑（`94fa9f1`，D=97→窗 102）**：`P3疑=10 / 失同步疑=9`（注入轮现跑，命令=§3.3 命令 2；疑点集合与 v0.1/v0.2 逐 commit 相同）。

### 4.3 失同步线（判定侧独触，同批无镜像/席位 = 疑）——疑点 6@v0.1→**9@v0.2**（v0 记 5；现跑=活树，M5 收口）

```console
grep -c '失同步=疑' /tmp/scan90.out
6
grep -A1 '失同步=疑' /tmp/scan90.out | sed -n '1,8p'
#  1 ec2ff2e [IJ  ] P3=— 失同步=疑 | feat(tools): claim_gate v1.8→v2.0 两档+护栏（建造位 #2 出厂件，978 行）
      文件侧: agent-festival/tools/claim_gate.py(IJ)
--
# 43 9321740 [DIJ ] P3=疑 失同步=疑 | fix(tools): claim_gate v1.2——singularity 攻击续清+边界自测规格
      文件侧: agent-festival/docs/review-loop-runbook.md(D); agent-festival/tools/claim_gate.py(IJ)
# 44 8331b9d [DIJ ] P3=疑 失同步=疑 | feat(runbook): §五 改写纪律六条强执行+claim_gate 机械闸（用户口令立法）
      文件侧: agent-festival/docs/review-loop-runbook.md(D); agent-festival/tools/claim_gate.py(IJ); impl-2026-11-01/64-agent-hub-alpha-survey.md(·)
--
grep -A1 '失同步=疑' /tmp/scan90.out | sed -n '9,16p'
# 53 e6b7bab [DIJ ] P3=疑 失同步=疑 | feat(e2e): N3 统一入口+端口化——三分流+X3 八行+R/K 改走入口+failingProbe 退役
      文件侧: agent-festival/go/internal/e2e/t2host_rk_test.go(IJ); agent-festival/go/internal/e2e/t2host_test.go(IJ); agent-festival/go/internal/e2e/x3_entry_test.go(IJ); impl-2026-11-01/60-n3-scope-prior.md(D)
--
# 60 a9e8c92 [IJ  ] P3=— 失同步=疑 | test(e2e): X-S2 空白子检验清偿——Events() 替身的条件性限定（retry 政权分叉）+NOTE 改真+残留件删
      文件侧: agent-festival/go/internal/e2e/zz_xs2_demo_test.go(IJ)
--
# 62 c251566 [IJ  ] P3=— 失同步=疑 | test(e2e): X-S2 两臂演示——参数形态信任洞实锤+机器内部形态免疫（overlay 风格零生产码）
      文件侧: agent-festival/go/internal/e2e/zz_xs2_demo_test.go(IJ)
```

逐条注记（行号=v0.1 锚，括号内 v0 原号）：#1 `ec2ff2e` 为**窗口头新增疑点**（闸 v2.0 出厂件独触判定面，
attack-corpus/minor-ledger 台账零同批——outer-r1 N1 预告的「现树+1」读数即此；与闸演进两批同族，转人工判读）。
#43 9321740（#40）/#44 8331b9d（#41）闸演进无台账同批——**候选真阳性**（闸与反例台账是本仓明示配对，corpus:NN 被
双向引用；闸规则增删而台账零动作=台账漂移形态）；#53 e6b7bab（#50）同型（N3 判定面重构、无归档/台账动作——待人工
判读 N3 是否触及 p0a_scan.tsv 语义）；#60 a9e8c92（#57）/#62 c251566（#59）为**已知假阳性族**：zz_xs2_demo_test.go
是 overlay 风格零生产码一次性演示件，本无镜像义务——「判定侧独触」对**新增无镜像义务件**过报，v1 修向=配对登记表＋
白名单（§4.6-C）。


**v0.2 新增（scan98 粘贴，闸修复独触笔逐行）**：

```console
#  1 fb53629 [IJ  ] P3=— 失同步=疑 | test(tools): RR30/RR31 分量钉补齐（内环复审 m-A 守恒律缺口）
      文件侧: agent-festival/tools/claim_gate.py(IJ)
#  2 d26a810 [IJ  ] P3=— 失同步=疑 | fix(tools): claim_gate v2.1→v2.2 F-1 围栏补洞（内环审计轮实锤）
      文件侧: agent-festival/tools/claim_gate.py(IJ)
#  4 5f289ac [IJ  ] P3=— 失同步=疑 | fix(tools): claim_gate v2.0→v2.1 修复轮 R1（外环 7M 之代码面 M1/M2/M3/M4/N4/N5）
      文件侧: agent-festival/tools/claim_gate.py(IJ)
```

注：三行为本战役闸修复 commit，claim_gate 独触（IJ）无镜像——失同步线判定为真（runbook/minor-ledger 镜像侧确未同批，
m15 stale 债的机检显形）。处置=登记转人工（runbook 滚动属主线辖区，非本战役可修）；与 §4.2 #43/#44 闸批次行同族并看。

### 4.4 已知正例 B-1 链命中验证（50 号 §七）

链锚（50 §七 U 原文）：「判定侧/席位侧/s1Judge/verify 实现版本＝本档提交
（`git log -1 --format=%h`，**冻结时期 c89219c，B-1 修复后 b02761e**）」。位置（v0.1 锚现跑；v0 在历史锚 `c9b5965`
下记 #86/#69，差=窗口头新增的 `26d9856`/`d1c1943`/`ec2ff2e` 三条，历史对照见 §4.1）：

```console
git log --oneline | cat -n | grep -E 'c89219c|b02761e'
     3	26d9856 docs(impl): 90 号——混装两线正源分类表 v0（建造位 #4：四列可多选 571 行全表+40 规则机检生成器嵌入+钉版规则三触发器+B-1 链双线试扫；P3 线命中病引入 c89219c，失同步线对同批形式同步病形失灵如实报；待夏核准钉版）
    72	b02761e fix(e2e): R72 B-1 组合面死锁+扫描决定性度量（第四轮碰撞亲验的两条）
    89	c89219c feat(e2e)+schema: P0-A 逾期语义恢复——事件时间代数+六格判定真源（期望先行冻结）
git show --name-only --format='%h %ad %s' --date=short c89219c
c89219c 2026-11-07 feat(e2e)+schema: P0-A 逾期语义恢复——事件时间代数+六格判定真源（期望先行冻结）

agent-festival/go/internal/e2e/correctness_probe_test.go
agent-festival/go/internal/e2e/e2e_test.go
agent-festival/go/internal/e2e/p0a_window_test.go
agent-festival/go/internal/e2e/testdata/p0a_scan.tsv
agent-festival/go/internal/e2e/verifier_test.go
agent-festival/schema/examples/judgment/valid/judgment-spec-credit.json
impl-2026-11-01/50-p0a-scope-prior.md
git show --name-only --format='%h %ad %s' --date=short b02761e
b02761e 2026-11-08 fix(e2e): R72 B-1 组合面死锁+扫描决定性度量（第四轮碰撞亲验的两条）

agent-festival/go/internal/e2e/p0a_window_test.go
agent-festival/go/internal/e2e/testdata/p0a_scan.tsv
agent-festival/go/internal/e2e/verifier_test.go
```

逐文件分类（§1.4 表行）：c89219c 触达件 = D∩M（judgment-spec-credit.json）＋D（50-p0a-scope-prior）
＋I∩J（*_test.go 测试件）＋I∩J∩M（verifier_test.go）＋M（p0a_scan.tsv）。两线判定：

| 链件 | P3 线 | 失同步线 | 判读 |
|---|---|---|---|
| **c89219c 病引入** | **疑 ✓ 命中**（D 集：判定规格夹具＋冻结口径 与 I 集测试实现同批——题面字面「feat(e2e)+schema」偷渡） | **—（漏，§4.5）** | P3 线命中成立 |
| **b02761e 修复** | —（无 D 集） | —（J 与 M 同批=原子同步治法，不误报） | 治法两线放行 ✓ |

（链中间件 #81 799315c（v0 #78）/#80 aad1906（v0 #77）均因同批触碰台账/镜像而两线静默——§4.5-3。）


### 4.5 如实报告：**v0 失同步线对本仓病体形态失灵**

失同步线谓词（任务原文口径）＝「判定侧独触为疑」。对 B-1 病引入提交 c89219c 实跑报「—」——漏检。这不是窗口或数据问题（谓词与分类都已就位），是**粒度失配**，原因如下（v0 的原因清单与「非恒真实证」段在 v0.1 无实质变化，仅注号随锚改）：

1. **同批形式同步 ≠ 内容共变**。B-1 的病形是「verifier_test.go 的 band mirror 前提条件化缺失」
   （53 号 §三：mirror 无条件认为资产>30 ⇒ rule 1 覆盖，P0-A 把 coverage 改为「且窗口干净」后
   蕴含失效）——而该文件在病引入同批**被触碰过**（归档 tsv 亦同批再生），「镜像对应文件未同批
   变更」谓词在文件粒度为假。70 §2.2 谓词能抓「整文件漏跟」的漂移，抓不住「改了镜像文件但
   镜像逻辑未随判定语义更新」的漂移——本仓 B-1 恰属后者。

2. **病体两面跨同一文件类**。s1Judge 与 band mirror 都是 *_test.go 内的确定性函数（50 §U），
   I2/J1/M3 三标同体：分类表的「判定侧/镜像侧」在文件粒度部分重叠，「J 集非空且 M 集空」
   谓词对两侧同体的文件天然失灵。

3. **台账镜像稀释**。M5（attack-corpus/minor-ledger）在 R67/R68 各清偿批恒被触碰，含测试改动
   的批次几乎必然「同批有 M」——失同步线在本仓近窗口内仅对「零台账动作」的极端独触敏感
   （#43 9321740、#44 8331b9d、#53 e6b7bab、#60 a9e8c92、#62 c251566——v0 号 #40/#41/#50/#57/#59，外加本轮窗口头 #1 ec2ff2e 同族新增，清单见 §4.3 粘贴），敏感度被台账面结构性压低。

**非恒真实证（mutation-RED 精神）**：谓词换成内容级则 B-1 当场可判——b02761e 自身即证明：
其修复形态（COMPOSED 面断言 `verify(s1Judge(e)).ok == true` 逐格永久钉死，53 号 §三）是内容级
共变机检，病态组合下必红；文件级 v0 谓词跑不出这个红。v0 的失同步线对本仓的价值因此限于
「台账漂移探针＋独触探针」两族，不覆盖 B-1 病体本体。

---

### 4.6 汇总命中矩阵 + v1 方向

| 目标 | P3 线 | 失同步线 |
|---|---|---|
| 病引入 c89219c | **命中** | 失灵（§4.5，如实） |
| 修复 b02761e | 正确放行 | 正确放行（治法不误报） |
| 窗口 30 全量（v0.1 锚） | 0 疑（窗口以文书面为主，校准性事实） | 1 疑（#1 闸出厂独触——窗口头新增条目；历史锚下为 0，见 §4.3 注） |

v1 方向 A/B/C（供夏核准时一并裁）：**A 配对登记表**——判定侧↔对应镜像面显式枚举
（p0a_window/e2e↔verifier band mirror↔p0a_scan.tsv；schema/*.schema.json↔examples 对应族；
claim_gate↔attack-corpus；runbook §五↔claim_gate 规则族；r0_baseline↔r1_code 孪生；
draft↔sent 文书对），失同步线谓词由「任意镜像未动」升级为「对应镜像未动」，同时以
无镜像义务白名单消 §4.3 演示件假阳性；**B 归档再生等值**（--rerun 同族：tsv/manifest 与重算
输出 diff=内容级共变证明，覆盖 §4.5-1 病形）；**C 组合断言推广**（COMPOSED 面 `verify(s1Judge(e))`
从 P0-A 八格推广到六格判定词表全对——即 70 §7.2「实现间组合」亚型的本仓落点）。
寿命注记：A/B/C 落地即进入 70 §2.2「生成制落地后此线自然消亡」轨道，残存面（派生旁路/
生成器病）按 75 §四.1 留常设探针——本 v0 双线试扫正是该过渡态的基线测量。

---

### 4.7 自指判定：本表对自身提交行的双线判定重跑（M6 修复位）

v0 曾自我预言「下一次重扫将为 77 自身建行（归 D9 系）」（v0 §1.2 尾注与 §五-7），而实跑归 Z1、双线皆不判——
outer-r1 M6 判为自指失灵。v0.1 的 D9 正则补 classification 面后，对「表自身」重跑判定，现跑粘贴：

```console
git ls-files | python3 -I /tmp/agentb/gen77.py table | grep 77-mixed
| `impl-2026-11-01/90-mixed-commit-classification-v0.md` | ✓ | · | · | · | D9·冻结口径面（scope-prior/预注册/战场冻结/裁决来件/正源分类表自指面） |
python3 -I /tmp/agentb/gen77.py scan 90 | grep 26d9856
#  3 26d9856 [D   ] P3=— 失同步=— | docs(impl): 90 号——混装两线正源分类表 v0（建造位 #4：四列可多选 571 行全表+40 规则机检生成器嵌入+钉版规则三触发
```

判读如实：#3 `26d9856`（本件自身引入提交）是纯文书批——P3 需同批 I 面、失同步需 J 独触，皆不成立，「—」为
正确非误报；修复前该文件不在任何集内，「表体改动＋实现改动同批」发生时两线同样失明，现在它是判据的直接输入。
辖域生效用合成反例实证（非历史提交，谓词直演——与 §4.1 同一 `classify()`）：

```console
python3 -I -c "import sys;sys.path.insert(0,'/tmp/agentb');import gen77 as G;f=['impl-2026-11-01/90-mixed-commit-classification-v0.md','agent-festival/go/internal/e2e/p0a_window_test.go'];cs=set();[cs.update(G.classify(x)[0]) for x in f];p3='D' in cs and 'I' in cs;de='J' in cs and 'M' not in cs;print(sorted(cs),'P3=','疑' if p3 else '—','失同步=','疑' if de else '—')"
['D', 'I', 'J'] P3= 疑 失同步= 疑
```

即：改表混入实现改动→P3=疑 ✓；改表混入判定面改动且无镜像同批→失同步=疑 ✓。表体与内嵌生成器自身的失同步
（规则改了、表没再生）在文件粒度下不可表达（§4.5-2 同型失配），由 §2.3 的 TREE-MATCHES/GEN-SYNC 双探针（触发器 2）
守——三层合观，70 §2.2「分类表自身按术语锚定规则管理」的意图至此有机械落点。v0 §五-7 的「不在表内」预言已改真。

---

## 五、边界声明（v0.1 辖域声明集中于此）

1. **本件是 v0.1 草案**（修复轮 R1 后）：分类=我方提议，**夏核准后钉版**（语义同 v0）；核准前一切钉版字段按未钉态运行，
   双线试扫结果不构成正式轮的定谳输入。
2. 全表辖域=被跟踪文件（生成锚 `ec2ff2e` 树；钉版锚=修复轮 commit 后树，见 §1.2 锚口径声明）；合并提交在窗口 90 内不存在（题面线性）；rename
   未启用 `-M`（diff-tree 取新路径），主线如出现 rename 流需重扫。
3. 四列全· 的 Z 族（计数 106=注入轮 §2.4 复跑派生值）是「扫描线永不判」声明，不是「无角色」声明——文书失同步
   （对外转述 vs 真源，63:157）归 M7 列，但 draft↔sent 逐对配对义务 v0.1 不机检（A 族 v1 辖域）。
4. 「判定侧独触」按任务口径实现为「J 集非空且 M 集为空」，「对应文件」降级为「任意镜像面」——
   该近似的失配面已实测为 §4.5，未粉饰。
5. 生成器/表格/规则索引三者关系：§2.2 生成器为唯一可执行真源；§1.4 表与 §2.1 索引皆为逐字导出物
   （§2.3 TREE-MATCHES 与 §2.1 导出命令实证——v0 的索引系手抄，m8 后改机检）；任何不一致以重跑生成器为准修复，**表不可手改**。
6. 回收方式：本件在沙箱分支 agent-b-flank，不自行合并、不推送；由主线 cherry-pick（76 号协议）。
7. 本件 77 自身：v0 记「不在表内（生成时未跟踪）——提交即触发器 1，下版重扫补行（D9 系）」——v0.1 兑现：
   D9 正则补 classification 面后自身行落 D 集，对自身提交行的双线判定与辖域实证见 §4.7（M6 收口）。

---

## 六、闸自扫登记（v0.1 发送前义务，R76-③；命中二分=93 号同款口径）

闸体：修复位 A 的在作 `agent-festival/tools/claim_gate.py`——粘贴时点为 v2.1（`--selftest` rc=0 本位现跑）；
A 后续如再改口径，本节命令重跑并随轮更新登记。比较锚版本匹配：v0 文件（`git show 26d9856:` 取回）在两版闸下同值。

```console
git show 26d9856:impl-2026-11-01/90-mixed-commit-classification-v0.md > /tmp/77-v0.md
git show HEAD:agent-festival/tools/claim_gate.py > /tmp/gate-v20.py
python3 -I /tmp/gate-v20.py /tmp/77-v0.md | head -1
claim gate: 49 counting claim(s) lack nearby command evidence:
python3 -I agent-festival/tools/claim_gate.py /tmp/77-v0.md | head -1
claim gate: 49 counting claim(s) lack nearby command evidence:
python3 -I agent-festival/tools/claim_gate.py impl-2026-11-01/90-mixed-commit-classification-v0.md
claim gate: 2 counting claim(s) lack nearby command evidence:
impl-2026-11-01/90-mixed-commit-classification-v0.md:53: count claim '一提交' without a command in the ±8-line window: > | **P3 偷渡族**（定义变更藏进修复提交） | 同一提交**双侧同触**（定义/规格类文件＋实现/修复类文件） | 「同一提交同时触碰两侧」→ 疑点，定义变更应独立提交独
impl-2026-11-01/90-mixed-commit-classification-v0.md:56: count claim '两条' without a command in the ±8-line window: > 「同时触碰两侧」对病族 B 是**治法形态**（合法的原子同步提交）不是病形……两条扫描线的文件分类
```

判定：**49 → 2，只减不增 ✓**（同闸 v2.1 口径；v2.0 下 v0 亦 49——锚版本匹配可比）。

基线命中逐族二分（左列=v0 原行号，v0 文件面）：

| v0 原命中行号 | 族 | v0.1 处置 |
|---|---|---|
| :31/:1080/:1100/:1112/:1127/:1145/:1154 | 散文量词断言 | TP——重写裸数值/去量词（§〇、§4.4-§4.6、§五-3、页脚） |
| :42/:45 | 70 §2.2 逐字引文内量词 | FP——逐字引文豁免（引文不可改写，证据=70 号正文本身；v0.1 随 §1.1 下移为现跑 :53/:56，见下） |
| :53/:84/:885 | 题头/段内量词 | TP——§1.4 题头改 inline 计数命令；§2.3 措辞去量词 |
| :448-:476/:674/:754 | D5 旧短语量词（表 8＋索引 1＋源码 1） | TP——§2.2 的 D5 短语改字面枚举，表/索引/源码三面清零（触发器 2，§3.2） |
| :890-:919/:985/:988 | 逐类计数无命令 | TP——§2.4 复跑块（列计数六对＋uniq 转储）＋§4.1 现跑粘贴 |
| :994-:1023 | 窗口题面引语行 | TP——§4.1 分段粘贴，每段自带命令行（±8 窗全覆盖） |
| :1026/:1041/:1044 | 疑点计数占位区 | TP——M5：§4.2/§4.3 命令+输出同段，占位符花括号形态清零 |
| :1087/:1088 | 链件量词断言 | TP——§4.4 改裸列举（触达件=D∩M＋D＋…） |

残余命中（现跑 2 个，全部 FP-逐字引文族）：:53 的「…独立提交独立审」与 :56 的「…两条扫描线的文件分类」——
均在 §1.1 的 70 §2.2 逐字引用块内；豁免理由：该面为引用面，改写=制造新的引用失真（本轮反而按原句补回漏字，
见 §1.1 勘补注）；同族先例=68 号「42/21 行为引语形豁免」。除此之外本件无未处置命中。

*v0.4-draft（外环 R2 修复轮，断网半程由编排者收官）· 版本↔commit 与全部现值读数一律看 §3.1 映射行与 §3.3 四条命令现跑——本行不落裸数（v0.4 指针化决议：等式=公式，页脚只是戳）· 前代戳文归档于 git 本文件历史**

