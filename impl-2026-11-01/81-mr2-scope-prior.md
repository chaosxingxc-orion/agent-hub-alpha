# 81 · MR !2 范围框定：入口契约修复批（五 RV 关闭——期望先行）

> 范式：xuefanfan 分层律（语义层=入口路由契约，执行资源=注入端口，不新增装配）+ SETools 书写纪律。
> 判定真源冻结：§三 X3-12..16 期望行冻结于任何实现之前——翻译自 MR !1 评审 199470159 §二+GH#4 契约。
> 分支：mr2-entry-contract（自 n3-unified-entry@534ac03）。

## 一、IN

1. **X3-12..16 五行期望冻结**（§三）+ 红探针先行（每行先红后修）。
2. **入口最小修**（nodeEntryExec，语义层）：
   - RV-01：B/C 路径增派发级授权检查（调用者须持 execute 级授权；read 不足即拒、零副作用）；
   - RV-02：storedResult 增 CmdDigest 字段（C/B/adoptAndDispatch 落存时写入）；A 支路比对 first.req.PayloadDigest≠stored.CmdDigest ⇒ Conflict；nil-first 重试 ⇒ 拒（命令材料缺失——历史读取是另一操作，评审 §二 RV-02 How）；
   - RV-03：B 支路走完整生命周期（method.execute→Adopt→Dispatch→store→**completeNode**）——K 产自己的正式结果、完成后重试走 A、方法失效后仍走 A（历史不被静默重执）；
   - RV-04：C 支路重排 **Adopt→Invoke→Dispatch**（Adopt=执行所有权 claim，机器单胜语义即所有权）；Adopt 返回 "reused"（同键同摘要并发第二名）⇒ 视为已被在途执行占用而拒；
   - RV-05：入口预检 outer opID ≠ first.req.Identity.OperationID ⇒ 拒，且 Δ执行调用=0 ∧ Δ采纳状态=0（预检在任何状态写入之前）。
3. **对方六探针原样入常驻套件**（mr1_review_test.go，逐字 vendor 自复现附件 199470012——不改断言；修复后须全绿）。
4. **既有测试随契约升级**（改调用点传 first/删手工补步骤——每处改动在 commit message 解释契约变化，评审 §六.4 要求）。
5. m-8 顺带修：60 号 hub 副本 X3-11b 行依据错列+行数漂移。

## 二、OUT

- 生效批（N1/N2/N4/条件提交机器/恢复链路）；schema；truth.tsv（#11 等本批）；真 LLM；
  七个 NGAReview 探针中反例 2/3/4+窗口（保持 FAIL，属生效批——不入常驻套件）；
  Invoke 失败后 adopted-无结果的恢复语义（边界见 §四）。

## 三、X3-12..16 冻结期望（翻译依据=评审 199470159 逐条）

| # | 场景 | 预期（冻结） | 依据 |
|---|---|---|---|
| X3-12 | C 首执：调用者仅 read 授权 | **拒**（err≠nil）+ state≠dispatched + 模型零调用 | RV-01：RequiredForDispatch ⊄ EffectiveAuthority ⇒ 不得 dispatched |
| X3-13 | A 重试：同 op、改 digest（κ=材料摘要——brID/evidence 变化不改 κ 则 serve 是正确语义：同材料同结果；**R7 M-2 改真**：κ=digest 唯一，brID/evidence 同 digest 变体=同 κ → serve 正确，非 Conflict——评审 How 的 κ 语义即 digest，vendor 探针三字段同改不单独覆盖变体） | **Conflict 拒**（err≠nil）+ 原结果不被覆盖 + 零模型调用 | RV-02：κ_new≠κ_old ⇒ Conflict；κ=digest 是评审 How 的机器绑定键 |
| X3-14 | B→完成→重试（含方法失效后+reopen） | B 后 stored=true+state=completed；重试走 **A** 返回 B 自己的正式结果；失效不静默重执（仍 A、零模型调用） | RV-03：K 产自己的正式结果+CommonLifecycle |
| X3-15 | 并发同 op 同材料首执（channel 阻塞模型） | 模型**恰被调 1 次**；第二请求在 1s 内收到在途拒 | RV-04：ActiveOwners(o)≤1——claim 先于 Invoke |
| X3-16 | outer opID ≠ req opID | **拒** + 模型零调用 + 内部 op 零采纳（State=""） | RV-05：RejectIdentityMismatch ∧ ΔExecutionCalls=0 ∧ ΔAdoptionState=0 |

## 四、frame-bench 碰撞（file+ctags 两层）

**支撑项**：
- S1（file 层·state-matrix.tsv）：X3-12..16=setup×force×expected 形态，冻结列先于实现——本档 §三即其形态；
- S2（file 层·run-matrix 逐行）：每行一探针先红后绿；对方六探针=外部逐行跑表的 vendor 形态；
- S3（file 层·REVIEW-LEDGER 补行即收敛）：实现期新发现=补行+红探针，不改冻结列；
- S4（ctags 层·符号全枚举）：`grep -c nodeEntryExec(` = x3:16+rk:2（实测）——签名**不变**（digest 经 first.req 传递），调用点改动=语义升级（nil→first）非签名涟漪；`storedResult` 增字段对 JSON 持久化向后兼容（新字段旧档缺省空串）。

**反向检查（singularity）**：
- R1 签名不变是否引入假语义？——nil-first 重试从「静默服务旧结果」变「拒」：是行为变化但正是评审要求的契约（区分 ExecuteRetry 与 ReadHistorical）；既有 4 个 nil-first 重试点全部升级传 first——不是放宽是收紧，无反向；
- R2 Adopt 先于 Invoke：Invoke 失败⇒adopted-无结果⇒在途拒永续？——**边界如实声明**：失败恢复属生效批（OUT §二）；本批 Adopt=claim 的语义按评审 RV-04 How「执行所有权不是业务决策采纳」执行；
- R3 "reused"⇒拒：真重试（completed 后）永不见 reused（A 支路先行）——无假阴性；并发第二名被拒=预期。

**结论：四支撑全过、三反向全闭——开工。**

## 五、验收

1. X3-12..16 五探针红→修→绿（突变验证：删任一修复对应行必红）；
2. **vendor 六探针全绿**（正对照+五反例转绿）；
3. 既有全量 `go test ./... -count=1 -race` 21 包绿；
4. 七 NGAReview 探针状态不变（3P/4F——反例 2/3/4/窗口属生效批，用 run_probes.py 复跑核对）；
5. gofmt/vet 干净。

## 六、守卫×路径×钉覆盖表（R5 遮蔽族五连的结构性回应——族复发 ≥2 换结构条款）

> 族史：R1 守卫零钉→R2 dispatch gate 遮蔽→R3 身份预检遮蔽→R4 nil inputs 遮蔽→
> R5 nil-first B 面遮蔽。根因=守卫位置与探针到达路径强耦合，守卫辖域变化后探针
> 不跟随。结构回应=覆盖表作为**一级制品**（每守卫每可达路径的钉显式列名，遮蔽
> 在表上不可表达——新增守卫/路径时表的行必须同步，空格即缺口）。

| 守卫（t2host_test.go） | 路径 | 钉（测试函数） | 突变杀灭 |
|---|---|---|---|
| 身份预检 outer≠inner | 全路径前置 | X3-16 + vendor MismatchedIdentity | 双红（R5 M3）|
| read 门（E5 面） | 全路径前置 | X3-6（匹配 first 形态——R6 M-1 去遮蔽） | 红（R6 清偿杀灭）|
| completed-无存档拒 | A 面 | X3-6b（既有） | 红（R6 M8 实测）|
| dispatch 门（Effect 秩） | B 面 | X3-12 B 面（合法 inputs） | 红（R4 清偿杀灭）|
| dispatch 门（Effect 秩） | C 面 | X3-12 C 面 + X3-12B + vendor ReadOnly | 三红（R5 M9/M10）|
| 效径命令材料门 nil-first/req | B 面 | X3-12 B-face nil-first 钉 | 红（R5 清偿杀灭）|
| 效径命令材料门 nil-first/req | C 面 | X3-3 nilfirst/nilreq | 红（R5 M2）|
| B 面方法错误拒（execute err） | B 面 | X3-12C（病态 inputs 经入口） | 红（R6 清偿杀灭）|
| nil-Model 门 | C 面 | X3-3 nilmodel（opID 对齐） | 红（R3 清偿杀灭）|
| A 支路 digest Conflict | A 面 | X3-13 + vendor ChangedRetry | 双红（R5 M4）|
| A 支路 nil-first 拒 | A 面 | X3-16A | 红（R5 M5）|
| 在途守卫（State≠""） | 全效径前置 | X3-7/X3-5/X3-11b（chainGrants+匹配 first+诊断词断言——R6 M-2 去遮蔽） | 红（R6 清偿杀灭）|
| claim-before-invoke 顺序（RV-04） | C 面 | X3-15 + vendor Concurrent | 双红（R6 M6 实测）|
| reused 特判 | C 面/commitOutcome | 行为与 fallback 等价（M11 突变绿=等价实证）——诊断分支，无独立钉义务 | 等价在案 |

**新增守卫/新增路径时本表必须同步增行——空格即缺口即下轮审查的 M。**
